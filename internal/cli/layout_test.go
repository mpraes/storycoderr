package cli_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"storycode/internal/analyzers/fastapi"
)

func TestRepoLayout_requiredPathsExist(t *testing.T) {
	root := moduleRoot(t)
	for _, rel := range []string{
		"cmd/storycode",
		"internal/cli",
		"internal/config",
		"internal/domain",
		"internal/storage",
		"internal/repository",
		"internal/indexer",
		"internal/analyzers",
		"internal/stories",
		"internal/verification",
		"internal/server",
		"internal/assets",
		"migrations",
		"web",
		"fixtures/fastapi-rag-demo",
		"docs",
		"scripts",
		".github/workflows",
		"CHANGELOG.md",
		"README.md",
		"LICENSE",
		"go.mod",
		"Makefile",
		".editorconfig",
		".gitattributes",
		".gitignore",
		"docs/development/coding-standards.md",
		"docs/development/definition-of-done.md",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing %s: %v", rel, err)
		}
	}
}

func TestRepoLayout_fixtureIsReadableWithoutExecutingPython(t *testing.T) {
	root := moduleRoot(t)
	path := filepath.Join(root, "fixtures", "fastapi-rag-demo", "app", "api", "chat.py")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}

	routes := fastapi.DetectRoutes(fastapi.File{Path: path, Source: source})
	found := false
	for _, route := range routes {
		if route.Method == "POST" && route.Path == "/v1/chat" && route.HandlerSymbol == "create_chat" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected POST /v1/chat create_chat in fixture, got %#v", routes)
	}
}

func moduleRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("go.mod not found at %s: %v", root, err)
	}
	return root
}
