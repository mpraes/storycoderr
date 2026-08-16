package repository_test

import (
	"path/filepath"
	"testing"

	"storycode/internal/repository"
	"storycode/internal/storage"
)

func TestNewID_returnsDistinctHex(t *testing.T) {
	a, err := repository.NewID()
	if err != nil {
		t.Fatal(err)
	}
	b, err := repository.NewID()
	if err != nil {
		t.Fatal(err)
	}
	if len(a) != 32 || len(b) != 32 {
		t.Fatalf("ids %q %q, expected 32 hex characters", a, b)
	}
	if a == b {
		t.Fatal("expected distinct ids")
	}
}

func TestEnsureRepository_isIdempotentByRoot(t *testing.T) {
	db, err := storage.Open(filepath.Join(t.TempDir(), "storycode.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := storage.Migrate(db); err != nil {
		t.Fatal(err)
	}
	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback() }()
	first, err := repository.EnsureRepository(tx, "demo", "/repo", 1)
	if err != nil {
		t.Fatal(err)
	}
	second, err := repository.EnsureRepository(tx, "demo", "/repo", 1)
	if err != nil {
		t.Fatal(err)
	}
	if first != second {
		t.Fatalf("ids %s vs %s, expected the same repository row", first, second)
	}
}
