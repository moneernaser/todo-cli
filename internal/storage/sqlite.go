package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	_ "github.com/mattn/go-sqlite3"

	"todo_cli/internal/model"
)

const schema = `
CREATE TABLE IF NOT EXISTS todos (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    description TEXT DEFAULT '',
    tags TEXT DEFAULT '[]',
    due_date DATETIME,
    created_at DATETIME NOT NULL,
    updated_at DATETIME NOT NULL,
    completed_at DATETIME,
    completed INTEGER DEFAULT 0,
    priority INTEGER DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_todos_completed ON todos(completed);
CREATE INDEX IF NOT EXISTS idx_todos_due_date ON todos(due_date);
CREATE INDEX IF NOT EXISTS idx_todos_priority ON todos(priority);
`

// SQLiteStorage implements Storage using SQLite
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage instance
func NewSQLiteStorage() (*SQLiteStorage, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, fmt.Errorf("failed to get database path: %w", err)
	}

	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

// NewSQLiteStorageWithPath creates a new SQLite storage at a specific path
func NewSQLiteStorageWithPath(dbPath string) (*SQLiteStorage, error) {
	// Ensure directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	db, err := sql.Open("sqlite3", dbPath+"?_foreign_keys=on")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Initialize schema
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return &SQLiteStorage{db: db}, nil
}

func getDBPath() (string, error) {
	return xdg.DataFile("todocli/todos.db")
}

