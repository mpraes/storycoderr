package storage

import (
	"database/sql"
	"fmt"
	"net/url"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

// DatabasePath returns .storycode/index/storycode.db under repoRoot.
//
//	path := storage.DatabasePath("/repo")
func DatabasePath(repoRoot string) string {
	return filepath.Join(repoRoot, ".storycode", "index", "storycode.db")
}

// EnsureFile creates parent dirs, opens the database, and applies migrations.
//
//	err := storage.EnsureFile(storage.DatabasePath(root))
func EnsureFile(path string) error {
	db, err := Open(path)
	if err != nil {
		return err
	}
	defer db.Close()
	return Migrate(db)
}

// Open creates or opens a SQLite database at path without CGO.
//
//	db, err := storage.Open(filepath.Join(dir, "storycode.db"))
func Open(path string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, fmt.Errorf("cannot create sqlite directory %q: %w (expected a writable directory)", filepath.Dir(path), err)
	}
	dsn, err := sqliteDSN(path)
	if err != nil {
		return nil, err
	}
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("cannot open sqlite %q, expected a file DSN: %w", dsn, err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("cannot ping sqlite %q, expected a writable database file: %w", path, err)
	}
	return db, nil
}

func sqliteDSN(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve sqlite path %q, expected an absolute filesystem path: %w", path, err)
	}
	u := url.URL{Scheme: "file", Path: filepath.ToSlash(abs)}
	u.RawQuery = "_pragma=foreign_keys(1)"
	return u.String(), nil
}
