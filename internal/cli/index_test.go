package cli_test

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storycode/internal/cli"
	"storycode/internal/storage"
)

func TestIndex_indexesFastAPIFixture(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	chatBefore := readChat(t, root)
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}

	stdout, stderr, code := runCLI(t, "index", root)

	if code != cli.ExitOK {
		t.Fatalf("index exit = %d; stderr=%q stdout=%q", code, stderr, stdout)
	}
	assertIndexPhases(t, stdout)
	db := openIndexDB(t, root)
	defer db.Close()
	assertCompletedIndexRun(t, db, "full")
	if countTable(t, db, "source_files") != 7 {
		t.Fatalf("source_files = %d, want 7 python files from the fixture", countTable(t, db, "source_files"))
	}
	if !hasSymbol(t, db, "app.api.chat.create_chat") {
		t.Fatal("missing symbol app.api.chat.create_chat")
	}
	if !hasEntryPoint(t, db, "http:POST:/v1/chat") {
		t.Fatal("missing entry point http:POST:/v1/chat")
	}
	if !hasCall(t, db, "app.api.chat.create_chat", "app.services.retrieval.RetrievalService.retrieve") {
		t.Fatal("missing call create_chat -> retrieve")
	}
	if readChat(t, root) != chatBefore {
		t.Fatal("index changed fixture source file app/api/chat.py")
	}
}

func TestIndex_repeatDoesNotDuplicateRows(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}
	if _, stderr, code := runCLI(t, "index", root); code != cli.ExitOK {
		t.Fatalf("first index exit = %d; stderr=%q", code, stderr)
	}

	_, stderr, code := runCLI(t, "index", root)
	if code != cli.ExitOK {
		t.Fatalf("second index exit = %d; stderr=%q", code, stderr)
	}

	db := openIndexDB(t, root)
	defer db.Close()
	if countTable(t, db, "source_files") != 7 {
		t.Fatalf("source_files duplicated: %d", countTable(t, db, "source_files"))
	}
	if countNamed(t, db, "code_symbols", "qualified_name", "app.api.chat.create_chat") != 1 {
		t.Fatal("create_chat symbol duplicated")
	}
	if countNamed(t, db, "entry_points", "key", "http:POST:/v1/chat") != 1 {
		t.Fatal("entry point duplicated")
	}
	if countTable(t, db, "index_runs") != 2 {
		t.Fatalf("index_runs = %d, want 2 completed runs", countTable(t, db, "index_runs"))
	}
}

func TestIndex_unchangedSecondRunIsIncremental(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}
	if _, stderr, code := runCLI(t, "index", root); code != cli.ExitOK {
		t.Fatalf("first index exit = %d; stderr=%q", code, stderr)
	}

	stdout, stderr, code := runCLI(t, "index", root)
	if code != cli.ExitOK {
		t.Fatalf("second index exit = %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(strings.ToLower(stdout), "incremental") {
		t.Fatalf("second run should report incremental, got %q", stdout)
	}
	db := openIndexDB(t, root)
	defer db.Close()
	assertCompletedIndexRun(t, db, "incremental")
}

func TestIndex_createsDatabaseWithoutInit(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")

	_, stderr, code := runCLI(t, "index", root)
	if code != cli.ExitOK {
		t.Fatalf("index exit = %d; stderr=%q", code, stderr)
	}
	if _, err := os.Stat(storage.DatabasePath(root)); err != nil {
		t.Fatalf("first index should create sqlite: %v", err)
	}
}

func TestIndex_statusShowsIndexCounts(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}
	if _, stderr, code := runCLI(t, "index", root); code != cli.ExitOK {
		t.Fatalf("index exit = %d; stderr=%q", code, stderr)
	}

	stdout, stderr, code := runCLI(t, "status", root)
	if code != cli.ExitOK {
		t.Fatalf("status exit = %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Index status: indexed") {
		t.Fatalf("want indexed, got %q", stdout)
	}
	for _, label := range []string{"Files:", "Symbols:", "Relations:", "Entry points:"} {
		if !strings.Contains(stdout, label) {
			t.Fatalf("status missing %s in %q", label, stdout)
		}
	}
	if !strings.Contains(stdout, "Files: 7") {
		t.Fatalf("want Files: 7, got %q", stdout)
	}
	if !strings.Contains(stdout, "Entry points: 1") {
		t.Fatalf("want Entry points: 1, got %q", stdout)
	}
}

