package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storycode/internal/cli"
	"storycode/internal/storage"
)

const wantConfigYAML = `version: 1

repository:
  include:
    - "**/*.py"
    - "tests/**/*.py"
    - "docs/**/*.md"
  exclude:
    - ".git/**"
    - ".venv/**"
    - "venv/**"
    - "__pycache__/**"
    - "node_modules/**"

analysis:
  languages:
    - python
  follow_symlinks: false
  max_file_size_bytes: 5242880

storage:
  mode: repository
  engine: sqlite
`

func TestInit_createsSQLiteDatabaseOnFixture(t *testing.T) {
	root := copyFixture(t)
	chatBefore := readChat(t, root)

	_, stderr, code := runCLI(t, "init", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitOK, stderr)
	}
	dbPath := filepath.Join(root, ".storycode", "index", "storycode.db")
	if _, err := os.Stat(dbPath); err != nil {
		t.Fatalf("missing sqlite database %s: %v", dbPath, err)
	}
	assertMigratedDatabase(t, dbPath)
	if readChat(t, root) != chatBefore {
		t.Fatal("init changed fixture source file app/api/chat.py")
	}
}

func TestInit_createsLayoutOnFixtureCopy(t *testing.T) {
	root := copyFixture(t)
	chatBefore := readChat(t, root)

	stdout, stderr, code := runCLI(t, "init", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitOK, stderr)
	}
	if !strings.Contains(stdout, ".storycode") {
		t.Fatalf("stdout should mention .storycode, got %q", stdout)
	}
	assertStorycodeLayout(t, root)
	got := readFile(t, filepath.Join(root, ".storycode", "config.yaml"))
	if got != wantConfigYAML {
		t.Fatalf("config.yaml mismatch\n got: %q\nwant: %q", got, wantConfigYAML)
	}
	if readChat(t, root) != chatBefore {
		t.Fatal("init changed fixture source file app/api/chat.py")
	}
}

func TestInit_isIdempotentAndKeepsExistingConfig(t *testing.T) {
	root := copyFixture(t)
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("first init exit = %d", code)
	}
	configPath := filepath.Join(root, ".storycode", "config.yaml")
	custom := "version: 99\n"
	if err := os.WriteFile(configPath, []byte(custom), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runCLI(t, "init", root)

	if code != cli.ExitOK {
		t.Fatalf("second init exit = %d; stderr=%q", code, stderr)
	}
	if readFile(t, configPath) != custom {
		t.Fatal("second init overwrote existing config.yaml without --force")
	}
	assertStorycodeLayout(t, root)
	assertMigratedDatabase(t, filepath.Join(root, ".storycode", "index", "storycode.db"))
}

func TestInit_forceOverwritesConfig(t *testing.T) {
	root := copyFixture(t)
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("first init exit = %d", code)
	}
	configPath := filepath.Join(root, ".storycode", "config.yaml")
	if err := os.WriteFile(configPath, []byte("version: 99\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runCLI(t, "init", "--force", root)

	if code != cli.ExitOK {
		t.Fatalf("force init exit = %d; stderr=%q", code, stderr)
	}
	if readFile(t, configPath) != wantConfigYAML {
		t.Fatal("init --force did not restore default config.yaml")
	}
}

func TestInit_pathWithSpaces(t *testing.T) {
	root := copyFixture(t)
	if !strings.Contains(root, " ") {
		t.Fatalf("expected a space in path, got %q", root)
	}

	_, stderr, code := runCLI(t, "init", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	assertStorycodeLayout(t, root)
}

func TestInit_fileTargetIsActionableError(t *testing.T) {
	path := filepath.Join(t.TempDir(), "not a directory")
	if err := os.WriteFile(path, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runCLI(t, "init", path)

	if code != cli.ExitError {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitError, stderr)
	}
	if !strings.Contains(stderr, path) {
		t.Fatalf("stderr should mention %q, got %q", path, stderr)
	}
}

func TestInit_helpMentionsForce(t *testing.T) {
	stdout, stderr, code := runCLI(t, "init", "--help")
	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "--force") {
		t.Fatalf("init --help should mention --force\n%s", stdout)
	}
}

func copyFixture(t *testing.T) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), "rag demo")
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(dst, os.DirFS(fixtureRoot(t))); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return dst
}

func assertStorycodeLayout(t *testing.T, root string) {
	t.Helper()
	base := filepath.Join(root, ".storycode")
	for _, rel := range []string{"config.yaml", "stories", "index", "cache"} {
		path := filepath.Join(base, rel)
		info, err := os.Stat(path)
		if err != nil {
			t.Fatalf("missing %s: %v", path, err)
		}
		if rel != "config.yaml" && !info.IsDir() {
			t.Fatalf("%s is not a directory", path)
		}
	}
}

func assertMigratedDatabase(t *testing.T, path string) {
	t.Helper()
	db, err := storage.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	t.Cleanup(func() { _ = db.Close() })
	var name string
	err = db.QueryRow(`SELECT name FROM sqlite_master WHERE type = 'table' AND name = 'repositories'`).Scan(&name)
	if err != nil {
		t.Fatalf("repositories table missing in %s: %v", path, err)
	}
}

func readChat(t *testing.T, root string) string {
	t.Helper()
	return readFile(t, filepath.Join(root, "app", "api", "chat.py"))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(body)
}
