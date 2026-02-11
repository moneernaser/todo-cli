package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"todo_cli/internal/model"
	"todo_cli/internal/storage"
)

var (
	listFilterTag string
	listDue       string
	listOverdue   bool
	listCompleted bool
	listPending   bool
	listAll       bool
	listSort      string
	listSortOrder string
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List todos",
	Long: `List todos with optional filters.

Examples:
  todo list                    # List pending todos
  todo list --all              # List all todos
  todo list --completed        # List completed todos
  todo list --filter-tag #work # Filter by tag
  todo list --due today        # Due today
  todo list --due tomorrow     # Due tomorrow
  todo list --due next-week    # Due within 7 days
  todo list --overdue          # Past due date
  todo list --sort priority    # Sort by priority`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		filter := storage.Filter{
			SortOrder: storage.SortDesc,
		}

		// Completed filter
		if listCompleted {
			completed := true
			filter.Completed = &completed
		} else if listPending || !listAll {
			completed := false
			filter.Completed = &completed
		}

		// Tag filter
		if listFilterTag != "" {
			filter.Tags = storage.ParseTags(listFilterTag)
		}

		// Due date filter
		if listOverdue {
			filter.DueDate = &storage.DueDateFilter{Type: storage.DueOverdue}
		} else if listDue != "" {
			switch strings.ToLower(listDue) {
			case "today":
				filter.DueDate = &storage.DueDateFilter{Type: storage.DueToday}
			case "tomorrow":
				filter.DueDate = &storage.DueDateFilter{Type: storage.DueTomorrow}
			case "next-week", "nextweek":
				filter.DueDate = &storage.DueDateFilter{Type: storage.DueNextWeek}
			default:
				dueDate, err := storage.ParseDueDate(listDue)
				if err != nil {
					return fmt.Errorf("invalid due date filter: %w", err)
				}
				filter.DueDate = &storage.DueDateFilter{
					Type:         storage.DueSpecific,
					SpecificDate: dueDate,
				}
			}
		}

		// Sort
		switch strings.ToLower(listSort) {
		case "priority", "p":
			filter.SortBy = storage.SortByPriority
			filter.SortOrder = storage.SortAsc // 1 (highest) first
		case "due", "d":
			filter.SortBy = storage.SortByDueDate
			filter.SortOrder = storage.SortAsc
		case "created", "c":
			filter.SortBy = storage.SortByCreated
		case "updated", "u":
			filter.SortBy = storage.SortByUpdated
		case "title", "t":
			filter.SortBy = storage.SortByTitle
			filter.SortOrder = storage.SortAsc
		}

		// Override sort order if specified
		if listSortOrder != "" {
			switch strings.ToLower(listSortOrder) {
			case "asc", "a":
				filter.SortOrder = storage.SortAsc
			case "desc", "d":
				filter.SortOrder = storage.SortDesc
			}
		}

		todos, err := store.List(filter)
		if err != nil {
			return fmt.Errorf("failed to list todos: %w", err)
		}

		if len(todos) == 0 {
			fmt.Println("No todos found.")
			return nil
		}

		printTodoList(todos)
		return nil
	},
}

func printTodoList(todos []model.Todo) {
	// Print header
	fmt.Printf("%-4s %-1s %-40s %-12s %-10s %s\n", "ID", "P", "Title", "Due", "Tags", "Status")
	fmt.Println(strings.Repeat("-", 85))

	for _, todo := range todos {
		status := "[ ]"
		if todo.Completed {
			status = "[x]"
		}

		priority := " "
		if todo.Priority > 0 {
			priority = fmt.Sprintf("%d", todo.Priority)
		}

		title := todo.Title
		if len(title) > 40 {
			title = title[:37] + "..."
		}

		dueStr := ""
		if todo.DueDate != nil {
			if todo.IsOverdue() && !todo.Completed {
				dueStr = todo.DueDate.Format("2006-01-02") + "!"
			} else if todo.IsDueToday() {
				dueStr = "today"
			} else if todo.IsDueTomorrow() {
				dueStr = "tomorrow"
			} else {
				dueStr = todo.DueDate.Format("2006-01-02")
			}
		}

		tagsStr := ""
		if len(todo.Tags) > 0 {
			tagsStr = strings.Join(todo.Tags, " ")
			if len(tagsStr) > 10 {
				tagsStr = tagsStr[:7] + "..."
			}
		}

		fmt.Printf("%-4d %-1s %-40s %-12s %-10s %s\n",
			todo.ID, priority, title, dueStr, tagsStr, status)
	}

	fmt.Printf("\nTotal: %d todo(s)\n", len(todos))
}

func formatDueDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02")
}

func init() {
	listCmd.Flags().StringVar(&listFilterTag, "filter-tag", "", "Filter by tag (e.g., '#work')")
	listCmd.Flags().StringVar(&listDue, "due", "", "Filter by due date (today, tomorrow, next-week, or date)")
	listCmd.Flags().BoolVar(&listOverdue, "overdue", false, "Show overdue todos")
	listCmd.Flags().BoolVar(&listCompleted, "completed", false, "Show completed todos")
	listCmd.Flags().BoolVar(&listPending, "pending", false, "Show pending todos (default)")
	listCmd.Flags().BoolVar(&listAll, "all", false, "Show all todos")
	listCmd.Flags().StringVar(&listSort, "sort", "", "Sort by: priority, due, created, updated, title")
	listCmd.Flags().StringVar(&listSortOrder, "order", "", "Sort order: asc, desc")
}
