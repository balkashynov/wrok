package commands

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/parser"
	"github.com/balkashynov/wrok/internal/tui"
)

var addCmd = &cobra.Command{
	Use:   "add [task description]",
	Short: "Add a new task",
	Long: `Add a new task with optional metadata.

Modes:
  Interactive: wrok add -i (or just 'wrok add' with no arguments)
  Quick: wrok add "Task title" (with optional flags)
  Smart parsing: wrok add "Fix bug #urgent @backend +high APP-42"

Smart parsing syntax:
  #tag1,tag2  - Tags (comma-separated or individual)
  @project    - Project name  
  +priority   - Priority (low/medium/high or 1/2/3)
  ABC-123     - JIRA ticket (auto-detected)
  due:3days   - Due date (dd/mm/yyyy, X days, X hours, X weeks)`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		initDB()
		interactive, _ := cmd.Flags().GetBool("interactive")
		
		// If no args and not explicitly interactive, go interactive
		if len(args) == 0 && !interactive {
			interactive = true
		}
		
		// If interactive mode or if there are parsing errors, use TUI
		if interactive {
			runInteractiveAdd(cmd, args)
		} else {
			// Check if we should parse the title for metadata
			title := strings.Join(args, " ")
			parsed := parser.ParseTitle(title)
			
			if len(parsed.Errors) > 0 {
				// There were parsing errors, fall back to interactive with pre-filled data
				fmt.Printf("⚠️  Found issues with parsing: %s\n", strings.Join(parsed.Errors, ", "))
				fmt.Println("Opening interactive mode for confirmation...")
				runInteractiveAddWithParsed(cmd, parsed)
			} else {
				// Direct creation with parsed or flag data
				runDirectAdd(cmd, parsed)
			}
		}
	},
}

// runInteractiveAdd starts interactive mode
func runInteractiveAdd(cmd *cobra.Command, args []string) {
	prefilled := make(map[string]string)
	
	// Pre-fill from arguments if provided
	if len(args) > 0 {
		prefilled["title"] = strings.Join(args, " ")
	}
	
	// Pre-fill from flags
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		prefilled["project"] = project
	}
	if tags, _ := cmd.Flags().GetStringSlice("tags"); len(tags) > 0 {
		prefilled["tags"] = strings.Join(tags, ", ")
	}
	if priority, _ := cmd.Flags().GetString("priority"); priority != "" {
		prefilled["priority"] = priority
	}
	if jira, _ := cmd.Flags().GetString("jira"); jira != "" {
		prefilled["jira"] = jira
	}
	if due, _ := cmd.Flags().GetString("due"); due != "" {
		prefilled["due_date"] = due
	}
	if note, _ := cmd.Flags().GetString("note"); note != "" {
		prefilled["notes"] = note
	}
	
	if err := tui.RunAddTaskTUI(prefilled); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// runInteractiveAddWithParsed starts interactive mode with parsed data
func runInteractiveAddWithParsed(cmd *cobra.Command, parsed parser.ParsedTask) {
	prefilled := make(map[string]string)
	prefilled["title"] = parsed.Title
	
	if parsed.Project != "" {
		prefilled["project"] = parsed.Project
	}
	if len(parsed.Tags) > 0 {
		prefilled["tags"] = strings.Join(parsed.Tags, ", ")
	}
	if parsed.Priority != "" {
		prefilled["priority"] = parsed.Priority
	}
	if parsed.JiraID != "" {
		prefilled["jira"] = parsed.JiraID
	}
	if parsed.DueDate != nil {
		// Convert time back to a readable format for the TUI input
		prefilled["due_date"] = parsed.DueDate.Format("02/01/2006")
	}
	
	// Override with any explicit flags
	if project, _ := cmd.Flags().GetString("project"); project != "" {
		prefilled["project"] = project
	}
	if tags, _ := cmd.Flags().GetStringSlice("tags"); len(tags) > 0 {
		prefilled["tags"] = strings.Join(tags, ", ")
	}
	if priority, _ := cmd.Flags().GetString("priority"); priority != "" {
		prefilled["priority"] = priority
	}
	if jira, _ := cmd.Flags().GetString("jira"); jira != "" {
		prefilled["jira"] = jira
	}
	if due, _ := cmd.Flags().GetString("due"); due != "" {
		prefilled["due_date"] = due
	}
	if note, _ := cmd.Flags().GetString("note"); note != "" {
		prefilled["notes"] = note
	}
	
	if err := tui.RunAddTaskTUI(prefilled); err != nil {
		fmt.Printf("Error: %v\n", err)
	}
}

