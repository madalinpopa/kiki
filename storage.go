package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

const (
	kikiDir       = "kiki"
	tasksFile     = "tasks.json"
	notesFile     = "notes.json"
	configDirPerm = 0o755
	dataFilePerm  = 0o644
	dateLayout    = "2006-01-02"
)

// Storage handles all file operations for Kiki
type Storage struct {
	basePath string
}

// GetConfigDir returns the kiki config directory path using XDG_CONFIG_HOME
func GetConfigDir() string {
	configHome := os.Getenv("XDG_CONFIG_HOME")
	if configHome == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			configHome = "."
		} else {
			configHome = filepath.Join(homeDir, ".config")
		}
	}
	return filepath.Join(configHome, kikiDir)
}

// NewStorage creates a new Storage instance and ensures the directory exists
func NewStorage() (*Storage, error) {
	basePath := GetConfigDir()
	if err := os.MkdirAll(basePath, configDirPerm); err != nil {
		return nil, fmt.Errorf("failed to create kiki directory: %w", err)
	}

	return &Storage{basePath: basePath}, nil
}

// InitStorage creates all required directories and files
func InitStorage() error {
	basePath := GetConfigDir()

	if err := os.MkdirAll(basePath, configDirPerm); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	tasksPath := filepath.Join(basePath, tasksFile)
	if _, err := os.Stat(tasksPath); os.IsNotExist(err) {
		emptyTasks := &TaskList{Tasks: []Task{}}
		data, _ := json.MarshalIndent(emptyTasks, "", "  ")
		if err := os.WriteFile(tasksPath, data, dataFilePerm); err != nil {
			return fmt.Errorf("failed to create tasks.json: %w", err)
		}
	}

	notesPath := filepath.Join(basePath, notesFile)
	if _, err := os.Stat(notesPath); os.IsNotExist(err) {
		emptyNotes := &NoteList{Notes: []Note{}}
		data, _ := json.MarshalIndent(emptyNotes, "", "  ")
		if err := os.WriteFile(notesPath, data, dataFilePerm); err != nil {
			return fmt.Errorf("failed to create notes.json: %w", err)
		}
	}

	return nil
}

// generateID creates a UUIDv7 string
func generateID() string {
	id, err := uuid.NewV7()
	if err != nil {
		return uuid.New().String()
	}
	return id.String()
}

// LoadTasks reads tasks from tasks.json
func (s *Storage) LoadTasks() (*TaskList, error) {
	path := filepath.Join(s.basePath, tasksFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &TaskList{Tasks: []Task{}}, nil
		}
		return nil, fmt.Errorf("failed to read tasks: %w", err)
	}

	var tasks TaskList
	if err := json.Unmarshal(data, &tasks); err != nil {
		return nil, fmt.Errorf("failed to parse tasks: %w", err)
	}
	return &tasks, nil
}

// SaveTasks writes tasks to tasks.json
func (s *Storage) SaveTasks(tasks *TaskList) error {
	path := filepath.Join(s.basePath, tasksFile)
	data, err := json.MarshalIndent(tasks, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize tasks: %w", err)
	}
	return os.WriteFile(path, data, dataFilePerm)
}

// LoadNotes reads notes from notes.json
func (s *Storage) LoadNotes() (*NoteList, error) {
	path := filepath.Join(s.basePath, notesFile)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &NoteList{Notes: []Note{}}, nil
		}
		return nil, fmt.Errorf("failed to read notes: %w", err)
	}

	var notes NoteList
	if err := json.Unmarshal(data, &notes); err != nil {
		return nil, fmt.Errorf("failed to parse notes: %w", err)
	}
	return &notes, nil
}

// SaveNotes writes notes to notes.json
func (s *Storage) SaveNotes(notes *NoteList) error {
	path := filepath.Join(s.basePath, notesFile)
	data, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to serialize notes: %w", err)
	}
	return os.WriteFile(path, data, dataFilePerm)
}

// AddTask creates a new task and saves it
func (s *Storage) AddTask(title string, dueDate *string, priority string, tags []string) (*Task, error) {
	tasks, err := s.LoadTasks()
	if err != nil {
		return nil, err
	}

	if priority == "" {
		priority = "medium"
	}
	if tags == nil {
		tags = []string{}
	}

	task := Task{
		ID:        generateID(),
		Title:     title,
		Completed: false,
		DueDate:   dueDate,
		Priority:  priority,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tasks.Tasks = append(tasks.Tasks, task)
	if err := s.SaveTasks(tasks); err != nil {
		return nil, err
	}

	return &task, nil
}

// AddNote creates a new note and saves it
func (s *Storage) AddNote(title, content string, tags []string) (*Note, error) {
	notes, err := s.LoadNotes()
	if err != nil {
		return nil, err
	}

	if tags == nil {
		tags = []string{}
	}

	note := Note{
		ID:        generateID(),
		Title:     title,
		Content:   content,
		Tags:      tags,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	notes.Notes = append(notes.Notes, note)
	if err := s.SaveNotes(notes); err != nil {
		return nil, err
	}

	return &note, nil
}

// isToday checks if a date string (YYYY-MM-DD) or time is today
func isToday(dateStr *string) bool {
	if dateStr == nil {
		return false
	}
	today := time.Now().Format(dateLayout)
	return *dateStr == today
}

// isTodayTime checks if a time.Time is today
func isTodayTime(t time.Time) bool {
	now := time.Now()
	return t.Year() == now.Year() && t.Month() == now.Month() && t.Day() == now.Day()
}
