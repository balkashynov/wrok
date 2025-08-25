package commands

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/models"
	"github.com/balkashynov/wrok/internal/tui"
)

var listCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List tasks",
	Long:    "List tasks with optional filters and formats. Opens interactive TUI by default, use --no-ui for simple output.",
	Run: func(cmd *cobra.Command, args []string) {
		// Get flags
		noUI, _ := cmd.Flags().GetBool("no-ui")
		jsonOutput, _ := cmd.Flags().GetBool("json")
		status, _ := cmd.Flags().GetString("status")
		project, _ := cmd.Flags().GetString("project")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		
		// Build query options
		opts := db.TaskQueryOptions{
			Status:  status,
			Project: project,
			Tags:    tags,
			OrderBy: "id DESC", // newest first by default
		}
		
		// Get tasks with filtering
		tasks, err := db.GetTasksWithOptions(opts)
		if err != nil {
			fmt.Printf("Error fetching tasks: %v\n", err)
			return
		}
		
		if len(tasks) == 0 {
			if noUI || jsonOutput {
				if jsonOutput {
					fmt.Println("[]")
				} else {
					fmt.Println("No tasks found.")
				}
			} else {
				fmt.Println("No tasks found. Use 'wrok add \"task description\"' to create your first task.")
			}
			return
		}
		
		// Check if we should open TUI or use non-UI mode
		if noUI || jsonOutput {
			if jsonOutput {
				renderJSON(tasks)
			} else {
				renderTable(tasks)
			}
		} else {
			// Launch interactive TUI
			runInteractiveList(tasks)
		}
	},
}

// renderJSON outputs tasks as JSON
func renderJSON(tasks []models.Task) {
	// Create a simplified structure for JSON output
	type JsonTask struct {
		ID       uint      `json:"id"`
		Title    string    `json:"title"`
		Status   string    `json:"status"`
		Project  string    `json:"project"`
		Priority string    `json:"priority"`
		JiraID   string    `json:"jira_id,omitempty"`
		Due      *time.Time `json:"due,omitempty"`
		Tags     []string  `json:"tags"`
		Notes    string    `json:"notes,omitempty"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}
	
	var jsonTasks []JsonTask
	for _, task := range tasks {
		// Get tag names
		var tagNames []string
		for _, tag := range task.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		
		// Priority display
		priorities := []string{"", "low", "medium", "high"}
		priorityStr := ""
		if task.Priority > 0 && task.Priority < len(priorities) {
			priorityStr = priorities[task.Priority]
		}
		
		jsonTask := JsonTask{
			ID:        task.ID,
			Title:     task.Title,
			Status:    task.Status,
			Project:   task.Project,
			Priority:  priorityStr,
			JiraID:    task.JiraID,
			Due:       task.Due,
			Tags:      tagNames,
			Notes:     task.Note,
			CreatedAt: task.CreatedAt,
			UpdatedAt: task.UpdatedAt,
		}
		jsonTasks = append(jsonTasks, jsonTask)
	}
	
	jsonBytes, err := json.MarshalIndent(jsonTasks, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonBytes))
}

// renderTable outputs tasks as a formatted table optimized for 80-column terminals
func renderTable(tasks []models.Task) {
	// Fixed column widths for 80-character terminals
	// Format: ID(4) TITLE(35) PROJECT(12) PRIORITY(8) TAGS(12) STATUS(6) = 77 chars + separators
	fmt.Printf("%-4s %-35s %-12s %-8s %-12s %s\n", "ID", "TITLE", "PROJECT", "PRIORITY", "TAGS", "STATUS")
	fmt.Println(strings.Repeat("-", 80))
	
	// Print each task
	for _, task := range tasks {
		// Get tag names
		var tagNames []string
		for _, tag := range task.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		tagsStr := strings.Join(tagNames, ",")
		
		// Priority display
		priorities := []string{"", "low", "med", "high"}
		priorityStr := ""
		if task.Priority > 0 && task.Priority < len(priorities) {
			priorityStr = priorities[task.Priority]
		}
		
		// Truncate long fields
		title := task.Title
		if len(title) > 33 {
			title = title[:30] + "..."
		}
		
		project := task.Project
		if len(project) > 10 {
			project = project[:7] + "..."
		}
		
		if len(tagsStr) > 10 {
			tagsStr = tagsStr[:7] + "..."
		}
		
		fmt.Printf("%-4d %-35s %-12s %-8s %-12s %s\n", 
			task.ID, 
			title, 
			project, 
			priorityStr, 
			tagsStr,
			task.Status)
	}
}

// runInteractiveList launches the TUI for task listing
func runInteractiveList(tasks []models.Task) {
	model := tui.NewListModel(tasks)
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running TUI: %v\n", err)
		return
	}
}

func init() {
	listCmd.Flags().Bool("no-ui", false, "Disable interactive TUI, output plain table")
	listCmd.Flags().Bool("json", false, "Output as JSON")
	listCmd.Flags().StringP("status", "s", "", "Filter by status: todo, done, archived")
	listCmd.Flags().StringP("project", "p", "", "Filter by project")
	listCmd.Flags().StringSliceP("tags", "t", []string{}, "Filter by tags (comma-separated)")
}