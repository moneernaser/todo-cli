package tui

import (
	"github.com/charmbracelet/lipgloss"
)

var (
	// Colors
	primaryColor   = lipgloss.Color("205")
	secondaryColor = lipgloss.Color("240")
	successColor   = lipgloss.Color("82")
	warningColor   = lipgloss.Color("214")
	errorColor     = lipgloss.Color("196")
	mutedColor     = lipgloss.Color("241")

	// App styles
	appStyle = lipgloss.NewStyle().Padding(1, 2)

	// Title bar
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("15")).
			Background(primaryColor).
			Padding(0, 1)

	// Status bar
	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("15")).
			Background(secondaryColor).
			Padding(0, 1)

	// List item styles
	selectedItemStyle = lipgloss.NewStyle().
				Foreground(primaryColor).
				Bold(true)

	normalItemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	completedItemStyle = lipgloss.NewStyle().
				Foreground(mutedColor).
				Strikethrough(true)

	// Priority styles
	priorityStyle = map[int]lipgloss.Style{
		1: lipgloss.NewStyle().Foreground(errorColor).Bold(true),   // Urgent
		2: lipgloss.NewStyle().Foreground(warningColor).Bold(true), // High
		3: lipgloss.NewStyle().Foreground(lipgloss.Color("226")),   // Medium
		4: lipgloss.NewStyle().Foreground(lipgloss.Color("117")),   // Low
		5: lipgloss.NewStyle().Foreground(mutedColor),              // Lowest
	}

	// Tag style
	tagStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("117")).
			Background(lipgloss.Color("236")).
			Padding(0, 1)

	// Due date styles
	dueTodayStyle = lipgloss.NewStyle().
			Foreground(warningColor).
			Bold(true)

	overdueStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	dueNormalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Detail view styles
	labelStyle = lipgloss.NewStyle().
			Foreground(secondaryColor).
			Width(12)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("252"))

	// Input styles
	inputLabelStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true)

	inputStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(secondaryColor).
			Padding(0, 1)

	focusedInputStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(primaryColor).
				Padding(0, 1)

	// Help text
	helpStyle = lipgloss.NewStyle().
			Foreground(mutedColor)

	// Error style
	errorStyle = lipgloss.NewStyle().
			Foreground(errorColor).
			Bold(true)

	// Success style
	successStyle = lipgloss.NewStyle().
			Foreground(successColor)

	// Selection indicator
	selectedIndicator   = lipgloss.NewStyle().Foreground(primaryColor).Render("● ")
	unselectedIndicator = "  "

	// Checkbox
	checkedBox   = lipgloss.NewStyle().Foreground(successColor).Render("[✓]")
	uncheckedBox = lipgloss.NewStyle().Foreground(secondaryColor).Render("[ ]")
)

func priorityIndicator(priority int) string {
	if priority == 0 {
		return " "
	}
	style, ok := priorityStyle[priority]
	if !ok {
		return " "
	}
	return style.Render("●")
}
