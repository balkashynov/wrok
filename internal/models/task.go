package models

import (
	"time"
	"gorm.io/gorm"
)

// Task represents a todo item
type Task struct {
	ID         uint           `gorm:"primarykey" json:"id"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`
	
	Title      string     `gorm:"not null" json:"title"`
	Project    string     `json:"project"`
	Status     string     `gorm:"default:todo" json:"status"` // todo, done, archived
	Priority   int        `gorm:"default:0" json:"priority"`   // 0=no priority, 1=low, 2=medium, 3=high
	Pinned     bool       `gorm:"default:false" json:"pinned"`
	Due        *time.Time `json:"due"`
	DoneAt     *time.Time `json:"done_at"`
	ArchivedAt *time.Time `json:"archived_at"`
	
	// Optional metadata
	JiraID string `json:"jira_id"`
	URL    string `json:"url"`
	Note   string `json:"note"`
	
	// Relationships
	Tags     []Tag     `gorm:"many2many:task_tags;" json:"tags"`
	Sessions []Session `gorm:"foreignKey:TaskID" json:"sessions"`
}

// Tag represents a task tag
type Tag struct {
	ID   uint   `gorm:"primarykey" json:"id"`
	Name string `gorm:"unique;not null" json:"name"`
	
	// Relationships
	Tasks []Task `gorm:"many2many:task_tags;" json:"-"`
}

// TaskTag is the join table for the many-to-many relationship
type TaskTag struct {
	TaskID uint `gorm:"primaryKey"`
	TagID  uint `gorm:"primaryKey"`
}