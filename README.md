# 🐱 Kiki

A sarcastic but helpful CLI assistant for managing tasks and notes, powered by GitHub Copilot. Built with the GitHub
Copilot SDK ([github/copilot-sdk](https://github.com/github/copilot-sdk)).

Quick taste (with the `??` alias):

```bash
?? add task: stop doomscrolling tomorrow high priority
# Kiki: Sure. High priority: touch grass at 9am. You’re welcome.

?? list my tasks
# Kiki: Here’s your to-do list. Try not to ignore it like last time.
# 1. Touch grass (due: 2025-01-29, priority: high)
# 2. Fix the login bug (due: 2025-01-30, priority: medium)
# 3. Review the PR (priority: low)

?? note: API uses OAuth 2.0 for auth
# Kiki: Noted. Try not to forget it like the rest of your “important” notes.

?? what tasks do I have today?
# Kiki: You have 2 tasks today. Try not to nap through both.
```

## Features

- **Task Management** - Add, list, complete, and delete tasks with priorities and due dates
- **Note-Taking** - Capture notes with tags and search through them
- **Natural Language** - Just tell Kiki what you want in plain English
- **Sarcastic Personality** - Get things done with a side of sass
- **Session Persistence** - Kiki maintains one Copilot session per day, remembering conversation context throughout the
  day

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

# Model selection
kiki --model gpt-4.1 -p "add task: review the PR"
```

## How I use it

Add this to your `~/.zshrc`:

```zsh
function __kiki_ask() {
  if [ -z "$1" ]; then
    kiki --help
    return 1
  fi
  kiki -p "$*" --model gpt-5-mini
}

alias '??'='noglob __kiki_ask'
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

## System Prompt

Kiki's system prompt lives in `system_prompt.txt` and is embedded into the binary at build time.

To customize it:

1. Edit `system_prompt.txt` (keep the `%s` placeholder for today's date).
2. Rebuild Kiki (`go build .` or `go install`).
3. Refresh the session so the new prompt is used:

```bash
kiki refresh
```

Note: Kiki uses one Copilot session per day. If you update the prompt mid-day, run `kiki refresh` to force a fresh
session with your changes.

## Configuration

Kiki uses the XDG Base Directory specification:

```
$XDG_CONFIG_HOME/kiki/
├── tasks.json
└── notes.json
```

If `XDG_CONFIG_HOME` is not set, defaults to `~/.config/kiki/`.

Logs are written to:

```
~/.kiki/kiki.log
```

## Requirements

- Go 1.21+
- GitHub Copilot subscription
