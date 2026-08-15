package fastapi

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestFixtureSource_requiredLayoutExists(t *testing.T) {
	root := fixtureRoot(t)
	for _, rel := range []string{
		"app/main.py",
		"app/api/chat.py",
		"app/services/retrieval.py",
		"app/services/generation.py",
		"app/repositories/vector_store.py",
		"app/models/chat.py",
		"tests/test_chat.py",
		"README.md",
		"pyproject.toml",
	} {
		path := filepath.Join(root, filepath.FromSlash(rel))
		if _, err := os.Stat(path); err != nil {
			t.Errorf("missing fixture file %s: %v", rel, err)
		}
	}
}

func TestFixtureSource_parserFindsPythonFilesWithoutExecuting(t *testing.T) {
	found := collectParsedPythonFiles(t, fixtureRoot(t))
	if len(found) < 7 {
		t.Fatalf("parser found %d Python files, want at least 7: %v", len(found), found)
	}
}

func collectParsedPythonFiles(t *testing.T, root string) []string {
	t.Helper()
	var found []string
	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil || d.IsDir() || filepath.Ext(path) != ".py" {
			return walkErr
		}
		parseFixturePython(t, path)
		found = append(found, path)
		return nil
	})
	if err != nil {
		t.Fatalf("walk fixture: %v", err)
	}
	return found
}

func parseFixturePython(t *testing.T, path string) {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	body := string(source)
	if strings.Contains(body, "os.system") || strings.Contains(body, "subprocess") {
		t.Errorf("%s must not spawn processes", path)
	}
	cst := parsePython(source)
	if cst == nil {
		t.Fatalf("parser returned nil for %s", path)
	}
	defer cst.Close()
	if cst.Root() == nil || cst.Root().Kind() != "module" {
		t.Fatalf("parser missed module CST for %s", path)
	}
}

func TestFixtureSource_detectsChatPOSTRoute(t *testing.T) {
	path := filepath.Join(fixtureRoot(t), "app", "api", "chat.py")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	body := string(source)
	if !strings.Contains(body, "@router.post") {
		t.Fatal("fixture handler must use @router.post")
	}
	if !strings.Contains(body, "async def create_chat") {
		t.Fatal("fixture handler must be async create_chat")
	}

	routes := DetectRoutes(File{Path: path, Source: source})
	for _, route := range routes {
		if route.Method == "POST" && route.Path == "/v1/chat" && route.HandlerSymbol == "create_chat" {
			return
		}
	}
	t.Fatalf("expected POST /v1/chat create_chat, got %#v", routes)
}

func TestFixtureSource_chatFlowPresentInStaticSource(t *testing.T) {
	root := fixtureRoot(t)
	assertChatCallChain(t, root)
	assertServiceSymbols(t, root)
	assertEndpointTestAndReadme(t, root)
}

func assertChatCallChain(t *testing.T, root string) {
	t.Helper()
	requireContains(t, "app/api/chat.py", readRel(t, root, "app/api/chat.py"),
		"create_chat", ".retrieve(", ".generate(", "ChatResponse")
}

func assertServiceSymbols(t *testing.T, root string) {
	t.Helper()
	requireContains(t, "app/services/retrieval.py", readRel(t, root, "app/services/retrieval.py"),
		"class RetrievalService", "def retrieve", ".search(")
	requireContains(t, "app/repositories/vector_store.py", readRel(t, root, "app/repositories/vector_store.py"),
		"class VectorStore", "def search")
	requireContains(t, "app/services/generation.py", readRel(t, root, "app/services/generation.py"),
		"class GenerationService", "def generate", "ChatResponse")
	requireContains(t, "app/models/chat.py", readRel(t, root, "app/models/chat.py"), "class ChatResponse")
}

func assertEndpointTestAndReadme(t *testing.T, root string) {
	t.Helper()
	endpointTest := readRel(t, root, "tests/test_chat.py")
	if !strings.Contains(endpointTest, "def test_") || !strings.Contains(endpointTest, "chat") {
		t.Fatal("tests/test_chat.py must define a test named after the chat endpoint")
	}
	requireContains(t, "README.md", readRel(t, root, "README.md"),
		"POST /v1/chat", "create_chat", "RetrievalService.retrieve",
		"VectorStore.search", "GenerationService.generate", "ChatResponse")
}

func TestFixtureSource_emptyContextAlternatePath(t *testing.T) {
	retrieval := readRel(t, fixtureRoot(t), "app/services/retrieval.py")
	generation := readRel(t, fixtureRoot(t), "app/services/generation.py")
	if !strings.Contains(retrieval, "return []") {
		t.Fatal("RetrievalService must return [] when VectorStore.search finds no context")
	}
	if !strings.Contains(generation, "if not ") && !strings.Contains(generation, "if len(") {
		t.Fatal("GenerationService must branch when retrieved context is empty")
	}
}

func fixtureRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", "fixtures", "fastapi-rag-demo"))
	if _, err := os.Stat(root); err != nil {
		t.Fatalf("fixture root %s: %v", root, err)
	}
	return root
}

func readRel(t *testing.T, root, rel string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(rel))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(body)
}

func requireContains(t *testing.T, name, body string, needles ...string) {
	t.Helper()
	for _, needle := range needles {
		if !strings.Contains(body, needle) {
			t.Errorf("%s missing %q", name, needle)
		}
	}
}
