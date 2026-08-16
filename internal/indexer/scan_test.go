package indexer

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestScan_fixtureFindsPythonPathsWithSlashes(t *testing.T) {
	result, err := Scan(defaultScanOptions(t, fixtureRoot(t)))
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	paths := filePaths(result)
	for _, want := range []string{
		"app/main.py",
		"app/api/chat.py",
		"app/services/retrieval.py",
		"app/services/generation.py",
		"app/repositories/vector_store.py",
		"app/models/chat.py",
		"tests/test_chat.py",
	} {
		if !containsPath(paths, want) {
			t.Errorf("missing %s in %v", want, paths)
		}
		if strings.Contains(want, `\`) {
			t.Fatalf("expected slash-separated path, got %q", want)
		}
	}
	if containsPath(paths, "README.md") || containsPath(paths, "pyproject.toml") {
		t.Fatalf("include should skip non-matching files, got %v", paths)
	}
}

func TestScan_excludesConfiguredAndWellKnownDirs(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "app/ok.py", "x")
	writeFile(t, root, ".venv/lib/hidden.py", "x")
	writeFile(t, root, ".git/hooks/x.py", "x")
	writeFile(t, root, "venv/lib/hidden.py", "x")
	writeFile(t, root, "node_modules/pkg/x.py", "x")
	writeFile(t, root, "__pycache__/ok.pyc", "x")

	result, err := Scan(defaultScanOptions(t, root))
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	paths := filePaths(result)
	if !containsPath(paths, "app/ok.py") {
		t.Fatalf("expected app/ok.py, got %v", paths)
	}
	for _, blocked := range []string{
		".venv/lib/hidden.py",
		".git/hooks/x.py",
		"venv/lib/hidden.py",
		"node_modules/pkg/x.py",
	} {
		if containsPath(paths, blocked) {
			t.Errorf("excluded path present: %s in %v", blocked, paths)
		}
	}
}

func TestScan_unicodeAndSpacesInRelativePaths(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "my app/módulos/café.py", "print(1)\n")
	writeFile(t, root, "docs/guia rápido.md", "# hi\n")

	result, err := Scan(defaultScanOptions(t, root))
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	paths := filePaths(result)
	if !containsPath(paths, "my app/módulos/café.py") {
		t.Fatalf("missing unicode/space python path in %v", paths)
	}
	if !containsPath(paths, "docs/guia rápido.md") {
		t.Fatalf("missing markdown path in %v", paths)
	}
}

func TestScan_skipsSymlinkWhenNotFollowing(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "app/real.py", "x")
	target := filepath.Join(root, "app", "real.py")
	link := filepath.Join(root, "app", "alias.py")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlinks not supported: %v", err)
	}

	result, err := Scan(defaultScanOptions(t, root))
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	paths := filePaths(result)
	if !containsPath(paths, "app/real.py") {
		t.Fatalf("missing real file in %v", paths)
	}
	if containsPath(paths, "app/alias.py") {
		t.Fatalf("followed or indexed symlink %v", paths)
	}
}

func TestScan_oversizeFileIsWarningNotFailure(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "app/small.py", "x")
	writeFile(t, root, "app/huge.py", strings.Repeat("a", 64))

	opts := defaultScanOptions(t, root)
	opts.MaxFileSizeBytes = 16
	result, err := Scan(opts)
	if err != nil {
		t.Fatalf("Scan: %v", err)
	}
	paths := filePaths(result)
	if !containsPath(paths, "app/small.py") {
		t.Fatalf("small file dropped: %v", paths)
	}
	if containsPath(paths, "app/huge.py") {
		t.Fatalf("oversize file indexed: %v", paths)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected oversize warning")
	}
	if !strings.Contains(result.Warnings[0].Message, "app/huge.py") {
		t.Fatalf("warning %q should include offending path", result.Warnings[0].Message)
	}
	if !strings.Contains(result.Warnings[0].Message, "16") {
		t.Fatalf("warning %q should include expected max size 16", result.Warnings[0].Message)
	}
}

func TestScan_unreadablePathIsWarning(t *testing.T) {
	root := t.TempDir()
	writeFile(t, root, "app/ok.py", "x")
	blocked := filepath.Join(root, "secret")
	if err := os.Mkdir(blocked, 0o000); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(blocked, 0o755)
	})
	if _, err := os.ReadDir(blocked); err == nil {
		t.Skip("filesystem does not enforce directory mode 000")
	}

	result, err := Scan(defaultScanOptions(t, root))
	if err != nil {
		t.Fatalf("Scan should not fail globally: %v", err)
	}
	if !containsPath(filePaths(result), "app/ok.py") {
		t.Fatalf("readable file missing: %v", filePaths(result))
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected a warning for unreadable directory")
	}
}

func defaultScanOptions(t *testing.T, root string) ScanOptions {
	t.Helper()
	return ScanOptions{
		Root: root,
		Include: []string{
			"**/*.py",
			"tests/**/*.py",
			"docs/**/*.md",
		},
		Exclude: []string{
			".git/**",
			".venv/**",
			"venv/**",
			"__pycache__/**",
			"node_modules/**",
		},
		FollowSymlinks:   false,
		MaxFileSizeBytes: 5242880,
	}
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

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func filePaths(result ScanResult) []string {
	out := make([]string, 0, len(result.Files))
	for _, f := range result.Files {
		out = append(out, f.Path)
	}
	return out
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}
