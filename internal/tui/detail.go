package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"todo_cli/internal/model"
	"todo_cli/internal/storage"
)

// DetailView displays a single todo's details
type DetailView struct {
	todo   *model.Todo
	width  int
	height int
}

// NewDetailView creates a new detail view
func NewDetailView() *DetailView {
	return &DetailView{}
}

// SetTodo sets the todo to display
func (d *DetailView) SetTodo(todo *model.Todo) {
	d.todo = todo
}

// Update handles input for the detail view
func (d *DetailView) Update(msg tea.Msg) tea.Cmd {
	return nil
}

func (d *DetailView) toggleComplete(store storage.Storage) tea.Cmd {
	if d.todo == nil {
		return nil
	}

	return func() tea.Msg {
		d.todo.Completed = !d.todo.Completed
		if d.todo.Completed {
			now := time.Now().UTC()
			d.todo.CompletedAt = &now
		} else {
			d.todo.CompletedAt = nil
		}

		if err := store.Update(d.todo); err != nil {
			return errMsg{err}
		}
		return todoUpdatedMsg{}
	}
}

// View renders the detail view
func (d *DetailView) View() string {
	if d.todo == nil {
		return "No todo selected"
	}

	var b strings.Builder

	// Title bar
	title := fmt.Sprintf("Todo #%d", d.todo.ID)
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Title
	b.WriteString(labelStyle.Render("Title:"))
	b.WriteString(valueStyle.Render(d.todo.Title))
	b.WriteString("\n")

	// Description
	if d.todo.Description != "" {
		b.WriteString(labelStyle.Render("Description:"))
		b.WriteString(valueStyle.Render(d.todo.Description))
		b.WriteString("\n")
	}

	// Status
	b.WriteString(labelStyle.Render("Status:"))
	if d.todo.Completed {
		b.WriteString(successStyle.Render("Completed"))
	} else {
		b.WriteString(valueStyle.Render("Pending"))
	}
	b.WriteString("\n")

	// Priority
	if d.todo.Priority > 0 {
		b.WriteString(labelStyle.Render("Priority:"))
		style := priorityStyle[d.todo.Priority]
		b.WriteString(style.Render(fmt.Sprintf("%d (%s)", d.todo.Priority, d.todo.PriorityString())))
		b.WriteString("\n")
	}

	// Tags
	if len(d.todo.Tags) > 0 {
		b.WriteString(labelStyle.Render("Tags:"))
		for _, tag := range d.todo.Tags {
			b.WriteString(tagStyle.Render(tag) + " ")
		}
		b.WriteString("\n")
	}

	// Due date
	if d.todo.DueDate != nil {
		b.WriteString(labelStyle.Render("Due:"))
		dueStr := d.todo.DueDate.Local().Format("2006-01-02 15:04")
		if d.todo.IsOverdue() && !d.todo.Completed {
			b.WriteString(overdueStyle.Render(dueStr + " (OVERDUE)"))
		} else if d.todo.IsDueToday() {
			b.WriteString(dueTodayStyle.Render(dueStr + " (today)"))
		} else if d.todo.IsDueTomorrow() {
			b.WriteString(dueNormalStyle.Render(dueStr + " (tomorrow)"))
		} else {
			b.WriteString(dueNormalStyle.Render(dueStr))
		}
		b.WriteString("\n")
	}

	// Timestamps
	b.WriteString("\n")
	b.WriteString(labelStyle.Render("Created:"))
	b.WriteString(valueStyle.Render(d.todo.CreatedAt.Local().Format("2006-01-02 15:04")))
	b.WriteString("\n")

	b.WriteString(labelStyle.Render("Updated:"))
	b.WriteString(valueStyle.Render(d.todo.UpdatedAt.Local().Format("2006-01-02 15:04")))
	b.WriteString("\n")

	if d.todo.CompletedAt != nil {
		b.WriteString(labelStyle.Render("Completed:"))
		b.WriteString(valueStyle.Render(d.todo.CompletedAt.Local().Format("2006-01-02 15:04")))
		b.WriteString("\n")
	}

	// Help
	b.WriteString("\n")
	b.WriteString(helpStyle.Render("space: toggle complete  e: edit  q/esc: back"))

	return b.String()
}
