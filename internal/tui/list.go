package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"todo_cli/internal/model"
	"todo_cli/internal/storage"
)

// ListView displays a list of todos
type ListView struct {
	store        storage.Storage
	todos        []model.Todo
	cursor       int
	selected     map[int64]bool
	filter       storage.Filter
	width        int
	height       int
	searchMode   bool
	searchInput  textinput.Model
	allTags      []string
	selectedTags map[string]bool
	tagCursor    int
	showPending  bool
}

// NewListView creates a new list view
func NewListView(store storage.Storage) *ListView {
	ti := textinput.New()
	ti.Placeholder = "Search..."
	ti.CharLimit = 100
	ti.Width = 30

	pending := false
	return &ListView{
		store:        store,
		selected:     make(map[int64]bool),
		selectedTags: make(map[string]bool),
		searchInput:  ti,
		showPending:  true,
		filter: storage.Filter{
			Completed: &pending,
		},
	}
}

// SetSize sets the view dimensions
func (l *ListView) SetSize(width, height int) {
	l.width = width
	l.height = height
}

func (l *ListView) loadTodos() tea.Cmd {
	return func() tea.Msg {
		todos, err := l.store.List(l.filter)
		if err != nil {
			return errMsg{err}
		}
		l.todos = todos
		if l.cursor >= len(todos) {
			l.cursor = max(0, len(todos)-1)
		}
		return todosLoadedMsg{}
	}
}

func (l *ListView) loadTags() tea.Cmd {
	return func() tea.Msg {
		tags, err := l.store.GetAllTags()
		if err != nil {
			return errMsg{err}
		}
		l.allTags = tags
		return nil
	}
}

// SelectedTodo returns the currently highlighted todo
func (l *ListView) SelectedTodo() *model.Todo {
	if l.cursor < 0 || l.cursor >= len(l.todos) {
		return nil
	}
	return &l.todos[l.cursor]
}

func (l *ListView) toggleComplete() tea.Cmd {
	todo := l.SelectedTodo()
	if todo == nil {
		return nil
	}

	return func() tea.Msg {
		todo.Completed = !todo.Completed
		if todo.Completed {
			now := time.Now().UTC()
			todo.CompletedAt = &now
		} else {
			todo.CompletedAt = nil
		}

		if err := l.store.Update(todo); err != nil {
			return errMsg{err}
		}
		return todoUpdatedMsg{}
	}
}

func (l *ListView) toggleSelect() {
	todo := l.SelectedTodo()
	if todo == nil {
		return
	}

	if l.selected[todo.ID] {
		delete(l.selected, todo.ID)
	} else {
		l.selected[todo.ID] = true
	}
}

func (l *ListView) deleteSelected() tea.Cmd {
	if len(l.selected) == 0 {
		// Delete current item if nothing selected
		todo := l.SelectedTodo()
		if todo == nil {
			return nil
		}
		return func() tea.Msg {
			if err := l.store.Delete(todo.ID); err != nil {
				return errMsg{err}
			}
			return todoDeletedMsg{}
		}
	}

	// Delete all selected
	return func() tea.Msg {
		for id := range l.selected {
			if err := l.store.Delete(id); err != nil {
				return errMsg{err}
			}
		}
		l.selected = make(map[int64]bool)
		return todoDeletedMsg{}
	}
}

func (l *ListView) cyclePriority() tea.Cmd {
	todo := l.SelectedTodo()
	if todo == nil {
		return nil
	}

	return func() tea.Msg {
		todo.Priority = (todo.Priority % 5) + 1
		if err := l.store.Update(todo); err != nil {
			return errMsg{err}
		}
		return todoUpdatedMsg{}
	}
}

func (l *ListView) toggleTagSelection() {
	if l.tagCursor == 0 {
		// "All tags" option - clear selection
		l.selectedTags = make(map[string]bool)
		return
	}

	idx := l.tagCursor - 1
	if idx < len(l.allTags) {
		tag := l.allTags[idx]
		if l.selectedTags[tag] {
			delete(l.selectedTags, tag)
		} else {
			l.selectedTags[tag] = true
		}
	}
}

func (l *ListView) applyTagFilter() {
	tags := make([]string, 0, len(l.selectedTags))
	for tag := range l.selectedTags {
		tags = append(tags, tag)
	}
	l.filter.Tags = tags
}

// Update handles input for the list view
func (l *ListView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "j", "down":
			l.cursor++
			if l.cursor >= len(l.todos) {
				l.cursor = 0
			}

		case "k", "up":
			l.cursor--
			if l.cursor < 0 {
				l.cursor = max(0, len(l.todos)-1)
			}

		case "g":
			l.cursor = 0

		case "G":
			l.cursor = max(0, len(l.todos)-1)

		case "tab":
			// Toggle between pending/all
			if l.showPending {
				l.filter.Completed = nil
				l.showPending = false
			} else {
				pending := false
				l.filter.Completed = &pending
				l.showPending = true
			}
			return l.loadTodos()
		}
	}

	return nil
}

// UpdateSearch handles search input
func (l *ListView) UpdateSearch(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	l.searchInput, cmd = l.searchInput.Update(msg)
	return cmd
}

