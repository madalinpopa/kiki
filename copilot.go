package main

import (
    _ "embed"
    "fmt"
    "log/slog"
    "strings"
    "time"

    copilot "github.com/github/copilot-sdk/go"
)

//go:embed system_prompt.txt
var systemPromptTemplate string

const sessionTimeout = 2 * time.Minute

// Kiki wraps the Copilot client for the CLI assistant
type Kiki struct {
    client  *copilot.Client
    storage *Storage
    tools   *ToolHandler
    logger  *slog.Logger
    model   string
}

// NewKiki creates a new Kiki instance
func NewKiki(storage *Storage, logger *slog.Logger, model string) (*Kiki, error) {
    if model == "" {
        model = defaultModel
    }
    client := copilot.NewClient(nil)
    if err := client.Start(); err != nil {
        return nil, fmt.Errorf("failed to start copilot client: %w", err)
    }

    tools := NewToolHandler(storage, logger)

    return &Kiki{
        client:  client,
        storage: storage,
        tools:   tools,
        logger:  logger,
        model:   model,
    }, nil
}

// Close shuts down the Copilot client
func (k *Kiki) Close() {
    if k.client != nil {
        k.client.Stop()
    }
}

// RefreshSession deletes today's session so a new one can be created.
func (k *Kiki) RefreshSession() (bool, error) {
    sessionID := getDailySessionID()
    sessions, err := k.client.ListSessions()
    if err != nil {
        return false, fmt.Errorf("listing sessions: %w", err)
    }

    found := false
    for _, session := range sessions {
        if session.SessionID == sessionID {
            found = true
            break
        }
    }

    if found {
        if err := k.client.DeleteSession(sessionID); err != nil {
            return false, fmt.Errorf("deleting session: %w", err)
        }
    }

    return found, nil
}

// getDailySessionID returns a session ID for today (one session per day)
func getDailySessionID() string {
    return fmt.Sprintf("kiki-%s", todayString())
}

// getOrCreateSession returns today's session, creating a new one if needed
func (k *Kiki) getOrCreateSession() (*copilot.Session, error) {
    sessionID := getDailySessionID()
    today := todayString()
    fullSystemPrompt := fmt.Sprintf(systemPromptTemplate, today)
    tools := k.tools.GetAllTools()

    // Try to resume existing session first
    session, err := k.client.ResumeSessionWithOptions(sessionID, &copilot.ResumeSessionConfig{
        Tools:     tools,
        Streaming: true,
    })
    if err == nil {
        return session, nil
    }

    // Session doesn't exist, create a new one with our custom ID
    session, err = k.client.CreateSession(&copilot.SessionConfig{
        SessionID: sessionID,
        Model:     k.model,
        Streaming: true,
        Tools:     tools,
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
    defer func(session *copilot.Session) {
        err := session.Destroy()
        if err != nil {
            k.logger.Error("failed to close session")
        }
    }(session)

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
                k.logger.Error("session error", "error", sessionError)
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
