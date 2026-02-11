package model

import (
	"time"
)

// Todo represents a single todo item
type Todo struct {
	ID          int64      `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Tags        []string   `json:"tags,omitempty"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	Completed   bool       `json:"completed"`
	Priority    int        `json:"priority,omitempty"` // 1-5 (1=highest), 0=no priority
}

// IsOverdue returns true if the todo has a due date in the past and is not completed
func (t *Todo) IsOverdue() bool {
	if t.Completed || t.DueDate == nil {
		return false
	}
	return t.DueDate.Before(time.Now())
}

// IsDueToday returns true if the todo is due today
func (t *Todo) IsDueToday() bool {
	if t.DueDate == nil {
		return false
	}
	now := time.Now()
	y1, m1, d1 := t.DueDate.Date()
	y2, m2, d2 := now.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsDueTomorrow returns true if the todo is due tomorrow
func (t *Todo) IsDueTomorrow() bool {
	if t.DueDate == nil {
		return false
	}
	tomorrow := time.Now().AddDate(0, 0, 1)
	y1, m1, d1 := t.DueDate.Date()
	y2, m2, d2 := tomorrow.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// IsDueWithinDays returns true if the todo is due within the given number of days
func (t *Todo) IsDueWithinDays(days int) bool {
	if t.DueDate == nil {
		return false
	}
	deadline := time.Now().AddDate(0, 0, days)
	return !t.DueDate.After(deadline)
}

// PriorityString returns a human-readable priority string
func (t *Todo) PriorityString() string {
	switch t.Priority {
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	case 5:
		return "Lowest"
	default:
		return ""
	}
}
