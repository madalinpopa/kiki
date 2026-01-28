# 🐱 Kiki

A sarcastic but helpful CLI assistant for managing tasks and notes, powered by GitHub Copilot.

## Features

- **Task Management** - Add, list, complete, and delete tasks with priorities and due dates
- **Note Taking** - Capture notes with tags and search through them
- **Natural Language** - Just tell Kiki what you want in plain English
- **Sarcastic Personality** - Get things done with a side of sass

## Installation

```bash
go install github.com/madalinpopa/kiki@latest
```

Or build from source:

```bash
git clone https://github.com/madalinpopa/kiki.git
cd kiki
go build .
```

## Setup

Initialize Kiki's configuration directory:

```bash
kiki init
```

This creates:

- `$XDG_CONFIG_HOME/kiki/tasks.json` (defaults to `~/.config/kiki/`)
- `$XDG_CONFIG_HOME/kiki/notes.json`

## Usage

Send prompts with the `-p` flag:

```bash
# Tasks
kiki -p "add task: fix the login bug"
kiki -p "add task: deploy to production tomorrow with high priority"
kiki -p "list my tasks"
kiki -p "what tasks do I have today?"
kiki -p "mark the login bug as done"
kiki -p "delete task 3"

# Notes
kiki -p "note: API uses OAuth 2.0 for authentication"
kiki -p "list my notes"
kiki -p "search notes for OAuth"
kiki -p "delete note about API"
```

## Tools

Kiki provides 8 tools for task and note management:

| Tool            | Description                                            |
|-----------------|--------------------------------------------------------|
| `add_task`      | Create a task with title, due date, priority, tags     |
| `list_tasks`    | List tasks (filter: all, today, incomplete, completed) |
| `complete_task` | Mark a task as done by ID, number, or title            |
| `delete_task`   | Remove a task by ID, number, or title                  |
| `add_note`      | Create a note with title, content, and tags            |
| `list_notes`    | List notes (filter: all, today, or by tag)             |
| `search_notes`  | Find notes by keyword in title or content              |
| `delete_note`   | Remove a note by ID, number, or title                  |

## Configuration

Kiki uses the XDG Base Directory specification:

```
$XDG_CONFIG_HOME/kiki/
├── tasks.json
└── notes.json
```

If `XDG_CONFIG_HOME` is not set, defaults to `~/.config/kiki/`.

## Requirements

- Go 1.21+
- GitHub Copilot subscription

## License

MIT
