package cmd

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var showCmd = &cobra.Command{
	Use:   "show <id>",
	Short: "Show todo details",
	Long: `Show detailed information about a specific todo.

Examples:
  todo show 1
  todo show 42`,
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

		// Print todo details
		fmt.Printf("Todo #%d\n", todo.ID)
		fmt.Println(strings.Repeat("=", 40))

		fmt.Printf("Title:       %s\n", todo.Title)

		if todo.Description != "" {
			fmt.Printf("Description: %s\n", todo.Description)
		}

		status := "Pending"
		if todo.Completed {
			status = "Completed"
		}
		fmt.Printf("Status:      %s\n", status)

		if todo.Priority > 0 {
			fmt.Printf("Priority:    %d (%s)\n", todo.Priority, todo.PriorityString())
		}

		if len(todo.Tags) > 0 {
			fmt.Printf("Tags:        %s\n", strings.Join(todo.Tags, " "))
		}

		if todo.DueDate != nil {
			dueStr := todo.DueDate.Local().Format("2006-01-02 15:04")
			if todo.IsOverdue() && !todo.Completed {
				dueStr += " (OVERDUE)"
			} else if todo.IsDueToday() {
				dueStr += " (today)"
			} else if todo.IsDueTomorrow() {
				dueStr += " (tomorrow)"
			}
			fmt.Printf("Due:         %s\n", dueStr)
		}

		fmt.Printf("Created:     %s\n", todo.CreatedAt.Local().Format("2006-01-02 15:04"))
		fmt.Printf("Updated:     %s\n", todo.UpdatedAt.Local().Format("2006-01-02 15:04"))

		if todo.CompletedAt != nil {
			fmt.Printf("Completed:   %s\n", todo.CompletedAt.Local().Format("2006-01-02 15:04"))
		}

		return nil
	},
}
