package cli

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"storycode/internal/config"
	"storycode/internal/indexer"
	"storycode/internal/storage"
)

func indexCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "index [dir]",
		Short: "Index the local repository",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runIndex,
	}
}

func runIndex(cmd *cobra.Command, args []string) error {
	root, err := initRoot(args, osFilesystem{})
	if err != nil {
		return err
	}
	return indexRepository(cmd, root)
}

func indexRepository(cmd *cobra.Command, root string) error {
	if err := storage.EnsureFile(storage.DatabasePath(root)); err != nil {
		return err
	}
	settings, err := config.LoadFile(filepath.Join(root, ".storycode", "config.yaml"))
	if err != nil {
		return err
	}
	db, err := storage.Open(storage.DatabasePath(root))
	if err != nil {
		return err
	}
	defer db.Close()
	err = indexer.Run(cmd.Context(), db, indexer.Options{
		Root:     root,
		Settings: settings,
		Out:      cmd.OutOrStdout(),
	})
	if err != nil {
		return fmt.Errorf("cannot index repository %q: %w (expected a readable project with Python sources)", root, err)
	}
	return nil
}
