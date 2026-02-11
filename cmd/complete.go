package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

var completeCmd = &cobra.Command{
	Use:   "complete <id>",
	Short: "Mark a todo as complete",
	Long: `Mark a todo as complete.

Examples:
  todo complete 1
  todo complete 42`,
	Aliases: []string{"done"},
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

		if todo.Completed {
			fmt.Printf("Todo #%d is already completed.\n", id)
			return nil
		}

		now := time.Now().UTC()
		todo.Completed = true
		todo.CompletedAt = &now

		if err := store.Update(todo); err != nil {
			return fmt.Errorf("failed to complete todo: %w", err)
		}

		fmt.Printf("Completed todo #%d: %s\n", todo.ID, todo.Title)
		return nil
	},
}
