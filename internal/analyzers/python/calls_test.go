package python_test

import (
	"strings"
	"testing"

	"storycode/internal/analyzers/python"
)

func TestExtractCalls_fixtureResolvesChatFlow(t *testing.T) {
	result := python.ExtractCalls(fixturePythonFiles(t))

	assertCall(t, result.Relations, callWant{
		from:       "app.api.chat.create_chat",
		to:         "app.services.retrieval.RetrievalService.retrieve",
		line:       15,
		confidence: python.ConfidenceHigh,
	})
	assertCall(t, result.Relations, callWant{
		from:       "app.api.chat.create_chat",
		to:         "app.services.generation.GenerationService.generate",
		line:       16,
		confidence: python.ConfidenceHigh,
	})
	assertCall(t, result.Relations, callWant{
		from:       "app.services.retrieval.RetrievalService.retrieve",
		to:         "app.repositories.vector_store.VectorStore.search",
		line:       9,
		confidence: python.ConfidenceHigh,
	})
}

func TestExtractCalls_unresolvedCalleeUsesExternalRef(t *testing.T) {
	source := []byte("def run():\n    unknown.retrieve(query)\n")
	result := python.ExtractCalls([]python.File{{Path: "app/run.py", Source: source}})

	rel := mustCallFrom(t, result.Relations, "app.run.run")
	if rel.ToSymbol != "" {
		t.Fatalf("to_symbol = %q, want empty for unresolved call", rel.ToSymbol)
	}
	if rel.ExternalRef != "unknown.retrieve" {
		t.Fatalf("external_ref = %q, want unknown.retrieve", rel.ExternalRef)
	}
	if rel.Kind != python.KindCalls {
		t.Fatalf("kind = %q, want %q", rel.Kind, python.KindCalls)
	}
	if rel.Confidence == python.ConfidenceHigh {
		t.Fatalf("confidence = %q, must not claim high for unresolved call", rel.Confidence)
	}
}

func TestExtractCalls_ambiguousMethodStaysExternal(t *testing.T) {
	source := []byte("" +
		"class StoreA:\n" +
		"    def search(self):\n" +
		"        return []\n" +
		"class StoreB:\n" +
		"    def search(self):\n" +
		"        return []\n" +
		"def run(store):\n" +
		"    store.search()\n")
	result := python.ExtractCalls([]python.File{{Path: "app/ambig.py", Source: source}})

	rel := mustCallFrom(t, result.Relations, "app.ambig.run")
	if rel.ToSymbol != "" {
		t.Fatalf("to_symbol = %q, must not pick a target when search is ambiguous", rel.ToSymbol)
	}
	if rel.ExternalRef != "store.search" {
		t.Fatalf("external_ref = %q, want store.search", rel.ExternalRef)
	}
	if rel.Confidence == python.ConfidenceHigh {
		t.Fatalf("confidence = %q, must not claim high when the callee is ambiguous", rel.Confidence)
	}
}

func TestExtractCalls_syntaxErrorDoesNotStopOtherFiles(t *testing.T) {
	files := []python.File{
		{Path: "ok.py", Source: []byte("def keep():\n    helper()\n")},
		{Path: "broken.py", Source: []byte("def (\n")},
	}
	result := python.ExtractCalls(files)

	rel := mustCallFrom(t, result.Relations, "ok.keep")
	if rel.ExternalRef != "helper" {
		t.Fatalf("external_ref = %q, want helper", rel.ExternalRef)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected a warning for the syntax-error file")
	}
	if result.Warnings[0].Path != "broken.py" {
		t.Fatalf("warning path = %q, want broken.py", result.Warnings[0].Path)
	}
	if !strings.Contains(result.Warnings[0].Message, "broken.py") {
		t.Fatalf("warning %q should include path broken.py", result.Warnings[0].Message)
	}
}

func TestExtractCalls_doesNotExecutePython(t *testing.T) {
	source := []byte("raise SystemExit('must not execute analyzed code')\n\ndef run():\n    helper()\n")
	result := python.ExtractCalls([]python.File{{Path: "unsafe.py", Source: source}})
	mustCallFrom(t, result.Relations, "unsafe.run")
}

type callWant struct {
	from       string
	to         string
	line       int
	confidence string
}

func assertCall(t *testing.T, rels []python.CallRelation, want callWant) {
	t.Helper()
	for _, rel := range rels {
		if rel.FromSymbol != want.from || rel.ToSymbol != want.to {
			continue
		}
		if rel.Kind != python.KindCalls {
			t.Fatalf("%s -> %s kind = %q, want %q", want.from, want.to, rel.Kind, python.KindCalls)
		}
		if rel.ExternalRef != "" {
			t.Fatalf("%s -> %s external_ref = %q, want empty when resolved", want.from, want.to, rel.ExternalRef)
		}
		if rel.Line != want.line {
			t.Fatalf("%s -> %s line = %d, want %d", want.from, want.to, rel.Line, want.line)
		}
		if rel.Confidence != want.confidence {
			t.Fatalf("%s -> %s confidence = %q, want %q", want.from, want.to, rel.Confidence, want.confidence)
		}
		return
	}
	t.Fatalf("missing call %s -> %s in %v", want.from, want.to, formatCalls(rels))
}

func mustCallFrom(t *testing.T, rels []python.CallRelation, from string) python.CallRelation {
	t.Helper()
	for _, rel := range rels {
		if rel.FromSymbol == from {
			return rel
		}
	}
	t.Fatalf("no call from %s in %v", from, formatCalls(rels))
	return python.CallRelation{}
}

func formatCalls(rels []python.CallRelation) []string {
	out := make([]string, 0, len(rels))
	for _, rel := range rels {
		target := rel.ToSymbol
		if target == "" {
			target = rel.ExternalRef
		}
		out = append(out, rel.FromSymbol+" -> "+target)
	}
	return out
}
