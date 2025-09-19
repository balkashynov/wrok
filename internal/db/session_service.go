package db

import (
	"fmt"
	"time"

	"github.com/balkashynov/wrok/internal/models"
)

// StartSession starts a new time tracking session for a task
func StartSession(taskID uint) (*models.Session, error) {
	// Check if task exists
	var task models.Task
	if err := DB.First(&task, taskID).Error; err != nil {
		return nil, fmt.Errorf("task #%d not found", taskID)
	}

	// Check if there's already an active session
	var activeSession models.Session
	err := DB.Where("finished_at IS NULL").First(&activeSession).Error
	if err == nil {
		// There's already an active session
		return nil, fmt.Errorf("session already active for task #%d. Stop it first with 'wrok stop'", activeSession.TaskID)
	}

	// Create new session
	session := models.Session{
		TaskID:    taskID,
		StartedAt: time.Now(),
	}

	if err := DB.Create(&session).Error; err != nil {
		return nil, err
	}

	// Load the task relationship
	DB.Preload("Task").First(&session, session.ID)

	return &session, nil
}

// StopActiveSession stops the currently active session
func StopActiveSession() (*models.Session, error) {
	var session models.Session
	
	// Find active session
	err := DB.Where("finished_at IS NULL").Preload("Task").First(&session).Error
	if err != nil {
		return nil, fmt.Errorf("no active session found")
	}

	// Stop the session
	now := time.Now()
	session.FinishedAt = &now
	session.DurationSeconds = int(now.Sub(session.StartedAt).Seconds())

	if err := DB.Save(&session).Error; err != nil {
		return nil, err
	}

	return &session, nil
}

// GetActiveSession returns the currently active session, if any
func GetActiveSession() (*models.Session, error) {
	var session models.Session
	
	err := DB.Where("finished_at IS NULL").Preload("Task").First(&session).Error
	if err != nil {
		return nil, nil // No active session is not an error
	}

	return &session, nil
}

// GetSessionsInRange returns all sessions within the specified date range
func GetSessionsInRange(startTime, endTime time.Time) ([]models.Session, error) {
	var sessions []models.Session

	err := DB.Where("started_at >= ? AND started_at <= ? AND finished_at IS NOT NULL", startTime, endTime).
		Preload("Task").
		Preload("Task.Tags").
		Order("started_at ASC").
		Find(&sessions).Error

	if err != nil {
		return nil, err
	}

	return sessions, nil
}

// GetTaskByID retrieves a task by ID
func GetTaskByID(id uint) (*models.Task, error) {
	var task models.Task
	
	err := DB.Preload("Tags").First(&task, id).Error
	if err != nil {
		return nil, fmt.Errorf("task #%d not found", id)
	}
	
	return &task, nil
}