# wrok

```
██╗    ██╗██████╗  ██████╗ ██╗  ██╗
██║    ██║██╔══██╗██╔═══██╗██║ ██╔╝
██║ █╗ ██║██████╔╝██║   ██║█████╔╝
██║███╗██║██╔══██╗██║   ██║██╔═██╗
╚███╔███╔╝██║  ██║╚██████╔╝██║  ██╗
 ╚══╝╚══╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝
```

A powerful command-line tool that combines lightweight todo management with time tracking. All data stored locally in SQLite.

## Features

- **Task Management**: Create, edit, and organize tasks with metadata (tags, projects, priority, JIRA links)
- **Time Tracking**: Start/stop timers with live session tracking and interactive UI
- **Smart Syntax**: Parse metadata directly from task titles using hashtags, @projects, +priority, and more
- **Interactive TUI**: Beautiful terminal interface with search, sort, and quick actions
- **Reporting**: Generate weekly timesheets for JIRA/Tempo integration
- **Local Storage**: All data stored in SQLite database (`~/.wrok/wrok.db`)

## Installation

### From Releases (Recommended)
Download the latest binary from [GitHub Releases](https://github.com/balkashynov/wrok/releases):

```bash
# Linux/macOS
curl -L https://github.com/balkashynov/wrok/releases/latest/download/wrok-linux-amd64 -o wrok
chmod +x wrok
sudo mv wrok /usr/local/bin/

# Or for macOS with Homebrew
brew install balkashynov/tap/wrok
```

### From Source
```bash
git clone https://github.com/balkashynov/wrok.git
cd wrok
go build -o wrok ./cmd/wrok
```

## Quick Start

```bash
# Create your first task
wrok add "Fix login bug #frontend @auth +high ABC-123 due:+2d"

# List tasks (interactive UI)
wrok ls

# Start working on a task
wrok start 1

# Stop timer
wrok stop

# Generate weekly timesheet
wrok jira
```

## Usage

### Task Management

**Create tasks with smart syntax:**
```bash
wrok add "Task title #tag1 #tag2 @project +priority ABC-123 due:+5d"
```

Smart syntax elements:
- `#hashtags` → Auto-create tags
- `@project` → Set project name
- `+priority` → Set priority (low/medium/high)
- `ABC-123` → Link JIRA ticket
- `due:+5d` → Set due date (5 days from now)

**List and manage tasks:**
```bash
wrok ls                    # Interactive UI
wrok ls --no-ui            # Simple text output
wrok ls --status done      # Filter by status
wrok ls --project backend  # Filter by project
wrok ls --today            # Today's tasks only
```

**Task lifecycle:**
```bash
wrok done 42        # Mark complete
wrok undone 42      # Mark as todo
wrok archive 42     # Archive task
wrok edit 42        # Edit task
```

### Time Tracking

**Start/stop sessions:**
```bash
wrok start 42       # Start timer (interactive UI)
wrok start 42 --no-ui  # Start without UI
wrok stop           # Stop current session
wrok status         # Show current status
```

**Generate reports:**
```bash
wrok jira           # Weekly timesheet (Tempo format)
```

### Interactive UI

The `wrok ls` command launches an interactive terminal UI with these controls:

- `↑/↓` - Navigate tasks
- `/` - Search tasks
- `f` - Sort options
- `e` - Edit selected task
- `s` - Start/stop timer
- `d` - Mark done/undone
- `a` - Archive/unarchive
- `esc/q` - Quit

## Examples

### Creating Tasks
```bash
# Simple task
wrok add "Review pull request"

# Task with metadata using flags
wrok add "Implement OAuth" -p auth -t security,backend --prio high

# Task with smart syntax (equivalent to above)
wrok add "Implement OAuth @auth #security #backend +high"

# Task with JIRA ticket and due date
wrok add "Fix database migration @backend +medium DEV-456 due:friday"
```

### Time Tracking Workflow
```bash
# Start working on task #5
wrok start 5

# Check current status
wrok status
# Output: ⏱ Working on #5 "Fix login bug" (1h 23m)

# Stop when done
wrok stop
# Output: ✅ Stopped timer for #5 "Fix login bug" (1h 45m total)

# View weekly report
wrok jira
# Output: Formatted timesheet table
```

### Filtering and Search
```bash
# Show only high priority tasks
wrok ls --no-ui | grep "+high"

# Search for specific tasks
wrok search "login"

# Show work for this week
wrok ls --week

# Filter by project and status
wrok ls --project frontend --status todo --no-ui
```

## Data Storage

All data is stored locally in SQLite:
- Database location: `~/.wrok/wrok.db`
- No cloud sync or external dependencies
- Export data: `sqlite3 ~/.wrok/wrok.db .dump`

## Development

### Building
```bash
go build -o wrok ./cmd/wrok
```

### Testing
```bash
go test ./...
```

### Database Schema
- `tasks`: id, title, project, tags, status, priority, jira_id, due_date, created_at, etc.
- `sessions`: id, task_id, started_at, finished_at, duration_seconds

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

MIT License - see [LICENSE](LICENSE) file for details.