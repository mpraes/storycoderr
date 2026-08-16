package indexer

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"storycode/internal/config"
	"storycode/internal/storage"
)

func TestRun_cancelledContextDoesNotWriteGraph(t *testing.T) {
	root := copyFixture(t)
	dbPath := storage.DatabasePath(root)
	if err := storage.EnsureFile(dbPath); err != nil {
		t.Fatal(err)
	}
	db, err := storage.Open(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err = Run(ctx, db, Options{Root: root, Settings: config.Defaults(), Out: &bytes.Buffer{}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	var n int
	if qerr := db.QueryRow(`SELECT COUNT(*) FROM source_files`).Scan(&n); qerr != nil {
		t.Fatal(qerr)
	}
	if n != 0 {
		t.Fatalf("cancelled index wrote %d source_files, want 0", n)
	}
}

func copyFixture(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	src := filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "fixtures", "fastapi-rag-demo"))
	dst := filepath.Join(t.TempDir(), "fastapi-rag-demo")
	if err := os.CopyFS(dst, os.DirFS(src)); err != nil {
		t.Fatal(err)
	}
	return dst
}
