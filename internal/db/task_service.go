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

// GetTasks retrieves tasks with optional filters
func GetTasks() ([]models.Task, error) {
	var tasks []models.Task
	
	// Preload tags relationship
	if err := DB.Preload("Tags").Find(&tasks).Error; err != nil {
		return nil, err
	}
	
	return tasks, nil
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