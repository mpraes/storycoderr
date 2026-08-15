package cli

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"storycode/internal/storage"
)

func statusCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status [dir]",
		Short: "Show repository index and story status",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runStatus,
	}
	cmd.Flags().Bool("json", false, "Print status as JSON")
	cmd.Flags().Bool("privacy", false, "Hide absolute filesystem paths")
	return cmd
}

func runStatus(cmd *cobra.Command, args []string) error {
	asJSON, err := cmd.Flags().GetBool("json")
	if err != nil {
		return fmt.Errorf("cannot read --json %v, expected a bool flag: %w", args, err)
	}
	root, err := initRoot(args, osFilesystem{})
	if err != nil {
		return err
	}
	report, err := collectStatus(root)
	if err != nil {
		return err
	}
	return writeStatus(cmd.OutOrStdout(), report, asJSON)
}

type statusReport struct {
	Repository  string `json:"repository"`
	IndexStatus string `json:"index_status"`
	Stories     int    `json:"stories"`
	Database    string `json:"database"`
	Config      string `json:"config"`
}

func collectStatus(root string) (statusReport, error) {
	report := statusReport{
		Repository:  filepath.Base(root),
		IndexStatus: "not_indexed",
		Database:    filepath.ToSlash(filepath.Join(".storycode", "index", "storycode.db")),
		Config:      filepath.ToSlash(filepath.Join(".storycode", "config.yaml")),
	}
	count, indexed, err := readIndexState(storage.DatabasePath(root))
	if err != nil {
		return statusReport{}, err
	}
	report.Stories = count
	if indexed {
		report.IndexStatus = "indexed"
	}
	return report, nil
}

func readIndexState(dbPath string) (int, bool, error) {
	if _, err := os.Stat(dbPath); err != nil {
		if os.IsNotExist(err) {
			return 0, false, nil
		}
		return 0, false, fmt.Errorf("cannot stat database %q: %w (expected a readable sqlite file)", dbPath, err)
	}
	return queryIndexState(dbPath)
}

func queryIndexState(dbPath string) (int, bool, error) {
	db, err := storage.Open(dbPath)
	if err != nil {
		return 0, false, err
	}
	defer db.Close()
	count, err := countStories(db)
	if err != nil {
		return 0, false, err
	}
	indexed, err := hasCompletedIndex(db)
	if err != nil {
		return 0, false, err
	}
	return count, indexed, nil
}

func countStories(db *sql.DB) (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM stories`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("cannot count stories, expected table stories: %w", err)
	}
	return count, nil
}

func hasCompletedIndex(db *sql.DB) (bool, error) {
	var n int
	err := db.QueryRow(`
SELECT COUNT(*) FROM index_runs
WHERE status IN ('completed', 'completed_with_warnings')
`).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("cannot read index_runs.status, expected completed or completed_with_warnings: %w", err)
	}
	return n > 0, nil
}

func writeStatus(out io.Writer, report statusReport, asJSON bool) error {
	if asJSON {
		return writeJSONStatus(out, report)
	}
	writeHumanStatus(out, report)
	return nil
}

func writeJSONStatus(out io.Writer, report statusReport) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("cannot encode status JSON for repository %q, expected statusReport object: %w", report.Repository, err)
	}
	return nil
}

func writeHumanStatus(out io.Writer, report statusReport) {
	fmt.Fprintf(out, "Repository: %s\n", report.Repository)
	fmt.Fprintf(out, "Index status: %s\n", humanIndexStatus(report.IndexStatus))
	fmt.Fprintf(out, "Stories: %d\n", report.Stories)
	fmt.Fprintf(out, "Database: %s\n", report.Database)
	fmt.Fprintf(out, "Config: %s\n", report.Config)
}

func humanIndexStatus(status string) string {
	if status == "not_indexed" {
		return "not indexed"
	}
	return status
}
