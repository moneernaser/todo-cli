package tui

import (
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"

	"todo_cli/internal/storage"
)

// View represents the current view state
type View int

const (
	ViewList View = iota
	ViewDetail
	ViewAdd
	ViewEdit
	ViewSearch
	ViewTagFilter
)

// App is the main TUI application model
type App struct {
	store    storage.Storage
	view     View
	list     *ListView
	detail   *DetailView
	input    *InputView
	width    int
	height   int
	err      error
	quitting bool
}

// Message types
type errMsg struct{ err error }
type todosLoadedMsg struct{}
type todoCreatedMsg struct{}
type todoUpdatedMsg struct{}
type todoDeletedMsg struct{}

// NewApp creates a new TUI application
func NewApp(store storage.Storage) *App {
	app := &App{
		store: store,
		view:  ViewList,
	}
	app.list = NewListView(store)
	app.detail = NewDetailView()
	app.input = NewInputView()
	return app
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return a.list.loadTodos()
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.list.SetSize(msg.Width, msg.Height-4) // Account for header/footer
		return a, nil

	case tea.KeyMsg:
		// Global quit
		if key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c"))) {
			a.quitting = true
			return a, tea.Quit
		}

	case errMsg:
		a.err = msg.err
		return a, nil

	case todosLoadedMsg:
		return a, nil

	case todoCreatedMsg, todoUpdatedMsg, todoDeletedMsg:
		// Reload list after modifications
		return a, a.list.loadTodos()
	}

	// Route to current view
	var cmd tea.Cmd
	switch a.view {
	case ViewList:
		cmd = a.updateList(msg)
	case ViewDetail:
		cmd = a.updateDetail(msg)
	case ViewAdd, ViewEdit:
		cmd = a.updateInput(msg)
	case ViewSearch:
		cmd = a.updateSearch(msg)
	case ViewTagFilter:
		cmd = a.updateTagFilter(msg)
	}

	return a, cmd
}

func (a *App) updateList(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			a.quitting = true
			return tea.Quit

		case "enter":
			if todo := a.list.SelectedTodo(); todo != nil {
				a.detail.SetTodo(todo)
				a.view = ViewDetail
			}
			return nil

		case "n":
			a.input.Reset()
			a.input.mode = InputModeAdd
			a.view = ViewAdd
			return a.input.Init()

		case "e":
			if todo := a.list.SelectedTodo(); todo != nil {
				a.input.SetTodo(todo)
				a.input.mode = InputModeEdit
				a.view = ViewEdit
				return a.input.Init()
			}
			return nil

		case "/":
			a.list.searchMode = true
			a.list.searchInput.Focus()
			a.view = ViewSearch
			return nil

		case "t":
			a.view = ViewTagFilter
			return a.list.loadTags()

		case " ":
			return a.list.toggleComplete()

		case "x":
			a.list.toggleSelect()
			return nil

		case "D":
			return a.list.deleteSelected()

		case "p":
			return a.list.cyclePriority()
		}
	}

	return a.list.Update(msg)
}

func (a *App) updateDetail(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc":
			a.view = ViewList
			return nil

		case "e":
			if a.detail.todo != nil {
				a.input.SetTodo(a.detail.todo)
				a.input.mode = InputModeEdit
				a.view = ViewEdit
				return a.input.Init()
			}
			return nil

		case " ":
			if a.detail.todo != nil {
				cmd := a.detail.toggleComplete(a.store)
				a.view = ViewList
				return tea.Batch(cmd, a.list.loadTodos())
			}
			return nil
		}
	}

	return a.detail.Update(msg)
}

func (a *App) updateInput(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.view = ViewList
			return nil

		case "ctrl+s", "enter":
			if a.input.focusIndex == len(a.input.inputs)-1 || msg.String() == "ctrl+s" {
				// Save
				if a.input.mode == InputModeAdd {
					cmd := a.input.createTodo(a.store)
					a.view = ViewList
					return cmd
				} else {
					cmd := a.input.updateTodo(a.store)
					a.view = ViewList
					return cmd
				}
			}
		}
	}

	return a.input.Update(msg)
}

func (a *App) updateSearch(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			a.list.searchMode = false
			a.list.searchInput.Blur()
			a.list.searchInput.SetValue("")
			a.list.filter.Search = ""
			a.view = ViewList
			return a.list.loadTodos()

		case "enter":
			a.list.filter.Search = a.list.searchInput.Value()
			a.list.searchMode = false
			a.list.searchInput.Blur()
			a.view = ViewList
			return a.list.loadTodos()
		}
	}

	return a.list.UpdateSearch(msg)
}

func (a *App) updateTagFilter(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "q":
			a.view = ViewList
			return nil

		case "enter":
			a.list.applyTagFilter()
			a.view = ViewList
			return a.list.loadTodos()

		case "j", "down":
			a.list.tagCursor++
			if a.list.tagCursor >= len(a.list.allTags)+1 {
				a.list.tagCursor = 0
			}
			return nil

		case "k", "up":
			a.list.tagCursor--
			if a.list.tagCursor < 0 {
				a.list.tagCursor = len(a.list.allTags)
			}
			return nil

		case " ":
			a.list.toggleTagSelection()
			return nil
		}
	}

	return nil
}

// View renders the application
func (a *App) View() string {
	if a.quitting {
		return ""
	}

	var content string

	switch a.view {
	case ViewList:
		content = a.list.View()
	case ViewDetail:
		content = a.detail.View()
	case ViewAdd:
		content = a.input.View()
	case ViewEdit:
		content = a.input.View()
	case ViewSearch:
		content = a.list.ViewSearch()
	case ViewTagFilter:
		content = a.list.ViewTagFilter()
	}

	if a.err != nil {
		content += "\n" + errorStyle.Render("Error: "+a.err.Error())
	}

	return appStyle.Render(content)
}