// Create inserts a new todo into the database
func (s *SQLiteStorage) Create(todo *model.Todo) error {
	now := time.Now().UTC()
	todo.CreatedAt = now
	todo.UpdatedAt = now

	tagsJSON, err := json.Marshal(todo.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	result, err := s.db.Exec(`
		INSERT INTO todos (title, description, tags, due_date, created_at, updated_at, completed_at, completed, priority)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, todo.Title, todo.Description, string(tagsJSON), nullableTime(todo.DueDate),
		todo.CreatedAt, todo.UpdatedAt, nullableTime(todo.CompletedAt),
		boolToInt(todo.Completed), todo.Priority)

	if err != nil {
		return fmt.Errorf("failed to insert todo: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}

	todo.ID = id
	return nil
}

// GetByID retrieves a todo by its ID
func (s *SQLiteStorage) GetByID(id int64) (*model.Todo, error) {
	row := s.db.QueryRow(`
		SELECT id, title, description, tags, due_date, created_at, updated_at, completed_at, completed, priority
		FROM todos WHERE id = ?
	`, id)

	todo, err := scanTodo(row)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("todo with ID %d not found", id)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get todo: %w", err)
	}

	return todo, nil
}

// List retrieves todos matching the given filter
func (s *SQLiteStorage) List(filter Filter) ([]model.Todo, error) {
	query := "SELECT id, title, description, tags, due_date, created_at, updated_at, completed_at, completed, priority FROM todos WHERE 1=1"
	args := []interface{}{}

	// Completed filter
	if filter.Completed != nil {
		query += " AND completed = ?"
		args = append(args, boolToInt(*filter.Completed))
	}

	// Tags filter
	if len(filter.Tags) > 0 {
		for _, tag := range filter.Tags {
			query += " AND tags LIKE ?"
			args = append(args, "%"+tag+"%")
		}
	}

	// Search filter
	if filter.Search != "" {
		query += " AND (title LIKE ? OR description LIKE ?)"
		searchPattern := "%" + filter.Search + "%"
		args = append(args, searchPattern, searchPattern)
	}

	// Due date filter
	if filter.DueDate != nil {
		now := time.Now()
		switch filter.DueDate.Type {
		case DueToday:
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			endOfDay := startOfDay.AddDate(0, 0, 1)
			query += " AND due_date >= ? AND due_date < ?"
			args = append(args, startOfDay, endOfDay)
		case DueTomorrow:
			startOfTomorrow := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).AddDate(0, 0, 1)
			endOfTomorrow := startOfTomorrow.AddDate(0, 0, 1)
			query += " AND due_date >= ? AND due_date < ?"
			args = append(args, startOfTomorrow, endOfTomorrow)
		case DueNextWeek:
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			endOfWeek := startOfDay.AddDate(0, 0, 7)
			query += " AND due_date >= ? AND due_date < ?"
			args = append(args, startOfDay, endOfWeek)
		case DueOverdue:
			startOfDay := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
			query += " AND due_date < ? AND completed = 0"
			args = append(args, startOfDay)
		case DueSpecific:
			if filter.DueDate.SpecificDate != nil {
				d := *filter.DueDate.SpecificDate
				startOfDay := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
				endOfDay := startOfDay.AddDate(0, 0, 1)
				query += " AND due_date >= ? AND due_date < ?"
				args = append(args, startOfDay, endOfDay)
			}
		}
	}

	// Sorting
	sortField := "created_at"
	switch filter.SortBy {
	case SortByCreated:
		sortField = "created_at"
	case SortByUpdated:
		sortField = "updated_at"
	case SortByDueDate:
		sortField = "due_date"
	case SortByPriority:
		sortField = "priority"
	case SortByTitle:
		sortField = "title"
	}

	sortOrder := "DESC"
	if filter.SortOrder == SortAsc {
		sortOrder = "ASC"
	}

	// Special handling for priority sort (0 should go last)
	if filter.SortBy == SortByPriority {
		if sortOrder == "ASC" {
			query += fmt.Sprintf(" ORDER BY CASE WHEN priority = 0 THEN 999 ELSE priority END %s", sortOrder)
		} else {
			query += fmt.Sprintf(" ORDER BY CASE WHEN priority = 0 THEN -1 ELSE priority END %s", sortOrder)
		}
	} else {
		query += fmt.Sprintf(" ORDER BY %s %s", sortField, sortOrder)
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query todos: %w", err)
	}
	defer rows.Close()

	var todos []model.Todo
	for rows.Next() {
		todo, err := scanTodoRows(rows)
		if err != nil {
			return nil, fmt.Errorf("failed to scan todo: %w", err)
		}
		todos = append(todos, *todo)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating todos: %w", err)
	}

	return todos, nil
}

// Update updates an existing todo
func (s *SQLiteStorage) Update(todo *model.Todo) error {
	todo.UpdatedAt = time.Now().UTC()

	tagsJSON, err := json.Marshal(todo.Tags)
	if err != nil {
		return fmt.Errorf("failed to marshal tags: %w", err)
	}

	result, err := s.db.Exec(`
		UPDATE todos SET
			title = ?, description = ?, tags = ?, due_date = ?,
			updated_at = ?, completed_at = ?, completed = ?, priority = ?
		WHERE id = ?
	`, todo.Title, todo.Description, string(tagsJSON), nullableTime(todo.DueDate),
		todo.UpdatedAt, nullableTime(todo.CompletedAt), boolToInt(todo.Completed),
		todo.Priority, todo.ID)

	if err != nil {
		return fmt.Errorf("failed to update todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("todo with ID %d not found", todo.ID)
	}

	return nil
}

// Delete removes a todo by ID
func (s *SQLiteStorage) Delete(id int64) error {
	result, err := s.db.Exec("DELETE FROM todos WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete todo: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("todo with ID %d not found", id)
	}

	return nil
}

// GetAllTags returns all unique tags from all todos
func (s *SQLiteStorage) GetAllTags() ([]string, error) {
	rows, err := s.db.Query("SELECT tags FROM todos")
	if err != nil {
		return nil, fmt.Errorf("failed to query tags: %w", err)
	}
	defer rows.Close()

	tagSet := make(map[string]struct{})
	for rows.Next() {
		var tagsJSON string
		if err := rows.Scan(&tagsJSON); err != nil {
			return nil, fmt.Errorf("failed to scan tags: %w", err)
		}

		var tags []string
		if err := json.Unmarshal([]byte(tagsJSON), &tags); err != nil {
			continue // Skip invalid JSON
		}

		for _, tag := range tags {
			tagSet[tag] = struct{}{}
		}
	}

	tags := make([]string, 0, len(tagSet))
	for tag := range tagSet {
		tags = append(tags, tag)
	}

	return tags, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}

// Helper functions

func scanTodo(row *sql.Row) (*model.Todo, error) {
	var todo model.Todo
	var tagsJSON string
	var dueDate, completedAt sql.NullTime
	var completed int

	err := row.Scan(
		&todo.ID, &todo.Title, &todo.Description, &tagsJSON,
		&dueDate, &todo.CreatedAt, &todo.UpdatedAt, &completedAt,
		&completed, &todo.Priority,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &todo.Tags); err != nil {
		todo.Tags = []string{}
	}

	if dueDate.Valid {
		todo.DueDate = &dueDate.Time
	}

	if completedAt.Valid {
		todo.CompletedAt = &completedAt.Time
	}

	todo.Completed = completed == 1

	return &todo, nil
}

func scanTodoRows(rows *sql.Rows) (*model.Todo, error) {
	var todo model.Todo
	var tagsJSON string
	var dueDate, completedAt sql.NullTime
	var completed int

	err := rows.Scan(
		&todo.ID, &todo.Title, &todo.Description, &tagsJSON,
		&dueDate, &todo.CreatedAt, &todo.UpdatedAt, &completedAt,
		&completed, &todo.Priority,
	)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(tagsJSON), &todo.Tags); err != nil {
		todo.Tags = []string{}
	}

	if dueDate.Valid {
		todo.DueDate = &dueDate.Time
	}

	if completedAt.Valid {
		todo.CompletedAt = &completedAt.Time
	}

	todo.Completed = completed == 1

	return &todo, nil
}

func nullableTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return *t
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ParseTags parses a tag string like "#work #urgent" into a slice
func ParseTags(tagStr string) []string {
	if tagStr == "" {
		return nil
	}

	parts := strings.Fields(tagStr)
	tags := make([]string, 0, len(parts))
	for _, part := range parts {
		tag := strings.TrimSpace(part)
		if tag != "" {
			// Ensure tag starts with #
			if !strings.HasPrefix(tag, "#") {
				tag = "#" + tag
			}
			tags = append(tags, tag)
		}
	}
	return tags
}

// ParseDueDate parses a due date string like "2026-02-14" or "today" into a time.Time
func ParseDueDate(dateStr string) (*time.Time, error) {
	if dateStr == "" {
		return nil, nil
	}

	now := time.Now()

	switch strings.ToLower(dateStr) {
	case "today":
		t := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location())
		return &t, nil
	case "tomorrow":
		t := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).AddDate(0, 0, 1)
		return &t, nil
	case "next-week", "nextweek":
		t := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, now.Location()).AddDate(0, 0, 7)
		return &t, nil
	default:
		// Try parsing as date
		formats := []string{
			"2006-01-02",
			"2006/01/02",
			"01-02-2006",
			"01/02/2006",
			"Jan 2, 2006",
			"January 2, 2006",
		}

		for _, format := range formats {
			if t, err := time.Parse(format, dateStr); err == nil {
				// Set to end of day
				t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, now.Location())
				return &t, nil
			}
		}

		return nil, fmt.Errorf("invalid date format: %s", dateStr)
	}
}
