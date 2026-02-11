package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"todo_cli/internal/storage"
)

var (
	store   storage.Storage
	rootCmd = &cobra.Command{
		Use:   "todo",
		Short: "A command-line TODO application",
		Long: `A command-line TODO application with both CLI and interactive TUI modes.

Manage your tasks with tags, due dates, and priorities.
Use 'todo tui' for an interactive terminal interface.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// Skip storage initialization for completion commands
			if cmd.Name() == "completion" || cmd.Parent() != nil && cmd.Parent().Name() == "completion" {
				return nil
			}

			var err error
			store, err = storage.NewSQLiteStorage()
			if err != nil {
				return fmt.Errorf("failed to initialize storage: %w", err)
			}
			return nil
		},
		PersistentPostRun: func(cmd *cobra.Command, args []string) {
			if store != nil {
				store.Close()
			}
		},
	}
)

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(addCmd)
	rootCmd.AddCommand(listCmd)
	rootCmd.AddCommand(showCmd)
	rootCmd.AddCommand(completeCmd)
	rootCmd.AddCommand(reopenCmd)
	rootCmd.AddCommand(editCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(tuiCmd)
	rootCmd.AddCommand(exportCmd)
	rootCmd.AddCommand(importCmd)
}
