package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	"todo_cli/internal/storage"
)

var (
	editTitle       string
	editTags        string
	editDue         string
	editPriority    int
	editDescription string
	editClearDue    bool
	editClearTags   bool
)

var editCmd = &cobra.Command{
	Use:   "edit <id>",
	Short: "Edit a todo",
	Long: `Edit an existing todo's title, tags, due date, priority, or description.

Examples:
  todo edit 1 --title "New title"
  todo edit 1 --tags "#work #updated"
  todo edit 1 --due 2026-02-20
  todo edit 1 --priority 2
  todo edit 1 --clear-due
  todo edit 1 --clear-tags`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("invalid ID: %s", args[0])
		}

		todo, err := store.GetByID(id)
		if err != nil {
			return err
		}

		modified := false

		if editTitle != "" {
			todo.Title = editTitle
			modified = true
		}

		if editDescription != "" {
			todo.Description = editDescription
			modified = true
		}

		if editTags != "" {
			todo.Tags = storage.ParseTags(editTags)
			modified = true
		}

		if editClearTags {
			todo.Tags = nil
			modified = true
		}

		if editDue != "" {
			dueDate, err := storage.ParseDueDate(editDue)
			if err != nil {
				return fmt.Errorf("invalid due date: %w", err)
			}
			todo.DueDate = dueDate
			modified = true
		}

		if editClearDue {
			todo.DueDate = nil
			modified = true
		}

		if cmd.Flags().Changed("priority") {
			if editPriority < 0 || editPriority > 5 {
				return fmt.Errorf("priority must be between 0 and 5 (1=highest, 5=lowest, 0=none)")
			}
			todo.Priority = editPriority
			modified = true
		}

		if !modified {
			fmt.Println("No changes specified. Use --title, --tags, --due, --priority, --desc, --clear-due, or --clear-tags.")
			return nil
		}

		if err := store.Update(todo); err != nil {
			return fmt.Errorf("failed to update todo: %w", err)
		}

		fmt.Printf("Updated todo #%d: %s\n", todo.ID, todo.Title)
		return nil
	},
}

func init() {
	editCmd.Flags().StringVar(&editTitle, "title", "", "New title")
	editCmd.Flags().StringVarP(&editTags, "tags", "t", "", "New tags (replaces existing)")
	editCmd.Flags().StringVarP(&editDue, "due", "d", "", "New due date")
	editCmd.Flags().IntVarP(&editPriority, "priority", "p", 0, "New priority (1-5, 0=none)")
	editCmd.Flags().StringVar(&editDescription, "desc", "", "New description")
	editCmd.Flags().BoolVar(&editClearDue, "clear-due", false, "Clear the due date")
	editCmd.Flags().BoolVar(&editClearTags, "clear-tags", false, "Clear all tags")
}
