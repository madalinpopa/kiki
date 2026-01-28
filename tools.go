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

			var filtered []TaskSummary
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

			if filtered == nil {
				filtered = []TaskSummary{}
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

			query := strings.ToLower(params.Query)
			found := false
			var matchedTitle string

			for i := range taskList.Tasks {
				if taskList.Tasks[i].ID == params.Query ||
					strings.Contains(strings.ToLower(taskList.Tasks[i].Title), query) {
					taskList.Tasks[i].Completed = true
					taskList.Tasks[i].UpdatedAt = time.Now()
					matchedTitle = taskList.Tasks[i].Title
					found = true
					break
				}
			}

			if !found {
				return CompleteTaskResult{
					Success: false,
					Message: fmt.Sprintf("No task found matching '%s'", params.Query),
				}, nil
			}

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

			query := strings.ToLower(params.Query)
			found := notFoundIndex
			var matchedTitle string

			for i := range taskList.Tasks {
				if taskList.Tasks[i].ID == params.Query ||
					strings.Contains(strings.ToLower(taskList.Tasks[i].Title), query) {
					found = i
					matchedTitle = taskList.Tasks[i].Title
					break
				}
			}

			if found == notFoundIndex {
				return DeleteTaskResult{
					Success: false,
					Message: fmt.Sprintf("No task found matching '%s'", params.Query),
				}, nil
			}

			taskList.Tasks = append(taskList.Tasks[:found], taskList.Tasks[found+1:]...)

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

			var filtered []NoteSummary
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
					preview := n.Content
					if len(preview) > notePreviewMax {
						preview = preview[:notePreviewMax] + "..."
					}
					filtered = append(filtered, NoteSummary{
						Number:  noteNum,
						ID:      n.ID,
						Title:   n.Title,
						Preview: preview,
						Tags:    n.Tags,
					})
				}
			}

			if filtered == nil {
				filtered = []NoteSummary{}
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
			var filtered []NoteSummary
			noteNum := noteNumberStart

			for _, n := range noteList.Notes {
				if strings.Contains(strings.ToLower(n.Title), query) ||
					strings.Contains(strings.ToLower(n.Content), query) {
					noteNum++
					preview := n.Content
					if len(preview) > notePreviewMax {
						preview = preview[:notePreviewMax] + "..."
					}
					filtered = append(filtered, NoteSummary{
						Number:  noteNum,
						ID:      n.ID,
						Title:   n.Title,
						Preview: preview,
						Tags:    n.Tags,
					})
				}
			}

			if filtered == nil {
				filtered = []NoteSummary{}
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

			query := strings.ToLower(params.Query)
			found := notFoundIndex
			var matchedTitle string

			for i := range noteList.Notes {
				if noteList.Notes[i].ID == params.Query ||
					strings.Contains(strings.ToLower(noteList.Notes[i].Title), query) {
					found = i
					matchedTitle = noteList.Notes[i].Title
					break
				}
			}

			if found == notFoundIndex {
				return DeleteNoteResult{
					Success: false,
					Message: fmt.Sprintf("No note found matching '%s'", params.Query),
				}, nil
			}

			noteList.Notes = append(noteList.Notes[:found], noteList.Notes[found+1:]...)

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
