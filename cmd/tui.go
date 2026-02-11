package cmd

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/spf13/cobra"

	"todo_cli/internal/tui"
)

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive TUI mode",
	Long: `Launch an interactive terminal UI for managing todos.

Keyboard shortcuts:
  j/↓       Move down
  k/↑       Move up
  Space     Toggle complete
  Enter     View details
  /         Search
  t         Filter by tag
  p         Set priority
  e         Edit todo
  n         New todo
  x         Toggle select
  D         Delete selected
  q/Esc     Quit / Back`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app := tui.NewApp(store)
		p := tea.NewProgram(app, tea.WithAltScreen())

		if _, err := p.Run(); err != nil {
			return fmt.Errorf("TUI error: %w", err)
		}

		return nil
	},
}