// runDirectAdd creates task directly without TUI
func runDirectAdd(cmd *cobra.Command, parsed parser.ParsedTask) {
	// Start with parsed data
	title := parsed.Title
	project := parsed.Project
	tags := parsed.Tags
	priority := parsed.Priority
	jiraID := parsed.JiraID
	dueDate := parsed.DueDate
	
	// Override with explicit flags (flags take precedence)
	if flagProject, _ := cmd.Flags().GetString("project"); flagProject != "" {
		project = flagProject
	}
	if flagTags, _ := cmd.Flags().GetStringSlice("tags"); len(flagTags) > 0 {
		tags = flagTags
	}
	if flagPriority, _ := cmd.Flags().GetString("priority"); flagPriority != "" {
		priority = flagPriority
	}
	if flagJira, _ := cmd.Flags().GetString("jira"); flagJira != "" {
		jiraID = flagJira
	}
	if flagDueDate, _ := cmd.Flags().GetString("due"); flagDueDate != "" {
		parsedDueDate, err := parser.ParseDueDate(flagDueDate)
		if err != nil {
			fmt.Printf("Error parsing due date: %v\n", err)
			return
		}
		dueDate = parsedDueDate
	}
	
	url, _ := cmd.Flags().GetString("url")
	note, _ := cmd.Flags().GetString("note")
	
	// Create task request
	req := db.CreateTaskRequest{
		Title:    title,
		Project:  project,
		Tags:     tags,
		Priority: priority,
		JiraID:   jiraID,
		URL:      url,
		Note:     note,
		DueDate:  dueDate,
	}
	
	// Create the task
	task, err := db.CreateTask(req)
	if err != nil {
		fmt.Printf("Error creating task: %v\n", err)
		return
	}
	
	// Success message
	fmt.Printf("Created task #%d: %s\n", task.ID, task.Title)
	if task.Project != "" {
		fmt.Printf("  Project: %s\n", task.Project)
	}
	if len(task.Tags) > 0 {
		var tagNames []string
		for _, tag := range task.Tags {
			tagNames = append(tagNames, tag.Name)
		}
		fmt.Printf("  Tags: %s\n", strings.Join(tagNames, ", "))
	}
	if task.Priority > 0 {
		priorities := []string{"", "low", "medium", "high"}
		fmt.Printf("  Priority: %s\n", priorities[task.Priority])
	}
	if task.JiraID != "" {
		fmt.Printf("  JIRA: %s\n", task.JiraID)
	}
	if task.Due != nil {
		fmt.Printf("  Due: %s\n", parser.FormatDueDate(task.Due))
	}
}

func init() {
	// Add flags to the add command
	addCmd.Flags().BoolP("interactive", "i", false, "Interactive mode with TUI")
	addCmd.Flags().StringP("project", "p", "", "Project name")
	addCmd.Flags().StringSliceP("tags", "t", []string{}, "Comma-separated tags")
	addCmd.Flags().StringP("priority", "", "", "Priority: low, medium, high, or 1-3")
	addCmd.Flags().StringP("jira", "", "", "JIRA ticket ID")
	addCmd.Flags().StringP("due", "", "", "Due date: dd/mm/yyyy, X days, X hours, X weeks")
	addCmd.Flags().StringP("url", "", "", "Related URL")
	addCmd.Flags().StringP("note", "", "", "Additional notes")
}