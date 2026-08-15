package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

type usageError struct {
	msg string
}

func (e usageError) Error() string {
	return e.msg
}

const (
	ExitOK    = 0
	ExitError = 1
	ExitUsage = 2
)

type BuildInfo struct {
	Version string
	Commit  string
}

// Run executes the StoryCode CLI with the given arguments.
// It never executes files from an analyzed repository.
//
//	code := Run([]string{"status", "--help"}, os.Stdout, os.Stderr, BuildInfo{Version: "dev", Commit: "unknown"})
func Run(args []string, stdout, stderr io.Writer, info BuildInfo) int {
	cmd := newRoot(info)
	if args == nil {
		args = []string{}
	}
	cmd.SetArgs(args)
	cmd.SetOut(stdout)
	cmd.SetErr(stderr)
	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(stderr, err)
		return exitCode(err)
	}
	return ExitOK
}

func exitCode(err error) int {
	var usage usageError
	if errors.As(err, &usage) {
		return ExitUsage
	}
	msg := err.Error()
	if strings.Contains(msg, "unknown command") ||
		strings.Contains(msg, "unknown flag") ||
		strings.Contains(msg, "arg(s)") {
		return ExitUsage
	}
	return ExitError
}

func newRoot(info BuildInfo) *cobra.Command {
	root := &cobra.Command{
		Use:           "storycode",
		Short:         "Build local-first stories from a Python codebase",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.CompletionOptions.DisableDefaultCmd = true
	setVersion(root, info)
	addCommands(root)
	return root
}

func setVersion(root *cobra.Command, info BuildInfo) {
	root.Version = info.Version
	root.SetVersionTemplate(
		fmt.Sprintf("{{.Name}} version {{.Version}} (commit %s)\n", info.Commit),
	)
}
