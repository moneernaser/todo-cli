package storage

import (
	"time"

	"todo_cli/internal/model"
)

// SortField defines what field to sort by
type SortField string

const (
	SortByCreated  SortField = "created"
	SortByUpdated  SortField = "updated"
	SortByDueDate  SortField = "due"
	SortByPriority SortField = "priority"
	SortByTitle    SortField = "title"
)

// SortOrder defines ascending or descending order
type SortOrder string

const (
	SortAsc  SortOrder = "asc"
	SortDesc SortOrder = "desc"
)

// DueDateFilterType defines types of due date filters
type DueDateFilterType string

const (
	DueToday    DueDateFilterType = "today"
	DueTomorrow DueDateFilterType = "tomorrow"
	DueNextWeek DueDateFilterType = "next-week"
	DueOverdue  DueDateFilterType = "overdue"
	DueSpecific DueDateFilterType = "specific"
)

// DueDateFilter defines a filter for due dates
type DueDateFilter struct {
	Type         DueDateFilterType
	SpecificDate *time.Time
}

// Filter defines criteria for filtering todos
type Filter struct {
	Tags      []string
	Completed *bool
	DueDate   *DueDateFilter
	SortBy    SortField
	SortOrder SortOrder
	Search    string
}

// Storage defines the interface for todo persistence
type Storage interface {
	Create(todo *model.Todo) error
	GetByID(id int64) (*model.Todo, error)
	List(filter Filter) ([]model.Todo, error)
	Update(todo *model.Todo) error
	Delete(id int64) error
	GetAllTags() ([]string, error)
	Close() error
}
