package commands

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/models"
)

var jiraCmd = &cobra.Command{
	Use:   "jira",
	Short: "Show weekly timesheet for JIRA reporting",
	Long: `Show a weekly timesheet of tracked time grouped by day.

Displays hours worked per day for the current calendar week, similar to Tempo interface.
Only shows days where work was tracked.

Example output:
  Task                    Mon  Tue  Wed  Thu  Fri  Sat  Sun  Total
  APP-123 Fix login bug     2    3    1    -    -    -    -      6
  APP-456 Add new feature   -    1    2    4    1    -    -      8
  Total                     2    4    3    4    1    0    0     14`,
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		if err := generateJiraTimesheet(); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

// generateJiraTimesheet creates and displays the weekly timesheet
func generateJiraTimesheet() error {
	// Get current calendar week (Monday to Sunday)
	now := time.Now()
	weekStart := getWeekStart(now)
	weekEnd := weekStart.AddDate(0, 0, 7).Add(-time.Second) // End of Sunday

	// Get all sessions from this week
	sessions, err := db.GetSessionsInRange(weekStart, weekEnd)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	if len(sessions) == 0 {
		fmt.Println("No time tracked this week.")
		return nil
	}

	// Group sessions by task and day
	taskDayHours := make(map[string]map[time.Weekday]float64)
	allTasks := make(map[string]*models.Task)

	for _, session := range sessions {
		taskKey := formatTaskKey(session.Task)
		weekday := session.StartedAt.Weekday()

		if taskDayHours[taskKey] == nil {
			taskDayHours[taskKey] = make(map[time.Weekday]float64)
		}

		hours := float64(session.DurationSeconds) / 3600.0
		taskDayHours[taskKey][weekday] += hours
		allTasks[taskKey] = &session.Task
	}

	// Calculate which days have any work
	activeDays := make(map[time.Weekday]bool)
	for _, dayHours := range taskDayHours {
		for day, hours := range dayHours {
			if hours > 0 {
				activeDays[day] = true
			}
		}
	}

	// Generate and display the table
	displayTimesheet(taskDayHours, allTasks, activeDays, weekStart)

	return nil
}

// getWeekStart returns the start of the calendar week (Monday) for the given time
func getWeekStart(t time.Time) time.Time {
	weekday := t.Weekday()
	daysFromMonday := int(weekday - time.Monday)
	if weekday == time.Sunday {
		daysFromMonday = 6 // Sunday is 6 days from Monday
	}

	weekStart := t.AddDate(0, 0, -daysFromMonday)
	// Set to start of day
	return time.Date(weekStart.Year(), weekStart.Month(), weekStart.Day(), 0, 0, 0, 0, weekStart.Location())
}

// formatTaskKey creates a display key for the task
func formatTaskKey(task models.Task) string {
	if task.JiraID != "" {
		return fmt.Sprintf("%s %s", task.JiraID, task.Title)
	}
	return fmt.Sprintf("#%d %s", task.ID, task.Title)
}

// displayTimesheet outputs the formatted timesheet table
func displayTimesheet(taskDayHours map[string]map[time.Weekday]float64, allTasks map[string]*models.Task, activeDays map[time.Weekday]bool, weekStart time.Time) {
	// Sort tasks by JIRA ID first, then by task ID
	var taskKeys []string
	for taskKey := range taskDayHours {
		taskKeys = append(taskKeys, taskKey)
	}
	sort.Slice(taskKeys, func(i, j int) bool {
		task1 := allTasks[taskKeys[i]]
		task2 := allTasks[taskKeys[j]]

		// Sort by JIRA ID first if both have JIRA
		if task1.JiraID != "" && task2.JiraID != "" {
			return task1.JiraID < task2.JiraID
		}
		// JIRA tasks come before non-JIRA tasks
		if task1.JiraID != "" && task2.JiraID == "" {
			return true
		}
		if task1.JiraID == "" && task2.JiraID != "" {
			return false
		}
		// Both non-JIRA, sort by ID
		return task1.ID < task2.ID
	})

	// Determine which days to show (only active days)
	var daysToShow []time.Weekday
	dayNames := []string{"Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"}
	weekdays := []time.Weekday{time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday, time.Sunday}

	for i, weekday := range weekdays {
		if activeDays[weekday] {
			daysToShow = append(daysToShow, weekday)
		} else {
			// Also show if it's a weekday and we have some work this week
			if i < 5 && len(activeDays) > 0 { // Mon-Fri
				daysToShow = append(daysToShow, weekday)
			}
		}
	}

	// Calculate column widths
	maxTaskNameWidth := 20
	for _, taskKey := range taskKeys {
		if len(taskKey) > maxTaskNameWidth {
			maxTaskNameWidth = len(taskKey)
		}
	}
	if maxTaskNameWidth > 40 {
		maxTaskNameWidth = 40 // Cap at 40 chars
	}

	dayColumnWidth := 5
	totalColumnWidth := 7

	// Print header
	fmt.Printf("%-*s", maxTaskNameWidth, "Task")
	for _, weekday := range daysToShow {
		dayIndex := (int(weekday) - 1 + 7) % 7 // Convert to 0-6 with Monday=0
		fmt.Printf("  %*s", dayColumnWidth-2, dayNames[dayIndex])
	}
	fmt.Printf("  %*s\n", totalColumnWidth-2, "Total")

	// Print separator
	fmt.Print(strings.Repeat("-", maxTaskNameWidth))
	for range daysToShow {
		fmt.Print("  " + strings.Repeat("-", dayColumnWidth-2))
	}
	fmt.Print("  " + strings.Repeat("-", totalColumnWidth-2))
	fmt.Println()

	// Print task rows
	weekTotals := make(map[time.Weekday]float64)
	grandTotal := 0.0

	for _, taskKey := range taskKeys {
		dayHours := taskDayHours[taskKey]

		// Truncate task name if too long
		displayTaskKey := taskKey
		if len(displayTaskKey) > maxTaskNameWidth {
			displayTaskKey = displayTaskKey[:maxTaskNameWidth-3] + "..."
		}

		fmt.Printf("%-*s", maxTaskNameWidth, displayTaskKey)

		taskTotal := 0.0
		for _, weekday := range daysToShow {
			hours := dayHours[weekday]
			if hours > 0 {
				roundedHours := math.Ceil(hours)
				fmt.Printf("  %*d", dayColumnWidth-2, int(roundedHours))
				weekTotals[weekday] += roundedHours
				taskTotal += roundedHours
			} else {
				fmt.Printf("  %*s", dayColumnWidth-2, "-")
			}
		}

		fmt.Printf("  %*d\n", totalColumnWidth-2, int(taskTotal))
		grandTotal += taskTotal
	}

	// Print total row
	fmt.Print(strings.Repeat("-", maxTaskNameWidth))
	for range daysToShow {
		fmt.Print("  " + strings.Repeat("-", dayColumnWidth-2))
	}
	fmt.Print("  " + strings.Repeat("-", totalColumnWidth-2))
	fmt.Println()

	fmt.Printf("%-*s", maxTaskNameWidth, "Total")
	for _, weekday := range daysToShow {
		total := weekTotals[weekday]
		if total > 0 {
			fmt.Printf("  %*d", dayColumnWidth-2, int(total))
		} else {
			fmt.Printf("  %*s", dayColumnWidth-2, "0")
		}
	}
	fmt.Printf("  %*d\n", totalColumnWidth-2, int(grandTotal))

	// Print week info
	fmt.Printf("\nWeek of %s to %s\n",
		weekStart.Format("Jan 2"),
		weekStart.AddDate(0, 0, 6).Format("Jan 2, 2006"))
}