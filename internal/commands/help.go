package commands

import (
	"fmt"

	"github.com/spf13/cobra"
)

var helpCmd = &cobra.Command{
	Use:   "help",
	Short: "Show comprehensive help for wrok",
	Long:  `Display detailed help for all wrok commands and flags.`,
	Run: func(cmd *cobra.Command, args []string) {
		showCustomHelp()
	},
}

func showCustomHelp() {
	fmt.Print(`
██╗    ██╗██████╗  ██████╗ ██╗  ██╗
██║    ██║██╔══██╗██╔═══██╗██║ ██╔╝
██║ █╗ ██║██████╔╝██║   ██║█████╔╝
██║███╗██║██╔══██╗██║   ██║██╔═██╗
╚███╔███╔╝██║  ██║╚██████╔╝██║  ██╗
 ╚══╝╚══╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝

wrok - CLI Todo + Time Tracker

COMMANDS:

  add <task>              Create a new task with smart parsing
    -p, --project         Set project name
    -t, --tags            Comma-separated tags
    --jira                JIRA ticket ID
    --prio                Priority: low|medium|high
    --due                 Due date (dd/mm/yyyy, +5d, +2h)
    --note                Additional notes
    --no-ui               Skip interactive TUI

    Smart syntax:
      #hashtags     Auto-create tags
      @project      Set project
      +priority     Set priority (low/medium/high)
      ABC-123       Link JIRA ticket
      due:+5d       Set due date (5 days from now)

    Example:
      wrok add "Fix login bug #frontend @auth +high ABC-123 due:+2d"

  ls                      List and manage tasks with interactive UI
    --status              Filter by status: todo|done|archived
    --project             Filter by project name
    --tags                Filter by tags (comma-separated)
    --today               Show only today's tasks
    --week                Show this week's tasks
    --no-ui               Simple text output
    --json                JSON output

    Quick actions:
      ↑/↓           Navigate tasks
      /             Search
      f             Sort
      e             Edit selected task
      s             Start/stop timer
      d             Mark done/undone
      a             Archive/unarchive
      esc/q         Quit

  edit <id>               Edit an existing task
    --no-ui               Edit via command line

  done <id>               Mark task as completed
  undone <id>             Mark task as todo

  archive <id>            Archive completed task
    --older-than          Archive tasks older than duration
  unarchive <id>          Restore archived task

  search <query>          Search tasks by title or content
    --exact               Exact match
    --prefix              Prefix match
    --suffix              Suffix match
    --fuzzy               Fuzzy search (default)

  start <id>              Start tracking time on a task
    --no-ui               Start without interactive timer
  stop                    Stop current time tracking session
  status                  Show current tracking status

  jira                    Generate weekly timesheet for JIRA reporting
  help                    Show this help

Use --no-ui flag with any command for CLI-only mode.

`)
}

type helpSection struct {
	title    string
	commands []helpCommand
}

type helpCommand struct {
	name        string
	description string
	examples    []string
	flags       []helpFlag
}

type helpFlag struct {
	name        string
	description string
}