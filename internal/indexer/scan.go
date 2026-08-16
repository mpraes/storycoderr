package indexer

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type ScanOptions struct {
	Root             string
	Include          []string
	Exclude          []string
	FollowSymlinks   bool
	MaxFileSizeBytes int64
}

type FoundFile struct {
	Path      string
	SizeBytes int64
}

type ScanWarning struct {
	Path    string
	Message string
}

type ScanResult struct {
	Files    []FoundFile
	Warnings []ScanWarning
}

// Scan lists repository files that match include/exclude globs.
// It stats entries only and never reads file contents or executes them.
//
//	result, err := Scan(ScanOptions{Root: dir, Include: []string{"**/*.py"}})
func Scan(opts ScanOptions) (ScanResult, error) {
	if err := validateScanOptions(opts); err != nil {
		return ScanResult{}, err
	}
	var result ScanResult
	walkErr := filepath.WalkDir(opts.Root, func(path string, d fs.DirEntry, err error) error {
		return visitScanPath(&result, opts, path, d, err)
	})
	if walkErr != nil {
		return ScanResult{}, fmt.Errorf("cannot scan repository %q: %w, expected a readable directory", opts.Root, walkErr)
	}
	sort.Slice(result.Files, func(i, j int) bool {
		return result.Files[i].Path < result.Files[j].Path
	})
	return result, nil
}

func validateScanOptions(opts ScanOptions) error {
	if strings.TrimSpace(opts.Root) == "" {
		return fmt.Errorf("scan root %q is empty, expected an existing directory path", opts.Root)
	}
	if len(opts.Include) == 0 {
		return fmt.Errorf("scan include %v is empty, expected at least one glob such as **/*.py", opts.Include)
	}
	return nil
}

func visitScanPath(result *ScanResult, opts ScanOptions, path string, d fs.DirEntry, err error) error {
	rel, relErr := relativeSlash(opts.Root, path)
	if relErr != nil {
		result.Warnings = append(result.Warnings, walkWarning(path, relErr))
		return skipOnDir(d)
	}
	if err != nil {
		result.Warnings = append(result.Warnings, walkWarning(rel, err))
		return skipOnDir(d)
	}
	if rel == "." {
		return nil
	}
	if d.IsDir() {
		return visitDir(rel, d.Name(), opts.Exclude)
	}
	return visitFile(result, opts, path, rel, d)
}

func visitDir(rel, name string, excludes []string) error {
	if skipDirName(name) || excludedPath(rel, excludes) {
		return fs.SkipDir
	}
	return nil
}

func visitFile(result *ScanResult, opts ScanOptions, abs, rel string, d fs.DirEntry) error {
	if isSymlink(d) && !opts.FollowSymlinks {
		return nil
	}
	if excludedPath(rel, opts.Exclude) || !includedPath(rel, opts.Include) {
		return nil
	}
	info, err := fileInfo(opts.FollowSymlinks, abs, d)
	if err != nil {
		result.Warnings = append(result.Warnings, walkWarning(rel, err))
		return nil
	}
	if opts.MaxFileSizeBytes > 0 && info.Size() > opts.MaxFileSizeBytes {
		result.Warnings = append(result.Warnings, oversizeWarning(rel, info.Size(), opts.MaxFileSizeBytes))
		return nil
	}
	result.Files = append(result.Files, FoundFile{Path: rel, SizeBytes: info.Size()})
	return nil
}

func fileInfo(follow bool, abs string, d fs.DirEntry) (fs.FileInfo, error) {
	if follow {
		return os.Stat(abs)
	}
	return d.Info()
}

func isSymlink(d fs.DirEntry) bool {
	return d.Type()&fs.ModeSymlink != 0
}

func skipOnDir(d fs.DirEntry) error {
	if d != nil && d.IsDir() {
		return fs.SkipDir
	}
	return nil
}

func skipDirName(name string) bool {
	switch name {
	case ".git", ".venv", "venv", "__pycache__", "node_modules":
		return true
	default:
		return false
	}
}

func excludedPath(rel string, excludes []string) bool {
	return matchesAny(rel, excludes)
}

func includedPath(rel string, includes []string) bool {
	return matchesAny(rel, includes)
}

func matchesAny(rel string, patterns []string) bool {
	for _, pattern := range patterns {
		if matchGlob(pattern, rel) {
			return true
		}
	}
	return false
}

func relativeSlash(root, path string) (string, error) {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return "", err
	}
	return slashPath(rel), nil
}

func walkWarning(rel string, err error) ScanWarning {
	return ScanWarning{
		Path:    rel,
		Message: fmt.Sprintf("cannot read path %q: %v, expected a readable file or directory", rel, err),
	}
}

func oversizeWarning(rel string, size, max int64) ScanWarning {
	return ScanWarning{
		Path:    rel,
		Message: fmt.Sprintf("skipped file %q size %d bytes, expected at most %d bytes", rel, size, max),
	}
}
