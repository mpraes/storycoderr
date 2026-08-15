package cli

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

type denyWriteFilesystem struct {
	osFilesystem
}

func (denyWriteFilesystem) MkdirAll(path string, _ os.FileMode) error {
	return fmt.Errorf("permission denied")
}

func (denyWriteFilesystem) WriteFile(path string, _ []byte, _ os.FileMode) error {
	return fmt.Errorf("permission denied")
}

type fixedRootFilesystem struct {
	osFilesystem
	wd     string
	wdErr  error
	abs    map[string]string
	absErr error
}

func (f fixedRootFilesystem) Getwd() (string, error) {
	if f.wdErr != nil {
		return "", f.wdErr
	}
	return f.wd, nil
}

func (f fixedRootFilesystem) Abs(path string) (string, error) {
	if f.absErr != nil {
		return "", f.absErr
	}
	if got, ok := f.abs[path]; ok {
		return got, nil
	}
	return "", fmt.Errorf("unexpected Abs(%q)", path)
}

type presentStatFilesystem struct {
	osFilesystem
}

func (presentStatFilesystem) Stat(path string) (os.FileInfo, error) {
	return namedFileInfo{name: filepath.Base(path)}, nil
}

type missingStatFilesystem struct {
	osFilesystem
}

func (missingStatFilesystem) Stat(path string) (os.FileInfo, error) {
	return nil, os.ErrNotExist
}

type namedFileInfo struct {
	name string
}

func (f namedFileInfo) Name() string       { return f.name }
func (f namedFileInfo) Size() int64        { return 0 }
func (f namedFileInfo) Mode() os.FileMode  { return 0 }
func (f namedFileInfo) ModTime() time.Time { return time.Time{} }
func (f namedFileInfo) IsDir() bool        { return false }
func (f namedFileInfo) Sys() any           { return nil }

func TestInitRoot_usesArgument(t *testing.T) {
	files := fixedRootFilesystem{abs: map[string]string{`.\repo`: `/abs/repo`}}
	got, err := initRoot([]string{`.\repo`}, files)
	if err != nil {
		t.Fatal(err)
	}
	if got != `/abs/repo` {
		t.Fatalf("initRoot = %q, want /abs/repo", got)
	}
}

func TestInitRoot_usesWorkingDirectory(t *testing.T) {
	got, err := initRoot(nil, fixedRootFilesystem{wd: `/cwd`})
	if err != nil {
		t.Fatal(err)
	}
	if got != `/cwd` {
		t.Fatalf("initRoot = %q, want /cwd", got)
	}
}

func TestAbsInitRoot_includesOffendingPath(t *testing.T) {
	_, err := absInitRoot("bad path", fixedRootFilesystem{absErr: errors.New("boom")})
	if err == nil || !strings.Contains(err.Error(), "bad path") {
		t.Fatalf("error should mention bad path, got %v", err)
	}
}

func TestJoinStorycodePath(t *testing.T) {
	if got := joinStorycodePath("/repo/.storycode", ""); got != "/repo/.storycode" {
		t.Fatalf("empty name = %q", got)
	}
	want := filepath.Join("/repo/.storycode", "cache")
	if got := joinStorycodePath("/repo/.storycode", "cache"); got != want {
		t.Fatalf("cache = %q, want %q", got, want)
	}
}

func TestCreateStorycodeDirs_permissionErrorIncludesPath(t *testing.T) {
	err := createStorycodeDirs("/repo/.storycode", denyWriteFilesystem{})
	if err == nil || !strings.Contains(err.Error(), "/repo/.storycode") {
		t.Fatalf("error should mention path, got %v", err)
	}
	if !strings.Contains(err.Error(), "writable directory") {
		t.Fatalf("error should mention expected shape, got %v", err)
	}
}

func TestWriteConfig_skipsExistingWithoutForce(t *testing.T) {
	wrote, err := writeConfig("/repo/.storycode/config.yaml", false, presentStatFilesystem{})
	if err != nil {
		t.Fatal(err)
	}
	if wrote {
		t.Fatal("wrote existing config without --force")
	}
}

func TestWriteConfig_writesWhenMissing(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	wrote, err := writeConfig(path, false, missingStatFilesystem{})
	if err != nil {
		t.Fatal(err)
	}
	if !wrote {
		t.Fatal("expected write when config is missing")
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != defaultConfigYAML {
		t.Fatalf("wrote %q", body)
	}
}

func TestWriteConfig_permissionErrorIncludesPath(t *testing.T) {
	_, err := writeConfig("/repo/.storycode/config.yaml", true, denyWriteFilesystem{})
	if err == nil || !strings.Contains(err.Error(), "/repo/.storycode/config.yaml") {
		t.Fatalf("error should mention path, got %v", err)
	}
}

func TestConfigExists(t *testing.T) {
	ok, err := configExists("/repo/.storycode/config.yaml", presentStatFilesystem{})
	if err != nil || !ok {
		t.Fatalf("present: ok=%v err=%v", ok, err)
	}
	ok, err = configExists("/repo/.storycode/config.yaml", missingStatFilesystem{})
	if err != nil || ok {
		t.Fatalf("missing: ok=%v err=%v", ok, err)
	}
}

func TestReportInit(t *testing.T) {
	var created, existing bytes.Buffer
	reportInit(&created, "/repo/.storycode", true)
	reportInit(&existing, "/repo/.storycode", false)
	if !strings.Contains(created.String(), "Initialized") {
		t.Fatalf("created = %q", created.String())
	}
	if !strings.Contains(existing.String(), "already initialized") {
		t.Fatalf("existing = %q", existing.String())
	}
}

func TestExitCode(t *testing.T) {
	if got := exitCode(usageError{msg: "story show requires a key"}); got != ExitUsage {
		t.Fatalf("usageError = %d, want %d", got, ExitUsage)
	}
	if got := exitCode(errors.New(`unknown command "explode"`)); got != ExitUsage {
		t.Fatalf("unknown command = %d, want %d", got, ExitUsage)
	}
	if got := exitCode(errors.New("cannot create directory")); got != ExitError {
		t.Fatalf("io error = %d, want %d", got, ExitError)
	}
}

func TestStoryShowArgs(t *testing.T) {
	if err := storyShowArgs(nil, []string{"chat-post"}); err != nil {
		t.Fatal(err)
	}
	err := storyShowArgs(nil, nil)
	if err == nil || !strings.Contains(err.Error(), "<key>") {
		t.Fatalf("missing key error = %v", err)
	}
}
