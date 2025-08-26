package db

import (
	"fmt"
	"strings"
	"time"

	"github.com/balkashynov/wrok/internal/models"
	"github.com/balkashynov/wrok/internal/parser"
)

// CreateTaskRequest holds the data needed to create a new task
type CreateTaskRequest struct {
	Title    string
	Project  string
	Tags     []string
	Priority string // can be "low/medium/high" or "1/2/3" or empty for no priority
	JiraID   string
	URL      string
	Note     string
	DueDate  *time.Time
}

// CreateTask creates a new task with tags
func CreateTask(req CreateTaskRequest) (*models.Task, error) {
	// Parse priority (optional)
	priority := parsePriority(req.Priority)
	
	// Normalize JIRA ID if provided and valid format
	normalizedJiraID := ""
	if req.JiraID != "" {
		if parser.IsValidJiraFormat(req.JiraID) {
			// Only normalize valid JIRA IDs
			normalized, _ := parser.NormalizeJiraID(req.JiraID)
			normalizedJiraID = normalized
		} else {
			// Keep invalid JIRA IDs as-is without uppercasing
			normalizedJiraID = req.JiraID
		}
	}
	
	// Create the task
	task := models.Task{
		Title:    req.Title,
		Project:  req.Project,
		Status:   "todo",
		Priority: priority,
		JiraID:   normalizedJiraID,
		URL:      req.URL,
		Note:     req.Note,
		Due:      req.DueDate,
	}

	// Process tags
	if len(req.Tags) > 0 {
		tags, err := findOrCreateTags(req.Tags)
		if err != nil {
			return nil, err
		}
		task.Tags = tags
	}

	// Save task to database
	if err := DB.Create(&task).Error; err != nil {
		return nil, err
	}

	return &task, nil
}

// parsePriority converts priority string to int
func parsePriority(priority string) int {
	priority = strings.ToLower(strings.TrimSpace(priority))
	if priority == "" {
		return 0 // 0 means no priority set
	}
	switch priority {
	case "low", "1":
		return 1
	case "medium", "2":
		return 2
	case "high", "3":
		return 3
	default:
		return 0 // invalid priority defaults to no priority
	}
}

// findOrCreateTags finds existing tags or creates new ones
func findOrCreateTags(tagNames []string) ([]models.Tag, error) {
	var tags []models.Tag
	
	for _, name := range tagNames {
		name = strings.TrimSpace(name)
		if name == "" {
			continue
		}
		
		var tag models.Tag
		
		// Try to find existing tag
		err := DB.Where("name = ?", name).First(&tag).Error
		if err != nil {
			// Tag doesn't exist, create it
			tag = models.Tag{Name: name}
			if err := DB.Create(&tag).Error; err != nil {
				return nil, err
			}
		}
		
		tags = append(tags, tag)
	}
	
	return tags, nil
}

// TaskQueryOptions holds options for querying tasks
type TaskQueryOptions struct {
	Status    string   // Filter by status
	Project   string   // Filter by project
	Tags      []string // Filter by tags (AND logic)
	JiraID    string   // Filter by JIRA ID
	Priority  string   // Filter by priority (low/medium/high)
	OrderBy   string   // Order by clause (e.g., "id DESC", "created_at ASC")
	Limit     int      // Limit results
	Offset    int      // Offset for pagination
}

// GetTasks retrieves all tasks (legacy function for backward compatibility)
func GetTasks() ([]models.Task, error) {
	opts := TaskQueryOptions{
		OrderBy: "id DESC", // Default order
	}
	return GetTasksWithOptions(opts)
}

