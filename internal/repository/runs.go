package repository

import (
	"database/sql"
	"fmt"
	"time"
)

type IndexRun struct {
	ID             string
	RepositoryID   string
	Kind           string
	Status         string
	StartedAt      string
	FilesScanned   int
	FilesIndexed   int
	FilesFailed    int
	SymbolsFound   int
	RelationsFound int
	WarningsJSON   string
}

func EnsureRepository(tx *sql.Tx, name, rootPath string, configVersion int) (string, error) {
	var id string
	err := tx.QueryRow(`SELECT id FROM repositories WHERE root_path = ?`, rootPath).Scan(&id)
	if err == nil {
		return id, nil
	}
	if err != sql.ErrNoRows {
		return "", fmt.Errorf("cannot load repository %q, expected a repositories row: %w", rootPath, err)
	}
	id, err = NewID()
	if err != nil {
		return "", err
	}
	now := nowUTC()
	_, err = tx.Exec(`
INSERT INTO repositories (id, name, root_path, config_version, status, created_at, updated_at)
VALUES (?, ?, ?, ?, 'indexing', ?, ?)
`, id, name, rootPath, configVersion, now, now)
	if err != nil {
		return "", fmt.Errorf("cannot insert repository %q, expected a new repositories row: %w", rootPath, err)
	}
	return id, nil
}

func InsertRun(tx *sql.Tx, run IndexRun) error {
	_, err := tx.Exec(`
INSERT INTO index_runs (
    id, repository_id, kind, status, started_at,
    files_scanned, files_indexed, files_failed, symbols_found, relations_found, warnings
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
`, run.ID, run.RepositoryID, run.Kind, run.Status, run.StartedAt,
		run.FilesScanned, run.FilesIndexed, run.FilesFailed, run.SymbolsFound, run.RelationsFound, nullIfEmpty(run.WarningsJSON))
	if err != nil {
		return fmt.Errorf("cannot insert index_run %q, expected a new index_runs row: %w", run.ID, err)
	}
	return nil
}

func FinishRun(tx *sql.Tx, run IndexRun) error {
	_, err := tx.Exec(`
UPDATE index_runs SET
    kind = ?, status = ?, finished_at = ?,
    files_scanned = ?, files_indexed = ?, files_failed = ?,
    symbols_found = ?, relations_found = ?, warnings = ?
WHERE id = ?
`, run.Kind, run.Status, nowUTC(),
		run.FilesScanned, run.FilesIndexed, run.FilesFailed,
		run.SymbolsFound, run.RelationsFound, nullIfEmpty(run.WarningsJSON), run.ID)
	if err != nil {
		return fmt.Errorf("cannot finish index_run %q, expected an existing index_runs row: %w", run.ID, err)
	}
	return nil
}

func SetRepositoryStatus(tx *sql.Tx, repoID, status string) error {
	_, err := tx.Exec(`UPDATE repositories SET status = ?, updated_at = ? WHERE id = ?`, status, nowUTC(), repoID)
	if err != nil {
		return fmt.Errorf("cannot update repository %q status %q, expected repositories.id: %w", repoID, status, err)
	}
	return nil
}

func NowUTC() string {
	return time.Now().UTC().Format(time.RFC3339)
}

func nowUTC() string {
	return NowUTC()
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
