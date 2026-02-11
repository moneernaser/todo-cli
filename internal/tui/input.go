package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"todo_cli/internal/model"
	"todo_cli/internal/storage"
)

// InputMode defines whether we're adding or editing
type InputMode int

const (
	InputModeAdd InputMode = iota
	InputModeEdit
)

// InputView handles adding and editing todos
type InputView struct {
	mode       InputMode
	todo       *model.Todo
	inputs     []textinput.Model
	focusIndex int
	labels     []string
	err        string
}

const (
	inputTitle = iota
	inputDescription
	inputTags
	inputDueDate
	inputPriority
)

// NewInputView creates a new input view
func NewInputView() *InputView {
	labels := []string{"Title", "Description", "Tags", "Due Date", "Priority (1-5)"}
	inputs := make([]textinput.Model, len(labels))

	for i := range inputs {
		ti := textinput.New()
		ti.CharLimit = 200
		ti.Width = 40

		switch i {
		case inputTitle:
			ti.Placeholder = "What needs to be done?"
		case inputDescription:
			ti.Placeholder = "Optional details..."
		case inputTags:
			ti.Placeholder = "#work #urgent"
		case inputDueDate:
			ti.Placeholder = "2026-02-14 or today/tomorrow"
		case inputPriority:
			ti.Placeholder = "1-5 (1=urgent, 5=lowest)"
			ti.CharLimit = 1
			ti.Width = 5
		}

		inputs[i] = ti
	}

	return &InputView{
		inputs: inputs,
		labels: labels,
	}
}

// Reset clears the input form
func (v *InputView) Reset() {
	v.todo = nil
	v.focusIndex = 0
	v.err = ""
	for i := range v.inputs {
		v.inputs[i].SetValue("")
		v.inputs[i].Blur()
	}
	v.inputs[0].Focus()
}

// SetTodo populates the form with existing todo data
func (v *InputView) SetTodo(todo *model.Todo) {
	v.todo = todo
	v.focusIndex = 0
	v.err = ""

	v.inputs[inputTitle].SetValue(todo.Title)
	v.inputs[inputDescription].SetValue(todo.Description)
	v.inputs[inputTags].SetValue(strings.Join(todo.Tags, " "))

	if todo.DueDate != nil {
		v.inputs[inputDueDate].SetValue(todo.DueDate.Format("2006-01-02"))
	} else {
		v.inputs[inputDueDate].SetValue("")
	}

	if todo.Priority > 0 {
		v.inputs[inputPriority].SetValue(string('0' + byte(todo.Priority)))
	} else {
		v.inputs[inputPriority].SetValue("")
	}

	for i := range v.inputs {
		v.inputs[i].Blur()
	}
	v.inputs[0].Focus()
}

// Init initializes the input view
func (v *InputView) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles input for the form
func (v *InputView) Update(msg tea.Msg) tea.Cmd {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "down":
			v.focusIndex++
			if v.focusIndex >= len(v.inputs) {
				v.focusIndex = 0
			}
			return v.updateFocus()

		case "shift+tab", "up":
			v.focusIndex--
			if v.focusIndex < 0 {
				v.focusIndex = len(v.inputs) - 1
			}
			return v.updateFocus()
		}
	}

	// Update the focused input
	var cmd tea.Cmd
	v.inputs[v.focusIndex], cmd = v.inputs[v.focusIndex].Update(msg)
	return cmd
}

func (v *InputView) updateFocus() tea.Cmd {
	cmds := make([]tea.Cmd, len(v.inputs))
	for i := range v.inputs {
		if i == v.focusIndex {
			cmds[i] = v.inputs[i].Focus()
		} else {
			v.inputs[i].Blur()
		}
	}
	return tea.Batch(cmds...)
}

func (v *InputView) createTodo(store storage.Storage) tea.Cmd {
	title := strings.TrimSpace(v.inputs[inputTitle].Value())
	if title == "" {
		v.err = "Title is required"
		return nil
	}

	todo := &model.Todo{
		Title:       title,
		Description: strings.TrimSpace(v.inputs[inputDescription].Value()),
		Tags:        storage.ParseTags(v.inputs[inputTags].Value()),
	}

	// Parse due date
	dueDateStr := strings.TrimSpace(v.inputs[inputDueDate].Value())
	if dueDateStr != "" {
		dueDate, err := storage.ParseDueDate(dueDateStr)
		if err != nil {
			v.err = "Invalid due date format"
			return nil
		}
		todo.DueDate = dueDate
	}

	// Parse priority
	priorityStr := strings.TrimSpace(v.inputs[inputPriority].Value())
	if priorityStr != "" {
		priority := int(priorityStr[0] - '0')
		if priority < 1 || priority > 5 {
			v.err = "Priority must be 1-5"
			return nil
		}
		todo.Priority = priority
	}

	return func() tea.Msg {
		if err := store.Create(todo); err != nil {
			return errMsg{err}
		}
		return todoCreatedMsg{}
	}
}

func (v *InputView) updateTodo(store storage.Storage) tea.Cmd {
	if v.todo == nil {
		return nil
	}

	title := strings.TrimSpace(v.inputs[inputTitle].Value())
	if title == "" {
		v.err = "Title is required"
		return nil
	}

	v.todo.Title = title
	v.todo.Description = strings.TrimSpace(v.inputs[inputDescription].Value())
	v.todo.Tags = storage.ParseTags(v.inputs[inputTags].Value())

	// Parse due date
	dueDateStr := strings.TrimSpace(v.inputs[inputDueDate].Value())
	if dueDateStr != "" {
		dueDate, err := storage.ParseDueDate(dueDateStr)
		if err != nil {
			v.err = "Invalid due date format"
			return nil
		}
		v.todo.DueDate = dueDate
	} else {
		v.todo.DueDate = nil
	}

	// Parse priority
	priorityStr := strings.TrimSpace(v.inputs[inputPriority].Value())
	if priorityStr != "" {
		priority := int(priorityStr[0] - '0')
		if priority < 1 || priority > 5 {
			v.err = "Priority must be 1-5"
			return nil
		}
		v.todo.Priority = priority
	} else {
		v.todo.Priority = 0
	}

	todo := v.todo
	return func() tea.Msg {
		if err := store.Update(todo); err != nil {
			return errMsg{err}
		}
		return todoUpdatedMsg{}
	}
}

// View renders the input form
func (v *InputView) View() string {
	var b strings.Builder

	// Title
	title := "New Todo"
	if v.mode == InputModeEdit {
		title = "Edit Todo"
	}
	b.WriteString(titleStyle.Render(title))
	b.WriteString("\n\n")

	// Form fields
	for i, input := range v.inputs {
		b.WriteString(inputLabelStyle.Render(v.labels[i]))
		b.WriteString("\n")

		if i == v.focusIndex {
			b.WriteString(focusedInputStyle.Render(input.View()))
		} else {
			b.WriteString(inputStyle.Render(input.View()))
		}
		b.WriteString("\n\n")
	}

	// Error message
	if v.err != "" {
		b.WriteString(errorStyle.Render("Error: " + v.err))
		b.WriteString("\n\n")
	}

	// Help
	b.WriteString(helpStyle.Render("tab/↓: next field  shift+tab/↑: prev field  ctrl+s/enter: save  esc: cancel"))

	return b.String()
}
