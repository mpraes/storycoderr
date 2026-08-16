package repository

import (
	"database/sql"
	"fmt"
)

type FileRow struct {
	ID          string
	Path        string
	ContentHash string
}

type FileWrite struct {
	ID            string
	RepositoryID  string
	Path          string
	Language      string
	Kind          string
	ContentHash   string
	SizeBytes     int64
	LineCount     int
	IsTestFile    bool
	LastSeenRunID string
}

func ListFiles(tx *sql.Tx, repoID string) ([]FileRow, error) {
	rows, err := tx.Query(`
SELECT id, path, content_hash FROM source_files
WHERE repository_id = ? AND deleted_at IS NULL
`, repoID)
	if err != nil {
		return nil, fmt.Errorf("cannot list source_files for repository %q, expected source_files rows: %w", repoID, err)
	}
	defer rows.Close()
	return scanFileRows(rows, repoID)
}

func scanFileRows(rows *sql.Rows, repoID string) ([]FileRow, error) {
	var out []FileRow
	for rows.Next() {
		var row FileRow
		if err := rows.Scan(&row.ID, &row.Path, &row.ContentHash); err != nil {
			return nil, fmt.Errorf("cannot scan source_file for repository %q, expected id path content_hash: %w", repoID, err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("cannot iterate source_files for repository %q: %w", repoID, err)
	}
	return out, nil
}

func UpsertFile(tx *sql.Tx, file FileWrite) error {
	res, err := tx.Exec(`
UPDATE source_files SET
    language = ?, kind = ?, content_hash = ?, size_bytes = ?, line_count = ?,
    is_test_file = ?, last_seen_index_run_id = ?, deleted_at = NULL
WHERE repository_id = ? AND path = ?
`, file.Language, file.Kind, file.ContentHash, file.SizeBytes, file.LineCount,
		boolInt(file.IsTestFile), file.LastSeenRunID, file.RepositoryID, file.Path)
	if err != nil {
		return fmt.Errorf("cannot update source_file %q, expected unique (repository_id, path): %w", file.Path, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return insertFile(tx, file)
}

func insertFile(tx *sql.Tx, file FileWrite) error {
	_, err := tx.Exec(`
INSERT INTO source_files (
    id, repository_id, path, language, kind, content_hash, size_bytes, line_count,
    is_generated, is_test_file, is_ignored, last_seen_index_run_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 0, ?, 0, ?)
`, file.ID, file.RepositoryID, file.Path, file.Language, file.Kind, file.ContentHash,
		file.SizeBytes, file.LineCount, boolInt(file.IsTestFile), file.LastSeenRunID)
	if err != nil {
		return fmt.Errorf("cannot insert source_file %q, expected a new source_files row: %w", file.Path, err)
	}
	return nil
}

func FileID(tx *sql.Tx, repoID, path string) (string, error) {
	var id string
	err := tx.QueryRow(`SELECT id FROM source_files WHERE repository_id = ? AND path = ?`, repoID, path).Scan(&id)
	if err != nil {
		return "", fmt.Errorf("cannot load source_file %q, expected a source_files row: %w", path, err)
	}
	return id, nil
}

func MarkMissingFiles(tx *sql.Tx, repoID, runID string, seen map[string]bool) error {
	files, err := ListFiles(tx, repoID)
	if err != nil {
		return err
	}
	now := nowUTC()
	for _, file := range files {
		if seen[file.Path] {
			continue
		}
		if err := softDeleteFile(tx, file.ID, now); err != nil {
			return err
		}
	}
	return touchSeenFiles(tx, repoID, runID, seen)
}

func softDeleteFile(tx *sql.Tx, fileID, deletedAt string) error {
	_, err := tx.Exec(`UPDATE source_files SET deleted_at = ? WHERE id = ?`, deletedAt, fileID)
	if err != nil {
		return fmt.Errorf("cannot mark source_file %q missing, expected source_files.id: %w", fileID, err)
	}
	return nil
}

func touchSeenFiles(tx *sql.Tx, repoID, runID string, seen map[string]bool) error {
	_, err := tx.Exec(`
UPDATE source_files SET last_seen_index_run_id = ?
WHERE repository_id = ? AND deleted_at IS NULL
`, runID, repoID)
	if err != nil {
		return fmt.Errorf("cannot touch source_files for repository %q run %q: %w", repoID, runID, err)
	}
	_ = seen
	return nil
}

func boolInt(v bool) int {
	if v {
		return 1
	}
	return 0
}
