# 🐱 Kiki

A sarcastic but helpful CLI assistant for managing tasks and notes, powered by GitHub Copilot. Built with the GitHub
Copilot SDK ([github/copilot-sdk](https://github.com/github/copilot-sdk)).

Quick taste (with the `??` alias):

```console
➜ ?? add task: Investigate a critical deployment failure due Jan 30
Task "Investigate a critical deployment failure" added for January 30—because nothing says fun like a looming crisis.
At least now your task list has some drama again!

➜ ?? add task: clean up Docker images (low) by Jan 31
Task "Clean up Docker images" added for January 31 with low priority—because nothing says excitement like digital housekeeping.
Your to-do list is back in business, superstar!

➜ ?? add task: prepare high-priority infrastructure docs by Feb 2
Task "Prepare high-priority infrastructure docs" added for February 2 with high priority—because nothing says "living on the edge" like documentation deadlines.
Your to-do list is officially intimidating again!

➜ ?? list my tasks
Here’s your current to-do list—finally looking like a real adult’s schedule:

1. Investigate a critical deployment failure (Due: 2026-01-30, Priority: medium)
2. Clean up Docker images (Due: 2026-01-31, Priority: low)
3. Prepare high-priority infrastructure docs (Due: 2026-02-02, Priority: high)

Look at you, balancing chaos and chores like a pro!

➜ ?? mark task 2 done
Task "Clean up Docker images" is now marked as done—look at you, conquering low-priority chores like a true champion.
Want to tackle something with a little more adrenaline next?

➜ ?? list my tasks
Here’s your current to-do list—one less Docker mess to worry about:

1. Investigate a critical deployment failure (Due: 2026-01-30, Priority: medium)
2. Clean up Docker images (Due: 2026-01-31, Priority: low) — completed
3. Prepare high-priority infrastructure docs (Due: 2026-02-02, Priority: high)

Look at you, making progress like a productivity machine (with a hint of sass, of course)!

➜ ?? list all high pr tasks
Here’s your high-priority task—because you clearly love living on the edge:

1. Prepare high-priority infrastructure docs (Due: 2026-02-02, Priority: high)

Just one, but it’s got “high-priority” written all over it. No pressure!

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
