package indexer

import (
	"database/sql"

	"storycode/internal/repository"
)

func persistFiles(tx *sql.Tx, repoID, runID string, files []indexedFile, opts Options) (bool, error) {
	report(opts.Out, "Persisting source files...")
	previous, err := repository.ListFiles(tx, repoID)
	if err != nil {
		return false, err
	}
	prevHash := fileHashMap(previous)
	changedPython := false
	seen := map[string]bool{}
	for _, file := range files {
		if file.ReadError != "" {
			continue
		}
		seen[file.Path] = true
		if isPython(file.Path) && prevHash[file.Path] != file.ContentHash {
			changedPython = true
		}
		if err := upsertIndexedFile(tx, repoID, runID, file); err != nil {
			return false, err
		}
	}
	if err := repository.MarkMissingFiles(tx, repoID, runID, seen); err != nil {
		return false, err
	}
	return changedPython || pythonMissing(previous, seen), nil
}

func upsertIndexedFile(tx *sql.Tx, repoID, runID string, file indexedFile) error {
	id, err := repository.NewID()
	if err != nil {
		return err
	}
	return repository.UpsertFile(tx, repository.FileWrite{
		ID:            id,
		RepositoryID:  repoID,
		Path:          file.Path,
		Language:      fileLanguage(file.Path),
		Kind:          fileKind(file.Path),
		ContentHash:   file.ContentHash,
		SizeBytes:     file.SizeBytes,
		LineCount:     file.LineCount,
		IsTestFile:    isTestFile(file.Path),
		LastSeenRunID: runID,
	})
}

func fileHashMap(rows []repository.FileRow) map[string]string {
	out := map[string]string{}
	for _, row := range rows {
		out[row.Path] = row.ContentHash
	}
	return out
}

func pythonMissing(previous []repository.FileRow, seen map[string]bool) bool {
	for _, row := range previous {
		if isPython(row.Path) && !seen[row.Path] {
			return true
		}
	}
	return false
}
