package indexer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"storycode/internal/repository"
)

func readIndexedFiles(root string, scanned ScanResult) ([]indexedFile, []warning) {
	var files []indexedFile
	var warns []warning
	for _, found := range scanned.Files {
		file, warn := readOneFile(root, found)
		files = append(files, file)
		if warn != nil {
			warns = append(warns, *warn)
		}
	}
	return files, warns
}

func readOneFile(root string, found FoundFile) (indexedFile, *warning) {
	file := indexedFile{Path: found.Path, SizeBytes: found.SizeBytes}
	body, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(found.Path)))
	if err != nil {
		file.ReadError = err.Error()
		warn := warning{
			Path:    found.Path,
			Message: fmt.Sprintf("cannot read file %q: %v, expected a readable indexed file", found.Path, err),
		}
		return file, &warn
	}
	file.Content = body
	file.ContentHash = hashBytes(body)
	file.LineCount = countLines(body)
	return file, nil
}

func hashBytes(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func countLines(body []byte) int {
	if len(body) == 0 {
		return 0
	}
	n := 1
	for i := 0; i < len(body); i++ {
		if body[i] == '\n' {
			n++
		}
	}
	if body[len(body)-1] == '\n' {
		n--
	}
	return n
}

func scanWarnings(scanned ScanResult) []warning {
	var out []warning
	for _, w := range scanned.Warnings {
		out = append(out, warning{Path: w.Path, Message: w.Message})
	}
	return out
}

func countIndexed(files []indexedFile) int {
	n := 0
	for _, file := range files {
		if file.ReadError == "" {
			n++
		}
	}
	return n
}

func countFailed(files []indexedFile) int {
	return len(files) - countIndexed(files)
}

func newRunID() (string, error) {
	return repository.NewID()
}

func repositoryNow() string {
	return repository.NowUTC()
}

func isPython(path string) bool {
	return strings.HasSuffix(strings.ToLower(path), ".py")
}

func fileLanguage(path string) string {
	switch {
	case isPython(path):
		return "python"
	case strings.HasSuffix(strings.ToLower(path), ".md"):
		return "markdown"
	default:
		return ""
	}
}

func fileKind(path string) string {
	slash := strings.ReplaceAll(path, "\\", "/")
	if strings.Contains(slash, "/tests/") || strings.HasPrefix(slash, "tests/") {
		return "test"
	}
	if strings.HasSuffix(strings.ToLower(slash), ".md") {
		return "documentation"
	}
	return "source_code"
}

func isTestFile(path string) bool {
	return fileKind(path) == "test"
}