func TestIndex_helpListsDirArg(t *testing.T) {
	stdout, stderr, code := runCLI(t, "index", "--help")
	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Index") && !strings.Contains(stdout, "index") {
		t.Fatalf("help missing index description:\n%s", stdout)
	}
}

func assertIndexPhases(t *testing.T, stdout string) {
	t.Helper()
	phases := []string{
		"Scanning files",
		"Persisting source files",
		"Extracting symbols",
		"Detecting FastAPI routes",
		"Extracting calls",
		"Completing index",
	}
	for _, phase := range phases {
		if !strings.Contains(stdout, phase) {
			t.Fatalf("progress missing phase %q in %q", phase, stdout)
		}
	}
}

func openIndexDB(t *testing.T, root string) *sql.DB {
	t.Helper()
	db, err := storage.Open(storage.DatabasePath(root))
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func assertCompletedIndexRun(t *testing.T, db *sql.DB, kind string) {
	t.Helper()
	var n int
	err := db.QueryRow(`
SELECT COUNT(*) FROM index_runs
WHERE kind = ? AND status IN ('completed', 'completed_with_warnings')
`, kind).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	if n < 1 {
		t.Fatalf("missing completed index_run kind=%s", kind)
	}
}

func countTable(t *testing.T, db *sql.DB, table string) int {
	t.Helper()
	var n int
	err := db.QueryRow(`SELECT COUNT(*) FROM ` + table + ` WHERE deleted_at IS NULL`).Scan(&n)
	if err != nil {
		err = db.QueryRow(`SELECT COUNT(*) FROM ` + table).Scan(&n)
	}
	if err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return n
}

func countNamed(t *testing.T, db *sql.DB, table, column, value string) int {
	t.Helper()
	var n int
	q := `SELECT COUNT(*) FROM ` + table + ` WHERE ` + column + ` = ?`
	if err := db.QueryRow(q, value).Scan(&n); err != nil {
		t.Fatal(err)
	}
	return n
}

func hasSymbol(t *testing.T, db *sql.DB, qualified string) bool {
	t.Helper()
	return countNamed(t, db, "code_symbols", "qualified_name", qualified) > 0
}

func hasEntryPoint(t *testing.T, db *sql.DB, key string) bool {
	t.Helper()
	return countNamed(t, db, "entry_points", "key", key) > 0
}

func hasCall(t *testing.T, db *sql.DB, fromName, toName string) bool {
	t.Helper()
	var n int
	err := db.QueryRow(`
SELECT COUNT(*) FROM code_relations r
JOIN code_symbols f ON f.id = r.from_symbol_id
JOIN code_symbols tgt ON tgt.id = r.to_symbol_id
WHERE f.qualified_name = ? AND tgt.qualified_name = ? AND r.kind = 'calls'
`, fromName, toName).Scan(&n)
	if err != nil {
		t.Fatal(err)
	}
	return n > 0
}

func TestIndex_doesNotExecuteFixturePython(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	marker := filepath.Join(root, "executed.marker")
	if err := os.WriteFile(filepath.Join(root, "app", "api", "chat.py"), []byte(""+
		"open('executed.marker','w').write('ran')\n"+
		"from fastapi import APIRouter\n"+
		"router = APIRouter(prefix='/v1')\n"+
		"@router.post('/chat')\n"+
		"async def create_chat():\n"+
		"    return {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	_, stderr, code := runCLI(t, "index", root)
	if code != cli.ExitOK {
		t.Fatalf("index exit = %d; stderr=%q", code, stderr)
	}
	if _, err := os.Stat(marker); err == nil {
		t.Fatal("index executed fixture Python; executed.marker exists")
	}
}
