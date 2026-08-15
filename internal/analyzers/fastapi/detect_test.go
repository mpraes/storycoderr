package fastapi_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"storycode/internal/analyzers/fastapi"
)

func TestDetectRoutes_fixtureFindsChatPOST(t *testing.T) {
	path, source := readFixture(t, "app/api/chat.py")

	routes := fastapi.DetectRoutes(fastapi.File{Path: path, Source: source})

	route := findRoute(t, routes, "POST", "/v1/chat")
	if route.HandlerSymbol != "create_chat" {
		t.Fatalf("handler = %q, want create_chat", route.HandlerSymbol)
	}
	if route.Framework != fastapi.FrameworkName {
		t.Fatalf("framework = %q, want %q", route.Framework, fastapi.FrameworkName)
	}
	if route.EntryPointKey != "http:POST:/v1/chat" {
		t.Fatalf("entry point key = %q, want http:POST:/v1/chat", route.EntryPointKey)
	}
	if route.Confidence != fastapi.ConfidenceHigh {
		t.Fatalf("confidence = %q, want %q", route.Confidence, fastapi.ConfidenceHigh)
	}
}

func TestDetectRoutes_unresolvedPrefixWarnsAndLowersConfidence(t *testing.T) {
	source := []byte(`
from app.api.chat import router

@router.post("/chat")
async def create_chat():
    return {}
`)

	routes := fastapi.DetectRoutes(fastapi.File{Path: "imported.py", Source: source})

	route := findRoute(t, routes, "POST", "/chat")
	if route.Confidence != fastapi.ConfidenceMedium {
		t.Fatalf("confidence = %q, want %q", route.Confidence, fastapi.ConfidenceMedium)
	}
	if len(route.Warnings) == 0 {
		t.Fatal("expected a warning when router prefix cannot be resolved")
	}
	if !strings.Contains(route.Warnings[0], "router") {
		t.Fatalf("warning %q should mention identifier router", route.Warnings[0])
	}
}

func TestDetectRoutes_includeRouterPrefixInSameFile(t *testing.T) {
	source := []byte(`
from fastapi import APIRouter, FastAPI

router = APIRouter()
app = FastAPI()
app.include_router(router, prefix="/v1")

@router.post("/chat")
async def create_chat():
    return {}
`)

	routes := fastapi.DetectRoutes(fastapi.File{Path: "included.py", Source: source})

	route := findRoute(t, routes, "POST", "/v1/chat")
	if route.HandlerSymbol != "create_chat" {
		t.Fatalf("handler = %q, want create_chat", route.HandlerSymbol)
	}
	if route.Confidence != fastapi.ConfidenceHigh {
		t.Fatalf("confidence = %q, want %q", route.Confidence, fastapi.ConfidenceHigh)
	}
}

func TestDetectRoutes_doesNotExecutePython(t *testing.T) {
	source := []byte(`
raise SystemExit("must not execute analyzed code")

from fastapi import APIRouter

router = APIRouter(prefix="/v1")

@router.post("/chat")
async def create_chat():
    return {}
`)

	routes := fastapi.DetectRoutes(fastapi.File{Path: "unsafe.py", Source: source})

	findRoute(t, routes, "POST", "/v1/chat")
}

func TestDetectRoutes_httpMethodsAndPathKeyword(t *testing.T) {
	source := []byte(`
from fastapi import APIRouter

router = APIRouter()

@router.get("/items")
async def list_items():
    return []

@router.put("/items")
async def replace_item():
    return {}

@router.patch("/items")
async def update_item():
    return {}

@router.delete("/items")
async def delete_item():
    return {}

@router.post(path="/items")
async def create_item():
    return {}
`)

	routes := fastapi.DetectRoutes(fastapi.File{Path: "methods.py", Source: source})

	cases := []struct {
		method  string
		handler string
	}{
		{"GET", "list_items"},
		{"PUT", "replace_item"},
		{"PATCH", "update_item"},
		{"DELETE", "delete_item"},
		{"POST", "create_item"},
	}
	for _, tc := range cases {
		route := findRoute(t, routes, tc.method, "/items")
		if route.HandlerSymbol != tc.handler {
			t.Fatalf("%s handler = %q, want %q", tc.method, route.HandlerSymbol, tc.handler)
		}
	}
}

func TestDetectRoutes_nonLiteralPrefixLowersConfidence(t *testing.T) {
	source := []byte(`
from fastapi import APIRouter

PREFIX = "/v1"
router = APIRouter(prefix=PREFIX)

@router.post("/chat")
async def create_chat():
    return {}
`)

	routes := fastapi.DetectRoutes(fastapi.File{Path: "dynamic.py", Source: source})

	route := findRoute(t, routes, "POST", "/chat")
	if route.Confidence != fastapi.ConfidenceMedium {
		t.Fatalf("confidence = %q, want %q", route.Confidence, fastapi.ConfidenceMedium)
	}
	if len(route.Warnings) == 0 {
		t.Fatal("expected a warning for non-literal APIRouter prefix")
	}
}

func readFixture(t *testing.T, rel string) (string, []byte) {
	t.Helper()
	path := filepath.Join(repoRoot(t), "fixtures", "fastapi-rag-demo", filepath.FromSlash(rel))
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read fixture %s: %v", path, err)
	}
	return path, source
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
	if _, err := os.Stat(filepath.Join(root, "go.mod")); err != nil {
		t.Fatalf("go.mod not found at %s: %v", root, err)
	}
	return root
}

func findRoute(t *testing.T, routes []fastapi.Route, method, path string) fastapi.Route {
	t.Helper()
	for _, route := range routes {
		if route.Method == method && route.Path == path {
			return route
		}
	}
	t.Fatalf("route %s %s not found in %#v", method, path, routes)
	return fastapi.Route{}
}
