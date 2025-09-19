# wrok - CLI Todo + Time Tracker

## Project Overview
A cross-platform command-line tool written in Go that combines lightweight todo management with time tracking. All data stored locally in SQLite (`~/.wrok/wrok.db`).

## Core Functionality
- Task management with metadata (tags, project, priority, JIRA links, due dates)
- Time tracking with start/stop sessions and live timer UI
- Smart syntax parsing for inline metadata (#tags, @project, +priority, JIRA-123, due:+5d)
- Interactive Bubble Tea TUI with search, sort, and quick actions
- Weekly timesheet generation for JIRA/Tempo reporting
- Task lifecycle management (archive/unarchive, done/undone)

## Key Commands
- `wrok add "task #tag @project +priority JIRA-123 due:+5d" [--no-ui]`
- `wrok ls [--status=todo|done|archived] [--project=name] [--today|--week] [--no-ui|--json]`
- `wrok start <id> [--no-ui]` / `wrok stop` / `wrok status`
- `wrok done <id>` / `wrok undone <id>` - mark complete/incomplete (auto-stop if active)
- `wrok edit <id> [--no-ui]` - edit task with TUI or CLI
- `wrok archive <id> [--older-than=30d]` / `wrok unarchive <id>`
- `wrok search <query> [--exact|--prefix|--suffix|--fuzzy]`
- `wrok jira` - generate weekly timesheet (Tempo format)
- `wrok help` - comprehensive help with examples

## Interactive TUI Features
### Main List UI (`wrok ls`)
- Responsive design with shimmer effects and live status updates
- Search (`/`), sort (`f`), navigation (`↑/↓`)
- Quick actions: edit (`e`), timer (`s`), done (`d`), archive (`a`)
- Split-panel view with task details on right
- Live elapsed time display for running tasks

### Timer UI (`wrok start <id>`)
- Split-screen layout: ASCII clock + task details
- Real-time elapsed time counter with animated display
- Responsive design (collapses on narrow screens <90px)
- ESC/Q to return to main list UI

### Add/Edit UI
- Smart syntax input with auto-completion hints
- Metadata preview and validation
- Form-based editing for all task properties

## Smart Syntax
Parse metadata directly from task titles:
- `#hashtags` → Auto-create tags
- `@project` → Set project name
- `+priority` → Set priority (low/medium/high)
- `ABC-123` → Link JIRA ticket
- `due:+5d` → Set due date (5 days from now)
- `due:friday` → Set due date to next Friday
- `due:31/12/2024` → Set specific due date

Example: `wrok add "Fix login bug #frontend @auth +high ABC-123 due:+2d"`

## Data Model (SQLite)
### tasks table
- id, title, project, status (todo/done/archived), priority, created_at, updated_at
- due_date, done_at, archived_at, jira_id, note
- Tags stored in separate many-to-many relationship

### sessions table
- id, task_id, started_at, finished_at, duration_seconds
- Constraint: only one active session at a time

### tags table + task_tags junction
- Normalized tag storage with many-to-many relationship

## Technical Implementation
- **Framework**: Go with Cobra CLI + Bubble Tea TUI
- **Database**: SQLite with GORM ORM
- **UI**: Lipgloss styling with responsive design
- **Animation**: Time-based shimmer effects and live updates
- **Architecture**: Clean separation of commands, TUI models, database services

## Current Status (Implemented)
✅ Complete task management (CRUD operations)
✅ Smart syntax parsing with inline metadata
✅ Interactive TUI with search, sort, navigation
✅ Time tracking with live timer UI and session management
✅ Weekly timesheet generation (JIRA/Tempo format)
✅ Comprehensive help system with examples
✅ Archive/unarchive functionality
✅ Live status updates and elapsed time display
✅ Responsive design for various terminal sizes

## Build/Package
- Single static binary for Linux, macOS, Windows
- GoReleaser for releases
- Distribution via GitHub Releases, Homebrew, Scoop

## Future Enhancements
- Data export (CSV/JSON)
- Pomodoro mode integration
- Git integration for commit linking
- Config file support for defaults
- Bulk operations and advanced filtering