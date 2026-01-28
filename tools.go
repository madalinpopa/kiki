package main

import (
	"fmt"
	"log/slog"
	"strings"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

const (
	taskNumberOffset = 1
	noteNumberStart  = 0
	notFoundIndex    = -1
	notePreviewMax   = 100
)

// ToolHandler wraps storage for tool operations
type ToolHandler struct {
	storage *Storage
	logger  *slog.Logger
}

// NewToolHandler creates a new tool handler
func NewToolHandler(storage *Storage, logger *slog.Logger) *ToolHandler {
	return &ToolHandler{storage: storage, logger: logger}
}

// AddTaskParams parameters for add_task tool
type AddTaskParams struct {
	Title    string   `json:"title" jsonschema:"The task title"`
	DueDate  *string  `json:"due_date,omitempty" jsonschema:"Due date in YYYY-MM-DD format"`
	Priority *string  `json:"priority,omitempty" jsonschema:"Priority level: low, medium, or high"`
	Tags     []string `json:"tags,omitempty" jsonschema:"Optional tags for categorization"`
}

// AddTaskResult result from add_task tool
type AddTaskResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	TaskID  string `json:"task_id,omitempty"`
}

// ListTasksParams parameters for list_tasks tool
type ListTasksParams struct {
	Filter string `json:"filter" jsonschema:"Filter: all, today, incomplete, or completed"`
}

// ListTasksResult result from list_tasks tool
type ListTasksResult struct {
	Tasks   []TaskSummary `json:"tasks"`
	Count   int           `json:"count"`
	Message string        `json:"message"`
}

// TaskSummary simplified task for listing
type TaskSummary struct {
	Number    int     `json:"number"`
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Completed bool    `json:"completed"`
	DueDate   *string `json:"due_date,omitempty"`
	Priority  string  `json:"priority"`
}

// CompleteTaskParams parameters for complete_task tool
type CompleteTaskParams struct {
	Query string `json:"query" jsonschema:"Task ID or title substring to match"`
}

// CompleteTaskResult result from complete_task tool
type CompleteTaskResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteTaskParams parameters for delete_task tool
type DeleteTaskParams struct {
	Query string `json:"query" jsonschema:"Task ID or title substring to match"`
}

// DeleteTaskResult result from delete_task tool
type DeleteTaskResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// AddNoteParams parameters for add_note tool
type AddNoteParams struct {
	Title   string   `json:"title" jsonschema:"The note title"`
	Content string   `json:"content" jsonschema:"The note content"`
	Tags    []string `json:"tags,omitempty" jsonschema:"Optional tags for categorization"`
}

// AddNoteResult result from add_note tool
type AddNoteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	NoteID  string `json:"note_id,omitempty"`
}

// ListNotesParams parameters for list_notes tool
type ListNotesParams struct {
	Filter string  `json:"filter" jsonschema:"Filter: all or today"`
	Tag    *string `json:"tag,omitempty" jsonschema:"Optional tag to filter by"`
}

// ListNotesResult result from list_notes tool
type ListNotesResult struct {
	Notes   []NoteSummary `json:"notes"`
	Count   int           `json:"count"`
	Message string        `json:"message"`
}

// NoteSummary simplified note for listing
type NoteSummary struct {
	Number  int      `json:"number"`
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Preview string   `json:"preview"`
	Tags    []string `json:"tags"`
}

// SearchNotesParams parameters for search_notes tool
type SearchNotesParams struct {
	Query string `json:"query" jsonschema:"Search term to find in title or content"`
}

// SearchNotesResult result from search_notes tool
type SearchNotesResult struct {
	Notes   []NoteSummary `json:"notes"`
	Count   int           `json:"count"`
	Message string        `json:"message"`
}

// DeleteNoteParams parameters for delete_note tool
type DeleteNoteParams struct {
	Query string `json:"query" jsonschema:"Note ID or title substring to match"`
}

// DeleteNoteResult result from delete_note tool
type DeleteNoteResult struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// GetAllTools returns all Kiki tools
func (h *ToolHandler) GetAllTools() []copilot.Tool {
	return []copilot.Tool{
		h.addTaskTool(),
		h.listTasksTool(),
		h.completeTaskTool(),
		h.deleteTaskTool(),
		h.addNoteTool(),
		h.listNotesTool(),
		h.searchNotesTool(),
		h.deleteNoteTool(),
	}
}

