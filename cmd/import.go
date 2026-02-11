package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"todo_cli/internal/model"
)

var importCmd = &cobra.Command{
	Use:   "import <filename>",
	Short: "Import todos from JSON",
	Long: `Import todos from a JSON file. Each todo will be created as a new entry.

Examples:
  todo import todos.json
  todo import backup.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		data, err := os.ReadFile(filename)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		var todos []model.Todo
		if err := json.Unmarshal(data, &todos); err != nil {
			return fmt.Errorf("failed to parse JSON: %w", err)
		}

		imported := 0
		for _, todo := range todos {
			// Create as new todo (ID will be assigned by storage)
			newTodo := &model.Todo{
				Title:       todo.Title,
				Description: todo.Description,
				Tags:        todo.Tags,
				DueDate:     todo.DueDate,
				Completed:   todo.Completed,
				CompletedAt: todo.CompletedAt,
				Priority:    todo.Priority,
			}

			if err := store.Create(newTodo); err != nil {
				fmt.Printf("Warning: failed to import todo '%s': %v\n", todo.Title, err)
				continue
			}
			imported++
		}

		fmt.Printf("Imported %d todo(s) from %s\n", imported, filename)
		return nil
	},
}
