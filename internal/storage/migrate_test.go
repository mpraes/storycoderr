package storage_test

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"storycode/internal/storage"
)

var mvpTables = []string{
	"repositories",
	"index_runs",
	"source_files",
	"code_symbols",
	"code_relations",
	"entry_points",
	"stories",
	"story_triggers",
	"story_actors",
	"story_paths",
	"scenes",
	"scene_transitions",
	"evidences",
	"evidence_references",
	"evidence_verifications",
}

func TestDatabasePath(t *testing.T) {
	got := storage.DatabasePath(filepath.Join("repo", "demo"))
	want := filepath.Join("repo", "demo", ".storycode", "index", "storycode.db")
	if got != want {
		t.Fatalf("DatabasePath = %q, want %q", got, want)
	}
}

func TestMigrate_appliesAllTablesOnTempDatabase(t *testing.T) {
	db := openTempDB(t)

	if err := storage.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	got := tableNames(t, db)
	for _, name := range mvpTables {
		if !got[name] {
			t.Errorf("missing table %s", name)
		}
	}
	if !got["schema_migrations"] {
		t.Fatal("missing schema_migrations")
	}
}

func TestMigrate_isIdempotent(t *testing.T) {
	db := openTempDB(t)
	if err := storage.Migrate(db); err != nil {
		t.Fatalf("first Migrate: %v", err)
	}
	if err := storage.Migrate(db); err != nil {
		t.Fatalf("second Migrate: %v", err)
	}
	if got := migrationVersions(t, db); len(got) != 1 || got[0] != 1 {
		t.Fatalf("versions = %v, want [1]", got)
	}
}

func TestMigrate_failedFilePreservesPreviousSchema(t *testing.T) {
	db := openTempDB(t)
	ok := fstest.MapFS{
		"0001_ok.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE keep_me (id TEXT PRIMARY KEY NOT NULL);\n",
		)},
	}
	if err := storage.MigrateFS(db, ok); err != nil {
		t.Fatalf("first MigrateFS: %v", err)
	}

	broken := fstest.MapFS{
		"0001_ok.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE keep_me (id TEXT PRIMARY KEY NOT NULL);\n",
		)},
		"0002_bad.sql": &fstest.MapFile{Data: []byte(
			"CREATE TABLE nope (id TEXT PRIMARY KEY NOT NULL);\n" +
				"CREATE TABLE broken (;\n",
		)},
	}
	err := storage.MigrateFS(db, broken)
	if err == nil {
		t.Fatal("expected failed migration")
	}
	if !strings.Contains(err.Error(), "0002_bad.sql") {
		t.Fatalf("error should mention 0002_bad.sql, got %v", err)
	}

	got := tableNames(t, db)
	if !got["keep_me"] {
		t.Fatal("previous table keep_me was dropped")
	}
	if got["nope"] {
		t.Fatal("partial migration table nope should not remain")
	}
	if versions := migrationVersions(t, db); len(versions) != 1 || versions[0] != 1 {
		t.Fatalf("versions = %v, want [1]", versions)
	}
}

func TestEnsureFile_createsDatabaseFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "index", "storycode.db")
	if err := storage.EnsureFile(path); err != nil {
		t.Fatalf("EnsureFile: %v", err)
	}
	db, err := storage.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	got := tableNames(t, db)
	if !got["repositories"] {
		t.Fatal("EnsureFile did not apply migrations")
	}
}

func TestOpen_pathWithSpaces(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "rag demo")
	path := filepath.Join(dir, "storycode.db")
	db, err := storage.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := storage.Migrate(db); err != nil {
		t.Fatalf("Migrate: %v", err)
	}
}

func openTempDB(t *testing.T) *sql.DB {
	t.Helper()
	path := filepath.Join(t.TempDir(), "storycode.db")
	db, err := storage.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func tableNames(t *testing.T, db *sql.DB) map[string]bool {
	t.Helper()
	rows, err := db.Query(`SELECT name FROM sqlite_master WHERE type = 'table'`)
	if err != nil {
		t.Fatalf("list tables: %v", err)
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("scan table name: %v", err)
		}
		out[name] = true
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate tables: %v", err)
	}
	return out
}

func migrationVersions(t *testing.T, db *sql.DB) []int {
	t.Helper()
	rows, err := db.Query(`SELECT version FROM schema_migrations ORDER BY version`)
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	defer rows.Close()
	var out []int
	for rows.Next() {
		var version int
		if err := rows.Scan(&version); err != nil {
			t.Fatalf("scan version: %v", err)
		}
		out = append(out, version)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("iterate versions: %v", err)
	}
	return out
}

func TestMigrateFS_rejectsUnversionedName(t *testing.T) {
	db := openTempDB(t)
	files := fstest.MapFS{
		"init.sql": &fstest.MapFile{Data: []byte("CREATE TABLE x (id TEXT);")},
	}
	err := storage.MigrateFS(db, files)
	if err == nil {
		t.Fatal("expected error for unversioned migration name")
	}
	if !strings.Contains(err.Error(), "init.sql") {
		t.Fatalf("error should mention init.sql, got %v", err)
	}
}