func (h *ToolHandler) addTaskTool() copilot.Tool {
	return copilot.DefineTool(
		"add_task",
		"Create a new task with optional due date, priority, and tags",
		func(params AddTaskParams, inv copilot.ToolInvocation) (AddTaskResult, error) {
			priority := "medium"
			if params.Priority != nil {
				priority = *params.Priority
			}

			task, err := h.storage.AddTask(params.Title, params.DueDate, priority, params.Tags)
			if err != nil {
				return AddTaskResult{Success: false, Message: err.Error()}, nil
			}

			return AddTaskResult{
				Success: true,
				Message: fmt.Sprintf("Task '%s' created with %s priority", task.Title, task.Priority),
				TaskID:  task.ID,
			}, nil
		},
	)
}

func (h *ToolHandler) listTasksTool() copilot.Tool {
	return copilot.DefineTool(
		"list_tasks",
		"List tasks with filter: all, today (due or created today), incomplete, or completed. Returns numbered list for easy reference.",
		func(params ListTasksParams, inv copilot.ToolInvocation) (ListTasksResult, error) {
			taskList, err := h.storage.LoadTasks()
			if err != nil {
				return ListTasksResult{Message: err.Error()}, nil
			}

			filtered := make([]TaskSummary, 0, len(taskList.Tasks))
			for i, t := range taskList.Tasks {
				include := false
				switch params.Filter {
				case "all":
					include = true
				case "today":
					include = isToday(t.DueDate) || isTodayTime(t.CreatedAt)
				case "incomplete":
					include = !t.Completed
				case "completed":
					include = t.Completed
				default:
					include = true
				}

				if include {
					filtered = append(filtered, TaskSummary{
						Number:    i + taskNumberOffset,
						ID:        t.ID,
						Title:     t.Title,
						Completed: t.Completed,
						DueDate:   t.DueDate,
						Priority:  t.Priority,
					})
				}
			}

			return ListTasksResult{
				Tasks:   filtered,
				Count:   len(filtered),
				Message: fmt.Sprintf("Found %d tasks", len(filtered)),
			}, nil
		},
	)
}

func (h *ToolHandler) completeTaskTool() copilot.Tool {
	return copilot.DefineTool(
		"complete_task",
		"Mark a task as completed by ID or title match",
		func(params CompleteTaskParams, inv copilot.ToolInvocation) (CompleteTaskResult, error) {
			taskList, err := h.storage.LoadTasks()
			if err != nil {
				return CompleteTaskResult{Success: false, Message: err.Error()}, nil
			}

			foundIndex, matchedTitle := findTaskIndex(taskList.Tasks, params.Query)
			if foundIndex == notFoundIndex {
				return CompleteTaskResult{
					Success: false,
					Message: fmt.Sprintf("No task found matching '%s'", params.Query),
				}, nil
			}

			taskList.Tasks[foundIndex].Completed = true
			taskList.Tasks[foundIndex].UpdatedAt = time.Now()

			if err := h.storage.SaveTasks(taskList); err != nil {
				return CompleteTaskResult{Success: false, Message: err.Error()}, nil
			}

			return CompleteTaskResult{
				Success: true,
				Message: fmt.Sprintf("Task '%s' marked as completed", matchedTitle),
			}, nil
		},
	)
}

func (h *ToolHandler) deleteTaskTool() copilot.Tool {
	return copilot.DefineTool(
		"delete_task",
		"Delete a task by ID or title match",
		func(params DeleteTaskParams, inv copilot.ToolInvocation) (DeleteTaskResult, error) {
			taskList, err := h.storage.LoadTasks()
			if err != nil {
				return DeleteTaskResult{Success: false, Message: err.Error()}, nil
			}

			foundIndex, matchedTitle := findTaskIndex(taskList.Tasks, params.Query)
			if foundIndex == notFoundIndex {
				return DeleteTaskResult{
					Success: false,
					Message: fmt.Sprintf("No task found matching '%s'", params.Query),
				}, nil
			}

			taskList.Tasks = append(taskList.Tasks[:foundIndex], taskList.Tasks[foundIndex+1:]...)

			if err := h.storage.SaveTasks(taskList); err != nil {
				return DeleteTaskResult{Success: false, Message: err.Error()}, nil
			}

			return DeleteTaskResult{
				Success: true,
				Message: fmt.Sprintf("Task '%s' deleted", matchedTitle),
			}, nil
		},
	)
}

func (h *ToolHandler) addNoteTool() copilot.Tool {
	return copilot.DefineTool(
		"add_note",
		"Create a new note with title, content, and optional tags",
		func(params AddNoteParams, inv copilot.ToolInvocation) (AddNoteResult, error) {
			note, err := h.storage.AddNote(params.Title, params.Content, params.Tags)
			if err != nil {
				return AddNoteResult{Success: false, Message: err.Error()}, nil
			}

			return AddNoteResult{
				Success: true,
				Message: fmt.Sprintf("Note '%s' created", note.Title),
				NoteID:  note.ID,
			}, nil
		},
	)
}

