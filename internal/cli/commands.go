package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func addCommands(root *cobra.Command) {
	root.AddCommand(initCommand())
	root.AddCommand(statusCommand())
	root.AddCommand(indexCommand())
	root.AddCommand(stubCommand("discover", "Discover stories from indexed entry points"))
	root.AddCommand(stubCommand("serve", "Serve the local StoryCode UI"))
	root.AddCommand(storyCommand())
	root.AddCommand(stubCommand("verify", "Verify stories against the current index"))
}

func stubCommand(use, short string) *cobra.Command {
	return &cobra.Command{
		Use:   use,
		Short: short,
		Args:  cobra.NoArgs,
		RunE:  stubRun,
	}
}

func stubRun(_ *cobra.Command, _ []string) error {
	return nil
}

func storyCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "story",
		Short: "Inspect local stories",
	}
	cmd.AddCommand(stubCommand("list", "List discovered stories"))
	cmd.AddCommand(storyShowCommand())
	return cmd
}

func storyShowCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "show <key>",
		Short: "Show one story by key",
		Args:  storyShowArgs,
		RunE:  stubRun,
	}
}

func storyShowArgs(_ *cobra.Command, args []string) error {
	if len(args) == 1 {
		return nil
	}
	return usageError{msg: fmt.Sprintf(
		"story show requires a story key, got %d arguments %v, expected shape: storycode story show <key>",
		len(args),
		args,
	)}
}
