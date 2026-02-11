package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"todo_cli/internal/storage"
)

var exportCmd = &cobra.Command{
	Use:   "export <filename>",
	Short: "Export todos to JSON",
	Long: `Export all todos to a JSON file.

Examples:
  todo export todos.json
  todo export backup.json`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		filename := args[0]

		// Get all todos (no filter)
		todos, err := store.List(storage.Filter{})
		if err != nil {
			return fmt.Errorf("failed to list todos: %w", err)
		}

		data, err := json.MarshalIndent(todos, "", "  ")
		if err != nil {
			return fmt.Errorf("failed to marshal todos: %w", err)
		}

		if err := os.WriteFile(filename, data, 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}

		fmt.Printf("Exported %d todo(s) to %s\n", len(todos), filename)
		return nil
	},
}
