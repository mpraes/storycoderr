package cli

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"storycode/internal/storage"
)

const defaultConfigYAML = `version: 1

repository:
  include:
    - "**/*.py"
    - "tests/**/*.py"
    - "docs/**/*.md"
  exclude:
    - ".git/**"
    - ".venv/**"
    - "venv/**"
    - "__pycache__/**"
    - "node_modules/**"

analysis:
  languages:
    - python
  follow_symlinks: false
  max_file_size_bytes: 5242880

storage:
  mode: repository
  engine: sqlite
`

func initCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init [dir]",
		Short: "Initialize StoryCode in a local repository",
		Args:  cobra.MaximumNArgs(1),
		RunE:  runInit,
	}
	cmd.Flags().Bool("force", false, "Overwrite existing .storycode/config.yaml")
	return cmd
}

type storycodeFS interface {
	MkdirAll(path string, perm os.FileMode) error
	WriteFile(path string, data []byte, perm os.FileMode) error
	Stat(path string) (os.FileInfo, error)
	Getwd() (string, error)
	Abs(path string) (string, error)
}

type osFilesystem struct{}

func (osFilesystem) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (osFilesystem) WriteFile(path string, data []byte, perm os.FileMode) error {
	return os.WriteFile(path, data, perm)
}

func (osFilesystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(path)
}

func (osFilesystem) Getwd() (string, error) {
	return os.Getwd()
}

func (osFilesystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

func runInit(cmd *cobra.Command, args []string) error {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return fmt.Errorf("cannot read --force %v, expected a bool flag: %w", args, err)
	}
	root, err := initRoot(args, osFilesystem{})
	if err != nil {
		return err
	}
	return initializeStorycode(cmd.OutOrStdout(), root, force, osFilesystem{})
}

func initRoot(args []string, files storycodeFS) (string, error) {
	if len(args) == 1 {
		return absInitRoot(args[0], files)
	}
	wd, err := files.Getwd()
	if err != nil {
		return "", fmt.Errorf("cannot resolve working directory %q, expected a readable directory: %w", args, err)
	}
	return wd, nil
}

func absInitRoot(path string, files storycodeFS) (string, error) {
	abs, err := files.Abs(path)
	if err != nil {
		return "", fmt.Errorf("cannot resolve init directory %q, expected an absolute filesystem path: %w", path, err)
	}
	return abs, nil
}

func initializeStorycode(out io.Writer, root string, force bool, files storycodeFS) error {
	storyDir := filepath.Join(root, ".storycode")
	if err := createStorycodeDirs(storyDir, files); err != nil {
		return err
	}
	if err := storage.EnsureFile(storage.DatabasePath(root)); err != nil {
		return err
	}
	wrote, err := writeConfig(filepath.Join(storyDir, "config.yaml"), force, files)
	if err != nil {
		return err
	}
	reportInit(out, storyDir, wrote)
	return nil
}

func createStorycodeDirs(storyDir string, files storycodeFS) error {
	for _, name := range []string{"", "stories", "index", "cache"} {
		path := joinStorycodePath(storyDir, name)
		if err := files.MkdirAll(path, 0o755); err != nil {
			return fmt.Errorf("cannot create directory %q: %w (expected a writable directory)", path, err)
		}
	}
	return nil
}

func joinStorycodePath(storyDir, name string) string {
	if name == "" {
		return storyDir
	}
	return filepath.Join(storyDir, name)
}

func writeConfig(path string, force bool, files storycodeFS) (bool, error) {
	if !force {
		exists, err := configExists(path, files)
		if err != nil {
			return false, err
		}
		if exists {
			return false, nil
		}
	}
	if err := files.WriteFile(path, []byte(defaultConfigYAML), 0o644); err != nil {
		return false, fmt.Errorf("cannot write config %q: %w (expected a writable file path)", path, err)
	}
	return true, nil
}

func configExists(path string, files storycodeFS) (bool, error) {
	_, err := files.Stat(path)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, fmt.Errorf("cannot stat config %q: %w (expected a readable file path)", path, err)
}

func reportInit(out io.Writer, storyDir string, wroteConfig bool) {
	if wroteConfig {
		fmt.Fprintf(out, "Initialized StoryCode in %s\n", storyDir)
		return
	}
	fmt.Fprintf(out, "StoryCode already initialized in %s\n", storyDir)
}
