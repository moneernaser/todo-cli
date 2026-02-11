package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"todo_cli/internal/model"
	"todo_cli/internal/storage"
)

var (
	addTags        string
	addDue         string
	addPriority    int
	addDescription string
)

var addCmd = &cobra.Command{
	Use:   "add <title>",
	Short: "Add a new todo",
	Long: `Add a new todo item with optional tags, due date, and priority.

Examples:
  todo add "Buy groceries"
  todo add "Finish report" --tags "#work #urgent"
  todo add "Call mom" --due 2026-02-14
  todo add "Important task" --priority 1
  todo add "Project task" --tags "#work" --due tomorrow --priority 2`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		title := args[0]

		todo := &model.Todo{
			Title:       title,
			Description: addDescription,
			Tags:        storage.ParseTags(addTags),
			Priority:    addPriority,
		}

		if addDue != "" {
			dueDate, err := storage.ParseDueDate(addDue)
			if err != nil {
				return fmt.Errorf("invalid due date: %w", err)
			}
			todo.DueDate = dueDate
		}

		if addPriority < 0 || addPriority > 5 {
			return fmt.Errorf("priority must be between 0 and 5 (1=highest, 5=lowest, 0=none)")
		}

		if err := store.Create(todo); err != nil {
			return fmt.Errorf("failed to create todo: %w", err)
		}

		fmt.Printf("Created todo #%d: %s\n", todo.ID, todo.Title)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVarP(&addTags, "tags", "t", "", "Tags (e.g., '#work #urgent')")
	addCmd.Flags().StringVarP(&addDue, "due", "d", "", "Due date (e.g., '2026-02-14', 'today', 'tomorrow')")
	addCmd.Flags().IntVarP(&addPriority, "priority", "p", 0, "Priority (1=highest, 5=lowest, 0=none)")
	addCmd.Flags().StringVar(&addDescription, "desc", "", "Description")
}
