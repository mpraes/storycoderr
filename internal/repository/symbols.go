package repository

import (
	"database/sql"
	"fmt"
)

type SymbolHit struct {
	ID            string
	QualifiedName string
	DisplayName   string
	Path          string
}

func LoadSymbolHits(tx *sql.Tx, repoID string) ([]SymbolHit, error) {
	rows, err := tx.Query(`
SELECT s.id, s.qualified_name, s.display_name, f.path
FROM code_symbols s
JOIN source_files f ON f.id = s.source_file_id
WHERE s.repository_id = ? AND s.deleted_at IS NULL
`, repoID)
	if err != nil {
		return nil, fmt.Errorf("cannot load symbols with paths for repository %q: %w", repoID, err)
	}
	defer rows.Close()
	return scanSymbolHits(rows, repoID)
}

func scanSymbolHits(rows *sql.Rows, repoID string) ([]SymbolHit, error) {
	var out []SymbolHit
	for rows.Next() {
		var hit SymbolHit
		if err := rows.Scan(&hit.ID, &hit.QualifiedName, &hit.DisplayName, &hit.Path); err != nil {
			return nil, fmt.Errorf("cannot scan symbol hit for repository %q: %w", repoID, err)
		}
		out = append(out, hit)
	}
	return out, rows.Err()
}
