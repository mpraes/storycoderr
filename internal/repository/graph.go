package repository

import (
	"database/sql"
	"fmt"
)

type SymbolWrite struct {
	ID            string
	RepositoryID  string
	SourceFileID  string
	QualifiedName string
	DisplayName   string
	Kind          string
	StartLine     int
	EndLine       int
	SemanticHash  string
	Confidence    string
	LastSeenRunID string
}

type RelationWrite struct {
	ID            string
	RepositoryID  string
	FromSymbolID  string
	ToSymbolID    string
	ToExternalRef string
	Kind          string
	SourceFileID  string
	Line          int
	Confidence    string
	LastSeenRunID string
}

type EntryWrite struct {
	ID            string
	RepositoryID  string
	HandlerID     string
	Kind          string
	Key           string
	Label         string
	Method        string
	Path          string
	Framework     string
	Confidence    string
	LastSeenRunID string
}

func ClearPythonGraph(tx *sql.Tx, repoID string) error {
	if err := nullEntryHandlers(tx, repoID); err != nil {
		return err
	}
	if _, err := tx.Exec(`DELETE FROM code_relations WHERE repository_id = ?`, repoID); err != nil {
		return fmt.Errorf("cannot delete code_relations for repository %q: %w", repoID, err)
	}
	if _, err := tx.Exec(`DELETE FROM code_symbols WHERE repository_id = ?`, repoID); err != nil {
		return fmt.Errorf("cannot delete code_symbols for repository %q: %w", repoID, err)
	}
	return nil
}

func nullEntryHandlers(tx *sql.Tx, repoID string) error {
	_, err := tx.Exec(`UPDATE entry_points SET handler_symbol_id = NULL WHERE repository_id = ?`, repoID)
	if err != nil {
		return fmt.Errorf("cannot clear entry_points handlers for repository %q: %w", repoID, err)
	}
	return nil
}

func InsertSymbol(tx *sql.Tx, sym SymbolWrite) error {
	_, err := tx.Exec(`
INSERT INTO code_symbols (
    id, repository_id, source_file_id, qualified_name, display_name, kind,
    start_line, end_line, semantic_hash, source_type, confidence, last_seen_index_run_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'static_analysis', ?, ?)
`, sym.ID, sym.RepositoryID, sym.SourceFileID, sym.QualifiedName, sym.DisplayName, sym.Kind,
		sym.StartLine, sym.EndLine, sym.SemanticHash, sym.Confidence, sym.LastSeenRunID)
	if err != nil {
		return fmt.Errorf("cannot insert code_symbol %q, expected a new code_symbols row: %w", sym.QualifiedName, err)
	}
	return nil
}

func InsertRelation(tx *sql.Tx, rel RelationWrite) error {
	_, err := tx.Exec(`
INSERT INTO code_relations (
    id, repository_id, from_symbol_id, to_symbol_id, to_external_ref, kind,
    source_file_id, line, source_type, confidence, last_seen_index_run_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, 'static_analysis', ?, ?)
`, rel.ID, rel.RepositoryID, rel.FromSymbolID, nullIfEmpty(rel.ToSymbolID), nullIfEmpty(rel.ToExternalRef),
		rel.Kind, nullIfEmpty(rel.SourceFileID), rel.Line, rel.Confidence, rel.LastSeenRunID)
	if err != nil {
		return fmt.Errorf("cannot insert code_relation from %q, expected a new code_relations row: %w", rel.FromSymbolID, err)
	}
	return nil
}

func UpsertEntryPoint(tx *sql.Tx, entry EntryWrite) error {
	res, err := tx.Exec(`
UPDATE entry_points SET
    handler_symbol_id = ?, kind = ?, label = ?, method = ?, path = ?,
    framework = ?, confidence = ?, last_seen_index_run_id = ?, deleted_at = NULL
WHERE repository_id = ? AND key = ?
`, nullIfEmpty(entry.HandlerID), entry.Kind, entry.Label, entry.Method, entry.Path,
		entry.Framework, entry.Confidence, entry.LastSeenRunID, entry.RepositoryID, entry.Key)
	if err != nil {
		return fmt.Errorf("cannot update entry_point %q, expected unique (repository_id, key): %w", entry.Key, err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	return insertEntry(tx, entry)
}

func insertEntry(tx *sql.Tx, entry EntryWrite) error {
	_, err := tx.Exec(`
INSERT INTO entry_points (
    id, repository_id, handler_symbol_id, kind, key, label, method, path,
    framework, source_type, confidence, last_seen_index_run_id
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'static_analysis', ?, ?)
`, entry.ID, entry.RepositoryID, nullIfEmpty(entry.HandlerID), entry.Kind, entry.Key, entry.Label,
		entry.Method, entry.Path, entry.Framework, entry.Confidence, entry.LastSeenRunID)
	if err != nil {
		return fmt.Errorf("cannot insert entry_point %q, expected a new entry_points row: %w", entry.Key, err)
	}
	return nil
}

func TouchGraph(tx *sql.Tx, repoID, runID string) error {
	stmts := []string{
		`UPDATE code_symbols SET last_seen_index_run_id = ? WHERE repository_id = ? AND deleted_at IS NULL`,
		`UPDATE code_relations SET last_seen_index_run_id = ? WHERE repository_id = ? AND deleted_at IS NULL`,
		`UPDATE entry_points SET last_seen_index_run_id = ? WHERE repository_id = ? AND deleted_at IS NULL`,
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt, runID, repoID); err != nil {
			return fmt.Errorf("cannot touch graph for repository %q run %q: %w", repoID, runID, err)
		}
	}
	return nil
}

func LoadSymbolIDs(tx *sql.Tx, repoID string) (map[string]string, error) {
	rows, err := tx.Query(`
SELECT id, qualified_name FROM code_symbols
WHERE repository_id = ? AND deleted_at IS NULL
`, repoID)
	if err != nil {
		return nil, fmt.Errorf("cannot load code_symbols for repository %q: %w", repoID, err)
	}
	defer rows.Close()
	out := map[string]string{}
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, fmt.Errorf("cannot scan code_symbol for repository %q: %w", repoID, err)
		}
		out[name] = id
	}
	return out, rows.Err()
}

func CountActive(tx *sql.Tx, table, repoID string) (int, error) {
	var n int
	q := `SELECT COUNT(*) FROM ` + table + ` WHERE repository_id = ? AND deleted_at IS NULL`
	if err := tx.QueryRow(q, repoID).Scan(&n); err != nil {
		q = `SELECT COUNT(*) FROM ` + table + ` WHERE repository_id = ?`
		if err2 := tx.QueryRow(q, repoID).Scan(&n); err2 != nil {
			return 0, fmt.Errorf("cannot count %s for repository %q: %w", table, repoID, err)
		}
	}
	return n, nil
}
