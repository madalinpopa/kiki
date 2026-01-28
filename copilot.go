package main

import (
	"fmt"
	"strings"
	"time"

	copilot "github.com/github/copilot-sdk/go"
)

const systemPrompt = `You are Kiki, a sarcastic but helpful personal assistant. You manage tasks and notes while being delightfully judgy.

## Personality
- Sarcastic but never mean - playful roasts only
- Brief responses (1-3 sentences)
- Always helpful despite the attitude

## IMPORTANT: Always Use Your Tools
You have custom tools for task and note management. ALWAYS use these tools - never create files manually.

### Task Tools (stored in ~/.kiki/tasks.json)
- add_task: Create tasks with title, optional due_date (YYYY-MM-DD), priority (low/medium/high), tags
- list_tasks: List tasks with filter (all, today, incomplete, completed)
- complete_task: Mark task done by ID or title match
- delete_task: Remove task by ID or title match

### Note Tools (stored in ~/.kiki/notes.json)
- add_note: Create notes with title, content, optional tags
- list_notes: List notes with filter (all, today) and optional tag
- search_notes: Find notes by keyword in title or content
- delete_note: Remove note by ID or title match

## Examples
User: "add task to fix the login bug"
→ Call add_task with title="Fix the login bug"

User: "what tasks do I have today?"
→ Call list_tasks with filter="today"

User: "done with the bug fix"
→ Call complete_task with query="bug fix"

User: "note: API uses OAuth 2.0 for auth"
→ Call add_note with title="API Auth" content="API uses OAuth 2.0 for auth"

Today's date is %s.`

const sessionTimeout = 2 * time.Minute

// Kiki wraps the Copilot client for the CLI assistant
type Kiki struct {
	client  *copilot.Client
	storage *Storage
	tools   *ToolHandler
}

// NewKiki creates a new Kiki instance
func NewKiki(storage *Storage) (*Kiki, error) {
	client := copilot.NewClient(nil)
	if err := client.Start(); err != nil {
		return nil, fmt.Errorf("failed to start copilot client: %w", err)
	}

	tools := NewToolHandler(storage)

	return &Kiki{
		client:  client,
		storage: storage,
		tools:   tools,
	}, nil
}

// Close shuts down the Copilot client
func (k *Kiki) Close() {
	if k.client != nil {
		k.client.Stop()
	}
}

// getDailySessionID returns a session ID for today (one session per day)
func getDailySessionID() string {
	return fmt.Sprintf("kiki-%s", todayString())
}

// getOrCreateSession returns today's session, creating a new one if needed
func (k *Kiki) getOrCreateSession() (*copilot.Session, error) {
	sessionID := getDailySessionID()
	today := todayString()
	fullSystemPrompt := fmt.Sprintf(systemPrompt, today)

	// Try to resume existing session first
	session, err := k.client.ResumeSession(sessionID)
	if err == nil {
		return session, nil
	}

	// Session doesn't exist, create a new one with our custom ID
	session, err = k.client.CreateSession(&copilot.SessionConfig{
		SessionID: sessionID,
		Model:     "gpt-4.1",
		Streaming: true,
		Tools:     k.tools.GetAllTools(),
		SystemMessage: &copilot.SystemMessageConfig{
			Mode:    "append",
			Content: fullSystemPrompt,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

// todayString returns today's date as YYYY-MM-DD
func todayString() string {
	return time.Now().Format(dateLayout)
}

// Run sends a prompt to Kiki and returns the response
func (k *Kiki) Run(prompt string) (string, error) {
	session, err := k.getOrCreateSession()
	if err != nil {
		return "", err
	}
	defer session.Destroy()

	// Collect response
	var responseBuilder strings.Builder
	var sessionError error

	// Set up streaming handler
	session.On(func(event copilot.SessionEvent) {
		switch event.Type {
		case "assistant.message_delta":
			if event.Data.DeltaContent != nil {
				fmt.Print(*event.Data.DeltaContent)
				responseBuilder.WriteString(*event.Data.DeltaContent)
			}
		case "assistant.message":
			if event.Data.Content != nil && responseBuilder.Len() == 0 {
				fmt.Print(*event.Data.Content)
				responseBuilder.WriteString(*event.Data.Content)
			}
		case "session.idle":
			fmt.Println()
		case "session.error":
			if event.Data.Error != nil {
				sessionError = fmt.Errorf("session error: %v", *event.Data.Error)
			}
		}
	})

	// Send the prompt and wait for completion (2 minute timeout)
	_, err = session.SendAndWait(copilot.MessageOptions{Prompt: prompt}, sessionTimeout)
	if err != nil {
		if sessionError != nil {
			return "", sessionError
		}
		return "", fmt.Errorf("failed to send message: %w", err)
	}

	return responseBuilder.String(), nil
}
