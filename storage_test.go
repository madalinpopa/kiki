package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetConfigDir(t *testing.T) {
	t.Run("uses XDG_CONFIG_HOME when set", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)

		// act
		got := GetConfigDir()

		// assert
		want := filepath.Join(configHome, kikiDir)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})

	t.Run("falls back to user home when XDG_CONFIG_HOME is empty", func(t *testing.T) {
		// arrange
		t.Setenv("XDG_CONFIG_HOME", "")
		homeDir, err := os.UserHomeDir()
		if err != nil {
			t.Fatalf("failed to get home dir: %v", err)
		}

		// act
		got := GetConfigDir()

		// assert
		want := filepath.Join(homeDir, ".config", kikiDir)
		if got != want {
			t.Fatalf("expected %q, got %q", want, got)
		}
	})
}

func TestNewStorage(t *testing.T) {
	t.Run("creates config directory and stores base path", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		basePath := filepath.Join(configHome, kikiDir)

		// act
		storage, err := NewStorage()

		// assert
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if storage.basePath != basePath {
			t.Fatalf("expected basePath %q, got %q", basePath, storage.basePath)
		}
		if _, err := os.Stat(basePath); err != nil {
			t.Fatalf("expected directory to exist: %v", err)
		}
	})
}

func TestInitStorage(t *testing.T) {
	t.Run("creates config directory and data files", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		basePath := filepath.Join(configHome, kikiDir)

		// act
		err := InitStorage()

		// assert
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if _, err := os.Stat(basePath); err != nil {
			t.Fatalf("expected config directory to exist: %v", err)
		}
		tasksPath := filepath.Join(basePath, tasksFile)
		notesPath := filepath.Join(basePath, notesFile)
		if _, err := os.Stat(tasksPath); err != nil {
			t.Fatalf("expected tasks file to exist: %v", err)
		}
		if _, err := os.Stat(notesPath); err != nil {
			t.Fatalf("expected notes file to exist: %v", err)
		}

		tasksData, err := os.ReadFile(tasksPath)
		if err != nil {
			t.Fatalf("failed to read tasks file: %v", err)
		}
		var tasks TaskList
		if err := json.Unmarshal(tasksData, &tasks); err != nil {
			t.Fatalf("failed to parse tasks file: %v", err)
		}
		if len(tasks.Tasks) != 0 {
			t.Fatalf("expected empty task list, got %d", len(tasks.Tasks))
		}

		notesData, err := os.ReadFile(notesPath)
		if err != nil {
			t.Fatalf("failed to read notes file: %v", err)
		}
		var notes NoteList
		if err := json.Unmarshal(notesData, &notes); err != nil {
			t.Fatalf("failed to parse notes file: %v", err)
		}
		if len(notes.Notes) != 0 {
			t.Fatalf("expected empty note list, got %d", len(notes.Notes))
		}
	})
}

func TestStorageTaskIO(t *testing.T) {
	t.Run("SaveTasks writes tasks and LoadTasks returns them", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		storage, err := NewStorage()
		if err != nil {
			t.Fatalf("failed to create storage: %v", err)
		}
		dueDate := "2030-01-02"
		input := &TaskList{
			Tasks: []Task{
				{
					ID:        "task-1",
					Title:     "Test task",
					Completed: true,
					DueDate:   &dueDate,
					Priority:  "high",
					Tags:      []string{"one", "two"},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		// act
		if err := storage.SaveTasks(input); err != nil {
			t.Fatalf("failed to save tasks: %v", err)
		}
		output, err := storage.LoadTasks()
		if err != nil {
			t.Fatalf("failed to load tasks: %v", err)
		}

		// assert
		if len(output.Tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(output.Tasks))
		}
		if output.Tasks[0].Title != input.Tasks[0].Title {
			t.Fatalf("expected title %q, got %q", input.Tasks[0].Title, output.Tasks[0].Title)
		}
		if output.Tasks[0].Priority != input.Tasks[0].Priority {
			t.Fatalf("expected priority %q, got %q", input.Tasks[0].Priority, output.Tasks[0].Priority)
		}
	})

	t.Run("LoadTasks returns empty list when file missing", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		storage, err := NewStorage()
		if err != nil {
			t.Fatalf("failed to create storage: %v", err)
		}

		// act
		tasks, err := storage.LoadTasks()

		// assert
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(tasks.Tasks) != 0 {
			t.Fatalf("expected empty tasks, got %d", len(tasks.Tasks))
		}
	})
}

func TestStorageNoteIO(t *testing.T) {
	t.Run("SaveNotes writes notes and LoadNotes returns them", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		storage, err := NewStorage()
		if err != nil {
			t.Fatalf("failed to create storage: %v", err)
		}
		input := &NoteList{
			Notes: []Note{
				{
					ID:        "note-1",
					Title:     "Test note",
					Content:   "Hello",
					Tags:      []string{"tag"},
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				},
			},
		}

		// act
		if err := storage.SaveNotes(input); err != nil {
			t.Fatalf("failed to save notes: %v", err)
		}
		output, err := storage.LoadNotes()
		if err != nil {
			t.Fatalf("failed to load notes: %v", err)
		}

		// assert
		if len(output.Notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(output.Notes))
		}
		if output.Notes[0].Title != input.Notes[0].Title {
			t.Fatalf("expected title %q, got %q", input.Notes[0].Title, output.Notes[0].Title)
		}
		if output.Notes[0].Content != input.Notes[0].Content {
			t.Fatalf("expected content %q, got %q", input.Notes[0].Content, output.Notes[0].Content)
		}
	})

	t.Run("LoadNotes returns empty list when file missing", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		storage, err := NewStorage()
		if err != nil {
			t.Fatalf("failed to create storage: %v", err)
		}

		// act
		notes, err := storage.LoadNotes()

		// assert
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if len(notes.Notes) != 0 {
			t.Fatalf("expected empty notes, got %d", len(notes.Notes))
		}
	})
}

