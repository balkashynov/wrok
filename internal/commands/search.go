package commands

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/models"
)

var searchCmd = &cobra.Command{
	Use:   "search [query]",
	Short: "Search tasks across all fields",
	Long: `Search tasks with comprehensive matching:
- Exact match (highest priority)
- Prefix match 
- Suffix match
- Fuzzy match (contains, lowest priority)

Search is case insensitive and searches across title, project, tags, JIRA ID, notes, status, and priority.`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		query := args[0]
		
		// Get flags from list command (reuse the same filtering options)
		status, _ := cmd.Flags().GetString("status")
		project, _ := cmd.Flags().GetString("project")
		tags, _ := cmd.Flags().GetStringSlice("tags")
		priority, _ := cmd.Flags().GetString("priority")
		jiraID, _ := cmd.Flags().GetString("jira")
		orderBy, _ := cmd.Flags().GetString("order")
		limit, _ := cmd.Flags().GetInt("limit")
		jsonOutput, _ := cmd.Flags().GetBool("json")

		// Build query options
		opts := db.TaskQueryOptions{
			Status:   status,
			Project:  project,
			Tags:     tags,
			JiraID:   jiraID,
			Priority: priority,
			OrderBy:  orderBy,
			Limit:    limit,
		}
		
		if orderBy == "" {
			opts.OrderBy = "id DESC" // Default order
		}

		// Perform search
		tasks, err := db.SearchTasks(query, opts)
		if err != nil {
			fmt.Printf("Error searching tasks: %v\n", err)
			os.Exit(1)
		}

		// Output results
		if jsonOutput {
			renderSearchJSON(tasks, query)
		} else {
			renderSearchTable(tasks, query)
		}
	},
}

// renderSearchJSON outputs search results as JSON
func renderSearchJSON(tasks []models.Task, query string) {
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
	
	type SearchResult struct {
		Query   string     `json:"query"`
		Count   int        `json:"count"`
		Tasks   []JsonTask `json:"tasks"`
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
	
	result := SearchResult{
		Query: query,
		Count: len(tasks),
		Tasks: jsonTasks,
	}
	
	jsonBytes, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Printf("Error marshaling JSON: %v\n", err)
		return
	}
	
	fmt.Println(string(jsonBytes))
}

// renderSearchTable outputs search results as a formatted table
func renderSearchTable(tasks []models.Task, query string) {
	fmt.Printf("Search results for '%s' (%d found):\n", query, len(tasks))
	if len(tasks) == 0 {
		fmt.Println("No tasks found matching your search.")
		return
	}
	
	fmt.Println()
	
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

func init() {
	// Add the same flags as list command
	searchCmd.Flags().StringP("status", "s", "", "Filter by status (todo/done/archived)")
	searchCmd.Flags().StringP("project", "p", "", "Filter by project")
	searchCmd.Flags().StringSliceP("tags", "t", []string{}, "Filter by tags")
	searchCmd.Flags().StringP("priority", "", "", "Filter by priority (low/medium/high)")
	searchCmd.Flags().StringP("jira", "j", "", "Filter by JIRA ID")
	searchCmd.Flags().StringP("order", "o", "", "Order by (e.g., 'id DESC', 'created_at ASC')")
	searchCmd.Flags().IntP("limit", "l", 0, "Limit number of results")
	searchCmd.Flags().Bool("json", false, "Output as JSON")
}