func (h *ToolHandler) listNotesTool() copilot.Tool {
	return copilot.DefineTool(
		"list_notes",
		"List notes with optional filter (all or today) and tag. Returns numbered list for easy reference.",
		func(params ListNotesParams, inv copilot.ToolInvocation) (ListNotesResult, error) {
			noteList, err := h.storage.LoadNotes()
			if err != nil {
				return ListNotesResult{Message: err.Error()}, nil
			}

			filtered := make([]NoteSummary, 0, len(noteList.Notes))
			noteNum := noteNumberStart
			for _, n := range noteList.Notes {
				include := true

				// Apply date filter
				if params.Filter == "today" {
					include = isTodayTime(n.CreatedAt)
				}

				// Apply tag filter
				if include && params.Tag != nil {
					hasTag := false
					for _, tag := range n.Tags {
						if strings.EqualFold(tag, *params.Tag) {
							hasTag = true
							break
						}
					}
					include = hasTag
				}

				if include {
					noteNum++
					filtered = append(filtered, noteSummaryFrom(n, noteNum))
				}
			}

			return ListNotesResult{
				Notes:   filtered,
				Count:   len(filtered),
				Message: fmt.Sprintf("Found %d notes", len(filtered)),
			}, nil
		},
	)
}

func (h *ToolHandler) searchNotesTool() copilot.Tool {
	return copilot.DefineTool(
		"search_notes",
		"Search notes by keyword in title or content. Returns numbered list for easy reference.",
		func(params SearchNotesParams, inv copilot.ToolInvocation) (SearchNotesResult, error) {
			noteList, err := h.storage.LoadNotes()
			if err != nil {
				return SearchNotesResult{Message: err.Error()}, nil
			}

			query := strings.ToLower(params.Query)
			filtered := make([]NoteSummary, 0, len(noteList.Notes))
			noteNum := noteNumberStart

			for _, n := range noteList.Notes {
				if strings.Contains(strings.ToLower(n.Title), query) ||
					strings.Contains(strings.ToLower(n.Content), query) {
					noteNum++
					filtered = append(filtered, noteSummaryFrom(n, noteNum))
				}
			}

			return SearchNotesResult{
				Notes:   filtered,
				Count:   len(filtered),
				Message: fmt.Sprintf("Found %d notes matching '%s'", len(filtered), params.Query),
			}, nil
		},
	)
}

func (h *ToolHandler) deleteNoteTool() copilot.Tool {
	return copilot.DefineTool(
		"delete_note",
		"Delete a note by ID or title match",
		func(params DeleteNoteParams, inv copilot.ToolInvocation) (DeleteNoteResult, error) {
			noteList, err := h.storage.LoadNotes()
			if err != nil {
				return DeleteNoteResult{Success: false, Message: err.Error()}, nil
			}

			foundIndex, matchedTitle := findNoteIndex(noteList.Notes, params.Query)
			if foundIndex == notFoundIndex {
				return DeleteNoteResult{
					Success: false,
					Message: fmt.Sprintf("No note found matching '%s'", params.Query),
				}, nil
			}

			noteList.Notes = append(noteList.Notes[:foundIndex], noteList.Notes[foundIndex+1:]...)

			if err := h.storage.SaveNotes(noteList); err != nil {
				return DeleteNoteResult{Success: false, Message: err.Error()}, nil
			}

			return DeleteNoteResult{
				Success: true,
				Message: fmt.Sprintf("Note '%s' deleted", matchedTitle),
			}, nil
		},
	)
}

func findTaskIndex(tasks []Task, query string) (int, string) {
	return findIndexByIDOrTitle(query, len(tasks), func(i int) (string, string) {
		return tasks[i].ID, tasks[i].Title
	})
}

func findNoteIndex(notes []Note, query string) (int, string) {
	return findIndexByIDOrTitle(query, len(notes), func(i int) (string, string) {
		return notes[i].ID, notes[i].Title
	})
}

func findIndexByIDOrTitle(query string, length int, accessor func(int) (string, string)) (int, string) {
	queryLower := strings.ToLower(query)
	for i := 0; i < length; i++ {
		id, title := accessor(i)
		if id == query || strings.Contains(strings.ToLower(title), queryLower) {
			return i, title
		}
	}
	return notFoundIndex, ""
}

func noteSummaryFrom(note Note, number int) NoteSummary {
	return NoteSummary{
		Number:  number,
		ID:      note.ID,
		Title:   note.Title,
		Preview: notePreview(note.Content),
		Tags:    note.Tags,
	}
}

func notePreview(content string) string {
	if len(content) > notePreviewMax {
		return content[:notePreviewMax] + "..."
	}
	return content
}