func TestStorageAddTask(t *testing.T) {
	t.Run("AddTask applies defaults and persists task", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		storage, err := NewStorage()
		if err != nil {
			t.Fatalf("failed to create storage: %v", err)
		}
		before := time.Now()
		dueDate := "2032-03-04"

		// act
		created, err := storage.AddTask("Write tests", &dueDate, "", nil)
		after := time.Now()

		// assert
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if created.ID == "" {
			t.Fatalf("expected ID to be set")
		}
		if created.Title != "Write tests" {
			t.Fatalf("expected title to be set")
		}
		if created.Completed {
			t.Fatalf("expected task to be incomplete")
		}
		if created.Priority != "medium" {
			t.Fatalf("expected default priority 'medium', got %q", created.Priority)
		}
		if created.Tags == nil || len(created.Tags) != 0 {
			t.Fatalf("expected empty tags")
		}
		if created.DueDate == nil || *created.DueDate != dueDate {
			t.Fatalf("expected due date %q", dueDate)
		}
		if created.CreatedAt.Before(before) || created.CreatedAt.After(after) {
			t.Fatalf("expected CreatedAt within call window")
		}
		if created.UpdatedAt.Before(before) || created.UpdatedAt.After(after) {
			t.Fatalf("expected UpdatedAt within call window")
		}

		loaded, err := storage.LoadTasks()
		if err != nil {
			t.Fatalf("failed to load tasks: %v", err)
		}
		if len(loaded.Tasks) != 1 {
			t.Fatalf("expected 1 task, got %d", len(loaded.Tasks))
		}
	})
}

func TestStorageAddNote(t *testing.T) {
	t.Run("AddNote applies defaults and persists note", func(t *testing.T) {
		// arrange
		configHome := t.TempDir()
		t.Setenv("XDG_CONFIG_HOME", configHome)
		storage, err := NewStorage()
		if err != nil {
			t.Fatalf("failed to create storage: %v", err)
		}
		before := time.Now()

		// act
		created, err := storage.AddNote("Idea", "Something", nil)
		after := time.Now()

		// assert
		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}
		if created.ID == "" {
			t.Fatalf("expected ID to be set")
		}
		if created.Title != "Idea" {
			t.Fatalf("expected title to be set")
		}
		if created.Content != "Something" {
			t.Fatalf("expected content to be set")
		}
		if created.Tags == nil || len(created.Tags) != 0 {
			t.Fatalf("expected empty tags")
		}
		if created.CreatedAt.Before(before) || created.CreatedAt.After(after) {
			t.Fatalf("expected CreatedAt within call window")
		}
		if created.UpdatedAt.Before(before) || created.UpdatedAt.After(after) {
			t.Fatalf("expected UpdatedAt within call window")
		}

		loaded, err := storage.LoadNotes()
		if err != nil {
			t.Fatalf("failed to load notes: %v", err)
		}
		if len(loaded.Notes) != 1 {
			t.Fatalf("expected 1 note, got %d", len(loaded.Notes))
		}
	})
}

func TestGenerateID(t *testing.T) {
	t.Run("generateID returns non-empty string", func(t *testing.T) {
		// arrange

		// act
		id := generateID()

		// assert
		if id == "" {
			t.Fatalf("expected ID to be non-empty")
		}
	})
}

func TestIsToday(t *testing.T) {
	t.Run("returns false for nil date", func(t *testing.T) {
		// arrange
		var date *string

		// act
		got := isToday(date)

		// assert
		if got {
			t.Fatalf("expected false")
		}
	})

	t.Run("returns true for today's date", func(t *testing.T) {
		// arrange
		today := time.Now().Format(dateLayout)

		// act
		got := isToday(&today)

		// assert
		if !got {
			t.Fatalf("expected true")
		}
	})

	t.Run("returns false for other date", func(t *testing.T) {
		// arrange
		other := time.Now().AddDate(0, 0, -1).Format(dateLayout)

		// act
		got := isToday(&other)

		// assert
		if got {
			t.Fatalf("expected false")
		}
	})
}

func TestIsTodayTime(t *testing.T) {
	t.Run("returns true for time today", func(t *testing.T) {
		// arrange
		now := time.Now()

		// act
		got := isTodayTime(now)

		// assert
		if !got {
			t.Fatalf("expected true")
		}
	})

	t.Run("returns false for time not today", func(t *testing.T) {
		// arrange
		other := time.Now().AddDate(0, 0, -1)

		// act
		got := isTodayTime(other)

		// assert
		if got {
			t.Fatalf("expected false")
		}
	})
}
