package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/tui"
)

var editCmd = &cobra.Command{
	Use:   "edit <task_id>",
	Short: "Edit an existing task",
	Long: `Edit an existing task in interactive mode.

Opens the same interface as 'wrok add' but with all fields pre-populated
with the current task data. You can modify any field and save changes.

Usage:
  wrok edit 42    - Edit task with ID 42`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Parse task ID
		taskID, err := strconv.ParseUint(args[0], 10, 32)
		if err != nil {
			fmt.Printf("Error: Invalid task ID '%s'. Please provide a valid numeric ID.\n", args[0])
			return
		}

		// Fetch existing task
		task, err := db.GetTaskByID(uint(taskID))
		if err != nil {
			fmt.Printf("Error: Task #%d not found.\n", taskID)
			return
		}

		// Create prefilled data from existing task
		prefilled := make(map[string]string)
		prefilled["title"] = task.Title
		prefilled["project"] = task.Project
		prefilled["jira"] = task.JiraID
		prefilled["notes"] = task.Note

		// Convert priority to string
		if task.Priority > 0 {
			priorities := []string{"", "low", "medium", "high"}
			if int(task.Priority) < len(priorities) {
				prefilled["priority"] = priorities[task.Priority]
			}
		}

		// Convert tags to comma-separated string with # prefix for each tag
		if len(task.Tags) > 0 {
			var tagNames []string
			for _, tag := range task.Tags {
				tagNames = append(tagNames, "#"+tag.Name)
			}
			prefilled["tags"] = strings.Join(tagNames, ", ")
		}

		// Convert due date to string
		if task.Due != nil {
			prefilled["due_date"] = task.Due.Format("02/01/2006")
		}

		// Launch edit TUI
		if err := tui.RunEditTaskTUI(task.ID, prefilled); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	},
}

func init() {
	// Add flags if needed in the future
}