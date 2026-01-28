package main

import "time"

// Task represents a todo item with metadata
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Completed bool      `json:"completed"`
	DueDate   *string   `json:"due_date,omitempty"` // YYYY-MM-DD format
	Priority  string    `json:"priority"`           // low, medium, high
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Note represents a text note with metadata
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Tags      []string  `json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskList holds all tasks
type TaskList struct {
	Tasks []Task `json:"tasks"`
}

// NoteList holds all notes
type NoteList struct {
	Notes []Note `json:"notes"`
}