// View renders the list view
func (l *ListView) View() string {
	var b strings.Builder

	// Title
	title := "TODO List"
	if l.filter.Search != "" {
		title += fmt.Sprintf(" (search: %s)", l.filter.Search)
	}
	if len(l.filter.Tags) > 0 {
		title += fmt.Sprintf(" (tags: %s)", strings.Join(l.filter.Tags, ", "))
	}
	if !l.showPending {
		title += " [ALL]"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	if len(l.todos) == 0 {
		b.WriteString(helpStyle.Render("No todos found. Press 'n' to add one."))
	} else {
		// Calculate visible items
		visibleHeight := l.height - 6
		if visibleHeight < 1 {
			visibleHeight = 10
		}

		start := 0
		if l.cursor >= visibleHeight {
			start = l.cursor - visibleHeight + 1
		}

		end := start + visibleHeight
		if end > len(l.todos) {
			end = len(l.todos)
		}

		for i := start; i < end; i++ {
			todo := l.todos[i]
			b.WriteString(l.renderTodoItem(todo, i == l.cursor))
			b.WriteString("\n")
		}
	}

	// Status bar
	b.WriteString("\n")
	statusText := fmt.Sprintf(" %d todos ", len(l.todos))
	if len(l.selected) > 0 {
		statusText += fmt.Sprintf("| %d selected ", len(l.selected))
	}
	b.WriteString(statusBarStyle.Render(statusText))

	// Help
	b.WriteString("\n")
	help := "j/k:navigate  space:toggle  n:new  e:edit  /:search  t:tags  D:delete  tab:all/pending  q:quit"
	b.WriteString(helpStyle.Render(help))

	return b.String()
}

func (l *ListView) renderTodoItem(todo model.Todo, selected bool) string {
	var parts []string

	// Selection indicator
	if l.selected[todo.ID] {
		parts = append(parts, selectedIndicator)
	} else if selected {
		parts = append(parts, lipgloss.NewStyle().Foreground(primaryColor).Render("> "))
	} else {
		parts = append(parts, unselectedIndicator)
	}

	// Checkbox
	if todo.Completed {
		parts = append(parts, checkedBox)
	} else {
		parts = append(parts, uncheckedBox)
	}

	// Priority
	parts = append(parts, " "+priorityIndicator(todo.Priority))

	// Title
	title := todo.Title
	if len(title) > 40 {
		title = title[:37] + "..."
	}

	var titleStyle lipgloss.Style
	if todo.Completed {
		titleStyle = completedItemStyle
	} else if selected {
		titleStyle = selectedItemStyle
	} else {
		titleStyle = normalItemStyle
	}
	parts = append(parts, " "+titleStyle.Render(title))

	// Due date
	if todo.DueDate != nil {
		dueStr := formatDue(todo)
		parts = append(parts, " "+dueStr)
	}

	// Tags (show first 2)
	if len(todo.Tags) > 0 {
		tagsToShow := todo.Tags
		if len(tagsToShow) > 2 {
			tagsToShow = tagsToShow[:2]
		}
		for _, tag := range tagsToShow {
			parts = append(parts, " "+tagStyle.Render(tag))
		}
	}

	return strings.Join(parts, "")
}

func formatDue(todo model.Todo) string {
	if todo.DueDate == nil {
		return ""
	}

	if todo.IsOverdue() && !todo.Completed {
		return overdueStyle.Render(todo.DueDate.Format("Jan 2") + "!")
	}
	if todo.IsDueToday() {
		return dueTodayStyle.Render("today")
	}
	if todo.IsDueTomorrow() {
		return dueNormalStyle.Render("tomorrow")
	}
	return dueNormalStyle.Render(todo.DueDate.Format("Jan 2"))
}

// ViewSearch renders the search mode
func (l *ListView) ViewSearch() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Search Todos"))
	b.WriteString("\n\n")

	b.WriteString(focusedInputStyle.Render(l.searchInput.View()))
	b.WriteString("\n\n")

	b.WriteString(helpStyle.Render("Enter: search  Esc: cancel"))

	return b.String()
}

// ViewTagFilter renders the tag filter view
func (l *ListView) ViewTagFilter() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("Filter by Tag"))
	b.WriteString("\n\n")

	// "All" option
	indicator := "  "
	if l.tagCursor == 0 {
		indicator = "> "
	}
	allSelected := len(l.selectedTags) == 0
	checkbox := uncheckedBox
	if allSelected {
		checkbox = checkedBox
	}
	b.WriteString(fmt.Sprintf("%s%s All tags\n", indicator, checkbox))

	// Individual tags
	for i, tag := range l.allTags {
		indicator := "  "
		if l.tagCursor == i+1 {
			indicator = "> "
		}
		checkbox := uncheckedBox
		if l.selectedTags[tag] {
			checkbox = checkedBox
		}
		b.WriteString(fmt.Sprintf("%s%s %s\n", indicator, checkbox, tagStyle.Render(tag)))
	}

	b.WriteString("\n")
	b.WriteString(helpStyle.Render("j/k: navigate  space: toggle  enter: apply  esc: cancel"))

	return b.String()
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
