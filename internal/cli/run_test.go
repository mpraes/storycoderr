package cli_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"storycode/internal/cli"
)

func TestRun_helpListsAllCommands(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--help")

	if code != cli.ExitOK {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitOK, stderr)
	}
	for _, name := range []string{"init", "status", "index", "discover", "serve", "story", "verify"} {
		if !strings.Contains(stdout, name) {
			t.Fatalf("help missing command %q\n%s", name, stdout)
		}
	}
}

func TestRun_eachCommandHasHelp(t *testing.T) {
	commands := [][]string{
		{"init", "--help"},
		{"status", "--help"},
		{"index", "--help"},
		{"discover", "--help"},
		{"serve", "--help"},
		{"story", "--help"},
		{"story", "list", "--help"},
		{"story", "show", "--help"},
		{"verify", "--help"},
	}
	for _, args := range commands {
		stdout, stderr, code := runCLI(t, args...)
		if code != cli.ExitOK {
			t.Fatalf("%v exit = %d, want %d; stderr=%q", args, code, cli.ExitOK, stderr)
		}
		if !strings.Contains(stdout, "Usage:") {
			t.Fatalf("%v help missing Usage:\n%s", args, stdout)
		}
	}
}

func TestRun_versionIncludesVersionAndCommit(t *testing.T) {
	stdout, stderr, code := runCLI(t, "--version")

	if code != cli.ExitOK {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitOK, stderr)
	}
	if !strings.Contains(stdout, "0.1.0") {
		t.Fatalf("version missing 0.1.0: %q", stdout)
	}
	if !strings.Contains(stdout, "abc1234") {
		t.Fatalf("version missing commit abc1234: %q", stdout)
	}
}

func TestRun_emptyCommandsExitZero(t *testing.T) {
	commands := [][]string{
		{"init", t.TempDir()},
		{"status"},
		{"index"},
		{"discover"},
		{"serve"},
		{"story", "list"},
		{"story", "show", "chat-post"},
		{"verify"},
	}
	for _, args := range commands {
		_, stderr, code := runCLI(t, args...)
		if code != cli.ExitOK {
			t.Fatalf("%v exit = %d, want %d; stderr=%q", args, code, cli.ExitOK, stderr)
		}
	}
}

func TestRun_storyShowMissingKeyIsUsageError(t *testing.T) {
	_, stderr, code := runCLI(t, "story", "show")

	if code != cli.ExitUsage {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitUsage, stderr)
	}
	if !strings.Contains(stderr, "story show") || !strings.Contains(stderr, "<key>") {
		t.Fatalf("stderr should mention story show <key>, got %q", stderr)
	}
}

func TestRun_unknownCommandIsUsageError(t *testing.T) {
	_, stderr, code := runCLI(t, "explode")

	if code != cli.ExitUsage {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitUsage, stderr)
	}
	if !strings.Contains(stderr, "explode") {
		t.Fatalf("stderr should mention explode, got %q", stderr)
	}
}

func TestRun_emptyCommandsDoNotChangeFixture(t *testing.T) {
	root := fixtureRoot(t)
	before := snapshotDir(t, root)

	commands := [][]string{
		{"status"},
		{"index"},
		{"discover"},
		{"serve"},
		{"story", "list"},
		{"story", "show", "chat-post"},
		{"verify"},
	}
	for _, args := range commands {
		_, stderr, code := runCLI(t, args...)
		if code != cli.ExitOK {
			t.Fatalf("%v exit = %d; stderr=%q", args, code, stderr)
		}
	}

	after := snapshotDir(t, root)
	if len(before) == 0 {
		t.Fatalf("fixture %s has no files", root)
	}
	for path, sum := range before {
		if after[path] != sum {
			t.Fatalf("fixture file %s changed", path)
		}
	}
	if len(after) != len(before) {
		t.Fatalf("fixture file count %d -> %d", len(before), len(after))
	}
}

func runCLI(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	code := cli.Run(args, &stdout, &stderr, cli.BuildInfo{Version: "0.1.0", Commit: "abc1234"})
	return stdout.String(), stderr.String(), code
}

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "fixtures", "fastapi-rag-demo"))
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("fixture %s: %v", root, err)
	}
	return root
}

func snapshotDir(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		sum, err := fileSHA256(path)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = sum
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", root, err)
	}
	return out
}

func fileSHA256(path string) (string, error) {
	body, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:]), nil
}
