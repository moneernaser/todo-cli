package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
)

var reopenCmd = &cobra.Command{
	Use:   "reopen <id>",
	Short: "Reopen a completed todo",
	Long: `Mark a completed todo as pending again.

Examples:
  todo reopen 1
  todo reopen 42`,
	Aliases: []string{"undo"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid ID: %s", args[0])
		}

		todo, err := store.GetByID(id)
		if err != nil {
			return err
		}

		if !todo.Completed {
			fmt.Printf("Todo #%d is not completed.\n", id)
			return nil
		}

		todo.Completed = false
		todo.CompletedAt = nil

		if err := store.Update(todo); err != nil {
			return fmt.Errorf("failed to reopen todo: %w", err)
		}

		fmt.Printf("Reopened todo #%d: %s\n", todo.ID, todo.Title)
		return nil
	},
}
