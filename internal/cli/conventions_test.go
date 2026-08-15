package cli_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConventions_makefileExposesFormatLintAndBuild(t *testing.T) {
	body := readRepoFile(t, "Makefile")
	for _, target := range []string{"format:", "format-check:", "lint:", "test:", "build:"} {
		if !strings.Contains(body, target) {
			t.Errorf("Makefile missing target %s", target)
		}
	}
	if !strings.Contains(body, "gofmt") {
		t.Error("Makefile should format Go with gofmt")
	}
	if !strings.Contains(body, "prettier") {
		t.Error("Makefile should format frontend with prettier")
	}
}

func TestConventions_ciRunsFormatLintTestAndBuild(t *testing.T) {
	body := readRepoFile(t, ".github/workflows/ci.yml")
	for _, step := range []string{"format-check", "lint", "test", "build"} {
		if !strings.Contains(body, "make "+step) {
			t.Errorf("CI missing make %s", step)
		}
	}
}

func TestConventions_codingStandardsDocumentRequiredRules(t *testing.T) {
	body := readRepoFile(t, "docs/development/coding-standards.md")
	for _, needle := range []string{
		"gofmt",
		"Prettier",
		"English",
		"snake_case",
	} {
		if !strings.Contains(body, needle) {
			t.Errorf("coding-standards.md missing %q", needle)
		}
	}
}

func TestConventions_fixtureStillParsedWithoutExecutingPython(t *testing.T) {
	TestRepoLayout_fixtureIsReadableWithoutExecutingPython(t)
}

func readRepoFile(t *testing.T, rel string) string {
	t.Helper()
	path := filepath.Join(moduleRoot(t), filepath.FromSlash(rel))
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", rel, err)
	}
	return string(body)
}
