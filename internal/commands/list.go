package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
)

var listCmd = &cobra.Command{
	Use:     "ls",
	Aliases: []string{"list"},
	Short:   "List tasks",
	Long:    "List tasks with optional filters for status, project, tags, and date ranges",
	Run: func(cmd *cobra.Command, args []string) {
		// Get all tasks
		tasks, err := db.GetTasks()
		if err != nil {
			fmt.Printf("Error fetching tasks: %v\n", err)
			return
		}
		
		if len(tasks) == 0 {
			fmt.Println("No tasks found. Use 'wrok add \"task description\"' to create your first task.")
			return
		}
		
		// Print table header
		fmt.Printf("%-4s %-6s %-40s %-15s %-8s %s\n", "ID", "STATUS", "TITLE", "PROJECT", "PRIORITY", "TAGS")
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
			priorityStr := priorities[task.Priority]
			
			// Truncate title if too long
			title := task.Title
			if len(title) > 38 {
				title = title[:35] + "..."
			}
			
			// Truncate project if too long
			project := task.Project
			if len(project) > 13 {
				project = project[:10] + "..."
			}
			
			fmt.Printf("%-4d %-6s %-40s %-15s %-8s %s\n", 
				task.ID, 
				task.Status, 
				title, 
				project, 
				priorityStr, 
				tagsStr)
		}
	},
}

func init() {
	listCmd.Flags().StringP("status", "s", "", "Filter by status: todo, done, archived")
	listCmd.Flags().StringP("project", "p", "", "Filter by project")
	listCmd.Flags().BoolP("today", "", false, "Show only today's tasks")
}