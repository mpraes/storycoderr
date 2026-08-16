package indexer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"

	"storycode/internal/config"
	"storycode/internal/repository"
)

type Options struct {
	Root     string
	Settings config.Settings
	Out      io.Writer
}

type indexedFile struct {
	Path        string
	SizeBytes   int64
	Content     []byte
	ContentHash string
	LineCount   int
	ReadError   string
}

type warning struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// Run indexes a repository into SQLite without executing analyzed files.
//
//	err := indexer.Run(ctx, db, indexer.Options{Root: dir, Settings: config.Defaults(), Out: os.Stdout})
func Run(ctx context.Context, db *sql.DB, opts Options) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	report(opts.Out, "Scanning files...")
	scanned, err := Scan(scanOptions(opts))
	if err != nil {
		return err
	}
	return applyIndex(ctx, db, opts, scanned)
}

func scanOptions(opts Options) ScanOptions {
	return ScanOptions{
		Root:             opts.Root,
		Include:          opts.Settings.Include,
		Exclude:          opts.Settings.Exclude,
		FollowSymlinks:   opts.Settings.FollowSymlinks,
		MaxFileSizeBytes: opts.Settings.MaxFileSizeBytes,
	}
}

func applyIndex(ctx context.Context, db *sql.DB, opts Options, scanned ScanResult) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("cannot begin index transaction for %q, expected a writable sqlite connection: %w", opts.Root, err)
	}
	defer func() { _ = tx.Rollback() }()
	if err := persistIndex(ctx, tx, opts, scanned); err != nil {
		return err
	}
	return tx.Commit()
}

func persistIndex(ctx context.Context, tx *sql.Tx, opts Options, scanned ScanResult) error {
	repoID, runID, err := startRun(tx, opts)
	if err != nil {
		return err
	}
	files, warns := readIndexedFiles(opts.Root, scanned)
	warns = append(scanWarnings(scanned), warns...)
	if err := ctx.Err(); err != nil {
		return err
	}
	changedPython, err := persistFiles(tx, repoID, runID, files, opts)
	if err != nil {
		return err
	}
	kind := runKind(changedPython, existingGraph(tx, repoID))
	if err := persistGraph(ctx, tx, opts, repoID, runID, files, changedPython, kind); err != nil {
		return err
	}
	return finish(tx, repoID, runID, kind, files, warns)
}

func startRun(tx *sql.Tx, opts Options) (string, string, error) {
	repoID, err := repository.EnsureRepository(tx, filepath.Base(opts.Root), opts.Root, opts.Settings.Version)
	if err != nil {
		return "", "", err
	}
	runID, err := newRunID()
	if err != nil {
		return "", "", err
	}
	run := repository.IndexRun{
		ID:           runID,
		RepositoryID: repoID,
		Kind:         "full",
		Status:       "running",
		StartedAt:    repositoryNow(),
	}
	if err := repository.InsertRun(tx, run); err != nil {
		return "", "", err
	}
	return repoID, runID, nil
}

func runKind(changedPython bool, hadGraph bool) string {
	if hadGraph && !changedPython {
		return "incremental"
	}
	return "full"
}

func existingGraph(tx *sql.Tx, repoID string) bool {
	n, err := repository.CountActive(tx, "code_symbols", repoID)
	return err == nil && n > 0
}

func finish(tx *sql.Tx, repoID, runID, kind string, files []indexedFile, warns []warning) error {
	status := "completed"
	if len(warns) > 0 {
		status = "completed_with_warnings"
	}
	symbols, _ := repository.CountActive(tx, "code_symbols", repoID)
	relations, _ := repository.CountActive(tx, "code_relations", repoID)
	run := repository.IndexRun{
		ID:             runID,
		RepositoryID:   repoID,
		Kind:           kind,
		Status:         status,
		FilesScanned:   len(files),
		FilesIndexed:   countIndexed(files),
		FilesFailed:    countFailed(files),
		SymbolsFound:   symbols,
		RelationsFound: relations,
		WarningsJSON:   warningsJSON(warns),
	}
	if err := repository.FinishRun(tx, run); err != nil {
		return err
	}
	return repository.SetRepositoryStatus(tx, repoID, "ready")
}

func warningsJSON(warns []warning) string {
	if len(warns) == 0 {
		return ""
	}
	body, err := json.Marshal(warns)
	if err != nil {
		return ""
	}
	return string(body)
}

func report(out io.Writer, line string) {
	if out == nil {
		return
	}
	fmt.Fprintln(out, line)
}