// GetTasksWithOptions retrieves tasks with filtering and sorting options
func GetTasksWithOptions(opts TaskQueryOptions) ([]models.Task, error) {
	var tasks []models.Task
	
	// Start with base query, preload tags
	query := DB.Preload("Tags")
	
	// Apply filters
	if opts.Status != "" {
		query = query.Where("status = ?", opts.Status)
	}
	
	if opts.Project != "" {
		query = query.Where("project LIKE ?", "%"+opts.Project+"%")
	}
	
	if opts.JiraID != "" {
		query = query.Where("jira_id LIKE ?", "%"+opts.JiraID+"%")
	}
	
	if opts.Priority != "" {
		// Convert priority string to int
		var priorityInt int
		switch strings.ToLower(opts.Priority) {
		case "low", "1":
			priorityInt = 1
		case "medium", "med", "2":
			priorityInt = 2
		case "high", "3":
			priorityInt = 3
		default:
			priorityInt = 0 // No priority
		}
		query = query.Where("priority = ?", priorityInt)
	}
	
	// Filter by tags (AND logic - task must have all specified tags)
	if len(opts.Tags) > 0 {
		// Use subquery to find tasks that have all specified tags
		for _, tag := range opts.Tags {
			query = query.Where("id IN (?)", 
				DB.Table("task_tags").
					Select("task_id").
					Joins("JOIN tags ON task_tags.tag_id = tags.id").
					Where("tags.name LIKE ?", "%"+tag+"%"))
		}
	}
	
	// Apply ordering
	if opts.OrderBy != "" {
		query = query.Order(opts.OrderBy)
	}
	
	// Apply pagination
	if opts.Limit > 0 {
		query = query.Limit(opts.Limit)
	}
	if opts.Offset > 0 {
		query = query.Offset(opts.Offset)
	}
	
	// Execute query
	if err := query.Find(&tasks).Error; err != nil {
		return nil, err
	}
	
	return tasks, nil
}


// GetActiveTask returns the currently active (tracking time) task
func GetActiveTask() (*models.Task, error) {
	// Find the task with an active session
	var session models.Session
	if err := DB.Where("end_time IS NULL").First(&session).Error; err != nil {
		if err.Error() == "record not found" {
			return nil, nil // No active task
		}
		return nil, err
	}
	
	// Get the task for this session
	return GetTaskByID(session.TaskID)
}

// MarkTaskDone marks a task as completed and stops any active session
func MarkTaskDone(taskID uint) (*models.Task, error) {
	// Get the task
	task, err := GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	
	if task.Status == "done" {
		return nil, fmt.Errorf("task #%d is already completed", taskID)
	}
	
	// Check if there's an active session for this task and stop it
	activeSession, err := GetActiveSession()
	if err == nil && activeSession != nil && activeSession.TaskID == taskID {
		_, err = StopActiveSession()
		if err != nil {
			return nil, fmt.Errorf("failed to stop active session: %w", err)
		}
	}
	
	// Update task status
	now := time.Now()
	task.Status = "done"
	task.DoneAt = &now
	
	if err := DB.Save(task).Error; err != nil {
		return nil, err
	}
	
	return task, nil
}

// ArchiveTask marks a task as archived and stops any active session
func ArchiveTask(taskID uint) (*models.Task, error) {
	// Get the task
	task, err := GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	
	if task.Status == "archived" {
		return nil, fmt.Errorf("task #%d is already archived", taskID)
	}
	
	// Check if there's an active session for this task and stop it
	activeSession, err := GetActiveSession()
	if err == nil && activeSession != nil && activeSession.TaskID == taskID {
		_, err = StopActiveSession()
		if err != nil {
			return nil, fmt.Errorf("failed to stop active session: %w", err)
		}
	}
	
	// Update task status
	now := time.Now()
	task.Status = "archived"
	task.ArchivedAt = &now
	
	if err := DB.Save(task).Error; err != nil {
		return nil, err
	}
	
	return task, nil
}

// UnarchiveTask moves an archived task back to todo status
func UnarchiveTask(taskID uint) (*models.Task, error) {
	// Get the task
	task, err := GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	
	if task.Status != "archived" {
		return nil, fmt.Errorf("task #%d is not archived", taskID)
	}
	
	// Update task status back to todo
	task.Status = "todo"
	task.ArchivedAt = nil // Clear archived timestamp
	
	if err := DB.Save(task).Error; err != nil {
		return nil, err
	}
	
	return task, nil
}

// MarkTaskUndone moves a done task back to todo status
func MarkTaskUndone(taskID uint) (*models.Task, error) {
	// Get the task
	task, err := GetTaskByID(taskID)
	if err != nil {
		return nil, err
	}
	
	if task.Status != "done" {
		return nil, fmt.Errorf("task #%d is not completed", taskID)
	}
	
	// Update task status back to todo
	task.Status = "todo"
	task.DoneAt = nil // Clear done timestamp
	
	if err := DB.Save(task).Error; err != nil {
		return nil, err
	}
	
	return task, nil
}