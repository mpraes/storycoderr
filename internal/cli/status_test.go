package cli_test

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"storycode/internal/cli"
	"storycode/internal/storage"
)

func TestStatus_notIndexedAfterInitOnFixture(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}

	stdout, stderr, code := runCLI(t, "status", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitOK, stderr)
	}
	want := strings.Join([]string{
		"Repository: fastapi-rag-demo",
		"Index status: not indexed",
		"Stories: 0",
		"Database: .storycode/index/storycode.db",
		"Config: .storycode/config.yaml",
		"",
	}, "\n")
	if stdout != want {
		t.Fatalf("stdout mismatch\n got: %q\nwant: %q", stdout, want)
	}
}

func TestStatus_jsonAfterInitOnFixture(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}

	stdout, stderr, code := runCLI(t, "status", "--json", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d, want %d; stderr=%q", code, cli.ExitOK, stderr)
	}
	got := parseStatusJSON(t, stdout)
	if got["repository"] != "fastapi-rag-demo" {
		t.Fatalf("repository = %v", got["repository"])
	}
	if got["index_status"] != "not_indexed" {
		t.Fatalf("index_status = %v, want not_indexed", got["index_status"])
	}
	if got["stories"] != float64(0) {
		t.Fatalf("stories = %v", got["stories"])
	}
	if got["database"] != ".storycode/index/storycode.db" {
		t.Fatalf("database = %v", got["database"])
	}
	if got["config"] != ".storycode/config.yaml" {
		t.Fatalf("config = %v", got["config"])
	}
}

func TestStatus_indexedAfterCompletedIndexRun(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}
	seedIndexRun(t, root, "completed", 1)

	stdout, stderr, code := runCLI(t, "status", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Index status: indexed") {
		t.Fatalf("want indexed, got %q", stdout)
	}
	if !strings.Contains(stdout, "Stories: 1") {
		t.Fatalf("want Stories: 1, got %q", stdout)
	}
}

func TestStatus_privacyHidesAbsolutePaths(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}

	stdout, stderr, code := runCLI(t, "status", "--privacy", root)
	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	if strings.Contains(stdout, root) {
		t.Fatalf("privacy leaked absolute path %q in %q", root, stdout)
	}
	jsonOut, stderr, code := runCLI(t, "status", "--json", "--privacy", root)
	if code != cli.ExitOK {
		t.Fatalf("json privacy exit = %d; stderr=%q", code, stderr)
	}
	if strings.Contains(jsonOut, root) {
		t.Fatalf("json privacy leaked absolute path %q in %q", root, jsonOut)
	}
}

func TestStatus_succeedsWithoutGitOnPath(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	if _, _, code := runCLI(t, "init", root); code != cli.ExitOK {
		t.Fatalf("init exit = %d", code)
	}
	t.Setenv("PATH", t.TempDir())

	stdout, stderr, code := runCLI(t, "status", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Index status: not indexed") {
		t.Fatalf("got %q", stdout)
	}
}

func TestStatus_beforeInitIsNotIndexed(t *testing.T) {
	root := copyFixtureAs(t, "fastapi-rag-demo")
	chatBefore := readChat(t, root)

	stdout, stderr, code := runCLI(t, "status", root)

	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	if !strings.Contains(stdout, "Index status: not indexed") {
		t.Fatalf("got %q", stdout)
	}
	if !strings.Contains(stdout, "Stories: 0") {
		t.Fatalf("got %q", stdout)
	}
	if readChat(t, root) != chatBefore {
		t.Fatal("status changed fixture source file app/api/chat.py")
	}
}

func TestStatus_helpMentionsJSONAndPrivacy(t *testing.T) {
	stdout, stderr, code := runCLI(t, "status", "--help")
	if code != cli.ExitOK {
		t.Fatalf("exit = %d; stderr=%q", code, stderr)
	}
	for _, flag := range []string{"--json", "--privacy"} {
		if !strings.Contains(stdout, flag) {
			t.Fatalf("status --help missing %s\n%s", flag, stdout)
		}
	}
}

func copyFixtureAs(t *testing.T, name string) string {
	t.Helper()
	dst := filepath.Join(t.TempDir(), name)
	if err := os.MkdirAll(dst, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.CopyFS(dst, os.DirFS(fixtureRoot(t))); err != nil {
		t.Fatalf("copy fixture: %v", err)
	}
	return dst
}

func parseStatusJSON(t *testing.T, stdout string) map[string]any {
	t.Helper()
	var out map[string]any
	if err := json.Unmarshal([]byte(stdout), &out); err != nil {
		t.Fatalf("json: %v\n%s", err, stdout)
	}
	return out
}

func seedIndexRun(t *testing.T, root, runStatus string, stories int) {
	t.Helper()
	db, err := storage.Open(storage.DatabasePath(root))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if _, err := db.Exec(`
INSERT INTO repositories (id, name, root_path, config_version, status, created_at, updated_at)
VALUES ('repo-1', 'fastapi-rag-demo', ?, 1, 'ready', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
`, root); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`
INSERT INTO index_runs (id, repository_id, kind, status, started_at, finished_at)
VALUES ('run-1', 'repo-1', 'full', ?, '2026-01-01T00:00:00Z', '2026-01-01T00:00:01Z')
`, runStatus); err != nil {
		t.Fatal(err)
	}
	insertSeedStories(t, db, stories)
}

func insertSeedStories(t *testing.T, db *sql.DB, stories int) {
	t.Helper()
	for i := 0; i < stories; i++ {
		_, err := db.Exec(`
INSERT INTO stories (
    id, repository_id, key, title, intent, status, source_type, confidence,
    verification_status, created_at, updated_at
) VALUES (?, 'repo-1', ?, 'Chat', 'answer', 'draft', 'analyzer', 'high',
    'unverified', '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z')
`, fmt.Sprintf("story-%d", i), fmt.Sprintf("chat-%d", i))
		if err != nil {
			t.Fatal(err)
		}
	}
}
