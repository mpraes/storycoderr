package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"storycode/internal/config"
)

func TestParse_defaultYAMLHasPythonGlobs(t *testing.T) {
	s, err := config.Parse(config.DefaultYAML)
	if err != nil {
		t.Fatal(err)
	}
	if s.Version != 1 {
		t.Fatalf("version = %d", s.Version)
	}
	if len(s.Include) != 3 {
		t.Fatalf("include = %v", s.Include)
	}
	if s.FollowSymlinks {
		t.Fatal("follow_symlinks should be false")
	}
	if s.MaxFileSizeBytes != 5242880 {
		t.Fatalf("max_file_size_bytes = %d", s.MaxFileSizeBytes)
	}
}

func TestLoadFile_missingUsesDefaults(t *testing.T) {
	s, err := config.LoadFile(filepath.Join(t.TempDir(), "missing.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	if s.Include[0] != "**/*.py" {
		t.Fatalf("include = %v", s.Include)
	}
}

func TestLoadFile_readsWrittenConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(config.DefaultYAML), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := config.LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if s.Exclude[0] != ".git/**" {
		t.Fatalf("exclude = %v", s.Exclude)
	}
}
