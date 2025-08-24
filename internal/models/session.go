package models

import (
	"time"
	"gorm.io/gorm"
)

// Session represents a time tracking session
type Session struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
	
	TaskID         uint       `gorm:"not null" json:"task_id"`
	StartedAt      time.Time  `gorm:"not null" json:"started_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	DurationSeconds int       `json:"duration_seconds"` // calculated field
	Note           string     `json:"note"`
	
	// Relationships
	Task Task `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"task"`
}