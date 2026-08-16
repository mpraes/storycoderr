package python_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"storycode/internal/analyzers/python"
)

func TestExtract_fixtureFindsChatAndServiceSymbols(t *testing.T) {
	files := fixturePythonFiles(t)
	result := python.Extract(files)

	createChat := mustFindDisplay(t, result.Symbols, "create_chat")
	assertSymbol(t, createChat, python.KindFunction, "app/api/chat.py", 13, 16)

	retrieve := mustFindDisplay(t, result.Symbols, "RetrievalService.retrieve")
	assertSymbol(t, retrieve, python.KindMethod, "app/services/retrieval.py", 8, 12)

	generate := mustFindDisplay(t, result.Symbols, "GenerationService.generate")
	assertSymbol(t, generate, python.KindMethod, "app/services/generation.py", 5, 8)

	if createChat.QualifiedName != "app.api.chat.create_chat" {
		t.Fatalf("create_chat qualified_name = %q, want app.api.chat.create_chat", createChat.QualifiedName)
	}
	if retrieve.QualifiedName != "app.services.retrieval.RetrievalService.retrieve" {
		t.Fatalf("retrieve qualified_name = %q, want app.services.retrieval.RetrievalService.retrieve", retrieve.QualifiedName)
	}
	if generate.QualifiedName != "app.services.generation.GenerationService.generate" {
		t.Fatalf("generate qualified_name = %q, want app.services.generation.GenerationService.generate", generate.QualifiedName)
	}
}

func TestExtract_fixtureEmitsRequiredKinds(t *testing.T) {
	result := python.Extract(fixturePythonFiles(t))
	want := []python.Kind{
		python.KindModule,
		python.KindClass,
		python.KindFunction,
		python.KindMethod,
		python.KindDecorator,
		python.KindImport,
	}
	for _, kind := range want {
		if !hasKind(result.Symbols, kind) {
			t.Errorf("fixture extract missing kind %q", kind)
		}
	}
}

func TestExtract_syntaxErrorDoesNotStopOtherFiles(t *testing.T) {
	brokenPath := "broken.py"
	files := []python.File{
		{Path: "ok.py", Source: []byte("def keep():\n    return 1\n")},
		{Path: brokenPath, Source: []byte("def (\n")},
	}

	result := python.Extract(files)

	keep := mustFindDisplay(t, result.Symbols, "keep")
	if keep.SourceFile != "ok.py" {
		t.Fatalf("keep source_file = %q, want ok.py", keep.SourceFile)
	}
	if hasDisplay(result.Symbols, "broken") {
		t.Fatal("did not expect symbols from the syntax-error file")
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected a warning for the syntax-error file")
	}
	warning := result.Warnings[0]
	if warning.Path != brokenPath {
		t.Fatalf("warning path = %q, want %q", warning.Path, brokenPath)
	}
	if !strings.Contains(warning.Message, brokenPath) {
		t.Fatalf("warning %q should include path %q", warning.Message, brokenPath)
	}
}

func TestExtract_doesNotExecutePython(t *testing.T) {
	source := []byte("raise SystemExit('must not execute analyzed code')\n\ndef safe():\n    return 1\n")
	result := python.Extract([]python.File{{Path: "unsafe.py", Source: source}})
	mustFindDisplay(t, result.Symbols, "safe")
}

func TestExtract_semanticHashStableForSameSpan(t *testing.T) {
	files := fixturePythonFiles(t)
	first := mustFindDisplay(t, python.Extract(files).Symbols, "create_chat")
	second := mustFindDisplay(t, python.Extract(files).Symbols, "create_chat")
	if first.SemanticHash == "" {
		t.Fatal("semantic_hash is empty")
	}
	if first.SemanticHash != second.SemanticHash {
		t.Fatalf("semantic_hash changed between runs: %q vs %q", first.SemanticHash, second.SemanticHash)
	}
}

func fixturePythonFiles(t *testing.T) []python.File {
	t.Helper()
	rels := []string{
		"app/main.py",
		"app/api/chat.py",
		"app/services/retrieval.py",
		"app/services/generation.py",
		"app/repositories/vector_store.py",
		"app/models/chat.py",
		"tests/test_chat.py",
	}
	files := make([]python.File, 0, len(rels))
	for _, rel := range rels {
		path := filepath.Join(fixtureRoot(t), filepath.FromSlash(rel))
		source, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("read fixture %s: %v", path, err)
		}
		files = append(files, python.File{Path: rel, Source: source})
	}
	return files
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

func mustFindDisplay(t *testing.T, symbols []python.Symbol, display string) python.Symbol {
	t.Helper()
	for _, symbol := range symbols {
		if symbol.DisplayName == display {
			return symbol
		}
	}
	t.Fatalf("symbol %q not found in %v", display, displayNames(symbols))
	return python.Symbol{}
}

func assertSymbol(t *testing.T, symbol python.Symbol, kind python.Kind, sourceFile string, start, end int) {
	t.Helper()
	if symbol.Kind != kind {
		t.Fatalf("%s kind = %q, want %q", symbol.DisplayName, symbol.Kind, kind)
	}
	if symbol.SourceFile != sourceFile {
		t.Fatalf("%s source_file = %q, want %q", symbol.DisplayName, symbol.SourceFile, sourceFile)
	}
	if symbol.StartLine != start || symbol.EndLine != end {
		t.Fatalf("%s lines = %d-%d, want %d-%d", symbol.DisplayName, symbol.StartLine, symbol.EndLine, start, end)
	}
	if symbol.EndLine < symbol.StartLine {
		t.Fatalf("%s end_line %d is before start_line %d", symbol.DisplayName, symbol.EndLine, symbol.StartLine)
	}
}

func hasKind(symbols []python.Symbol, kind python.Kind) bool {
	for _, symbol := range symbols {
		if symbol.Kind == kind {
			return true
		}
	}
	return false
}

func hasDisplay(symbols []python.Symbol, display string) bool {
	for _, symbol := range symbols {
		if symbol.DisplayName == display {
			return true
		}
	}
	return false
}

func displayNames(symbols []python.Symbol) []string {
	names := make([]string, 0, len(symbols))
	for _, symbol := range symbols {
		names = append(names, symbol.DisplayName)
	}
	return names
}
