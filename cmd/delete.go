package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var deleteYes bool

var deleteCmd = &cobra.Command{
	Use:   "delete <id>",
	Short: "Delete a todo",
	Long: `Delete a todo by ID. Requires confirmation unless --yes is specified.

Examples:
  todo delete 1
  todo delete 1 --yes`,
	Aliases: []string{"rm", "remove"},
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

		if !deleteYes {
			fmt.Printf("Delete todo #%d: %s? [y/N] ", todo.ID, todo.Title)
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read response: %w", err)
			}

			response = strings.TrimSpace(strings.ToLower(response))
			if response != "y" && response != "yes" {
				fmt.Println("Cancelled.")
				return nil
			}
		}

		if err := store.Delete(id); err != nil {
			return fmt.Errorf("failed to delete todo: %w", err)
		}

		fmt.Printf("Deleted todo #%d: %s\n", todo.ID, todo.Title)
		return nil
	},
}

func init() {
	deleteCmd.Flags().BoolVarP(&deleteYes, "yes", "y", false, "Skip confirmation")
}
