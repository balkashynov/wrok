package tui

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
)

// RunAddTaskTUI starts the interactive add task TUI
func RunAddTaskTUI(prefilled map[string]string) error {
	model := NewAddTaskModel(prefilled)
	
	p := tea.NewProgram(model, tea.WithAltScreen())
	finalModel, err := p.Run()
	
	// Handle exit messages after TUI closes
	if err != nil {
		return err
	}
	
	if m, ok := finalModel.(AddTaskModel); ok {
		if m.cancelled {
			fmt.Println("❌ Task creation cancelled.")
		} else if m.completed && m.createdTaskID > 0 {
			fmt.Printf("✅ New task \"%s\" added - ID: %d\n", m.createdTaskTitle, m.createdTaskID)
		} else if m.err != nil {
			fmt.Printf("❌ Error: %v\n", m.err)
		}
	}
	
	return nil
}