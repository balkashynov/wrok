package commands

import (
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/tui"
)

var startCmd = &cobra.Command{
	Use:   "start [task-id]",
	Short: "Start tracking time on a task",
	Long: `Start tracking time on a task. Opens interactive timer by default, use --no-ui for simple start.

Examples:
  wrok start 42        # Start timer with interactive UI
  wrok start 42 --no-ui # Start timer without UI`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		taskID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Printf("Error: invalid task ID '%s'\n", args[0])
			return
		}

		session, err := db.StartSession(uint(taskID))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		// Check if --no-ui flag is set
		noUI, _ := cmd.Flags().GetBool("no-ui")
		if noUI {
			// Simple non-interactive start
			fmt.Printf("⏱️  Started tracking time for task #%d: %s\n", session.TaskID, session.Task.Title)
			fmt.Printf("Started at: %s\n", session.StartedAt.Format("15:04:05"))
		} else {
			// Interactive timer UI
			if err := tui.RunTimerTUI(session); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	},
}

var stopCmd = &cobra.Command{
	Use:   "stop",
	Short: "Stop tracking time",
	Run: func(cmd *cobra.Command, args []string) {
		session, err := db.StopActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		duration := time.Duration(session.DurationSeconds) * time.Second
		fmt.Printf("⏹️  Stopped tracking time for task #%d: %s\n", session.TaskID, session.Task.Title)
		fmt.Printf("Session duration: %s\n", formatDuration(duration))
	},
}

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current time tracking status",
	Run: func(cmd *cobra.Command, args []string) {
		session, err := db.GetActiveSession()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			return
		}

		if session == nil {
			fmt.Println("No active time tracking session")
			return
		}

		elapsed := time.Since(session.StartedAt)
		fmt.Printf("⏱️  Currently tracking: task #%d: %s\n", session.TaskID, session.Task.Title)
		fmt.Printf("Started at: %s\n", session.StartedAt.Format("15:04:05"))
		fmt.Printf("Elapsed time: %s\n", formatDuration(elapsed))
	},
}

func init() {
	// Add --no-ui flag to start command
	startCmd.Flags().Bool("no-ui", false, "Start timer without interactive UI")
}

// formatDuration formats a duration in a human-readable way
func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
}