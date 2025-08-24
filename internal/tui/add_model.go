package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/parser"
)

// createBigText converts text to big block letters (simplified version)
func createBigText(text string) string {
	if len(text) == 0 {
		return "No title..."
	}
	
	// Limit length to prevent overflow
	if len(text) > 10 {
		text = text[:7] + "..."
	}
	
	text = strings.ToUpper(text)
	
	// Simple 3-line block font
	line1 := ""
	line2 := ""
	line3 := ""
	
	for _, char := range text {
		switch char {
		case 'A':
			line1 += "███ "
			line2 += "█▀█ "
			line3 += "█▄█ "
		case 'B':
			line1 += "██▄ "
			line2 += "██▄ "
			line3 += "██▀ "
		case 'C':
			line1 += "▄██ "
			line2 += "█   "
			line3 += "▀██ "
		case 'D':
			line1 += "██▄ "
			line2 += "█ █ "
			line3 += "██▀ "
		case 'E':
			line1 += "███ "
			line2 += "██▄ "
			line3 += "███ "
		case 'F':
			line1 += "███ "
			line2 += "██▄ "
			line3 += "█   "
		case 'G':
			line1 += "▄██ "
			line2 += "█▄█ "
			line3 += "▀██ "
		case 'H':
			line1 += "█ █ "
			line2 += "███ "
			line3 += "█ █ "
		case 'I':
			line1 += "███ "
			line2 += " █  "
			line3 += "███ "
		case 'J':
			line1 += "███ "
			line2 += "  █ "
			line3 += "██▀ "
		case 'K':
			line1 += "█ █ "
			line2 += "██  "
			line3 += "█ █ "
		case 'L':
			line1 += "█   "
			line2 += "█   "
			line3 += "███ "
		case 'M':
			line1 += "█▄█ "
			line2 += "█▀█ "
			line3 += "█ █ "
		case 'N':
			line1 += "█▄█ "
			line2 += "█▀█ "
			line3 += "█ █ "
		case 'O':
			line1 += "▄█▄ "
			line2 += "█ █ "
			line3 += "▀█▀ "
		case 'P':
			line1 += "██▄ "
			line2 += "██▀ "
			line3 += "█   "
		case 'Q':
			line1 += "▄█▄ "
			line2 += "█ █ "
			line3 += "▀██ "
		case 'R':
			line1 += "██▄ "
			line2 += "██▄ "
			line3 += "█ █ "
		case 'S':
			line1 += "▄██ "
			line2 += "▀█▄ "
			line3 += "██▀ "
		case 'T':
			line1 += "███ "
			line2 += " █  "
			line3 += " █  "
		case 'U':
			line1 += "█ █ "
			line2 += "█ █ "
			line3 += "▀█▀ "
		case 'V':
			line1 += "█ █ "
			line2 += "█ █ "
			line3 += " █  "
		case 'W':
			line1 += "█ █ "
			line2 += "█▀█ "
			line3 += "█▄█ "
		case 'X':
			line1 += "█ █ "
			line2 += " █  "
			line3 += "█ █ "
		case 'Y':
			line1 += "█ █ "
			line2 += " █  "
			line3 += " █  "
		case 'Z':
			line1 += "███ "
			line2 += " █  "
			line3 += "███ "
		case ' ':
			line1 += "  "
			line2 += "  "
			line3 += "  "
		case '.':
			line1 += "   "
			line2 += "   "
			line3 += "▄  "
		default:
			// For unsupported chars, use a simple block
			line1 += "█ "
			line2 += "█ "
			line3 += "█ "
		}
	}
	
	return fmt.Sprintf("%s\n%s\n%s", line1, line2, line3)
}


// Step represents the current step in the wizard
type Step int

const (
	StepTitle Step = iota
	StepProject
	StepTags
	StepPriority
	StepJira
	StepDueDate
	StepNotes
	StepComplete
)

// AddTaskModel represents the TUI model for adding tasks
type AddTaskModel struct {
	currentStep Step
	inputs      []textinput.Model
	width       int
	height      int
	
	// Task data
	title     string
	project   string
	tags      []string
	priority  string
	jiraID    string
	dueDate   string
	notes     string
	
	// Pre-filled data from flags or parsing
	prefilled map[string]string
	
	// State
	err           error
	completed     bool
	cancelled     bool
	validationErr string
	createdTaskID uint
	createdTaskTitle string
	
	// Tag input state
	isAddingTags bool
}

// NewAddTaskModel creates a new add task TUI model
func NewAddTaskModel(prefilled map[string]string) AddTaskModel {
	inputs := make([]textinput.Model, 7)
	
	// Title input
	inputs[0] = textinput.New()
	inputs[0].Placeholder = "Enter task title... (required)"
	inputs[0].Focus()
	inputs[0].CharLimit = 200
	inputs[0].Width = 60 // Explicit width control
	
	// Project input
	inputs[1] = textinput.New()
	inputs[1].Placeholder = "Project name (Enter to skip)"
	inputs[1].CharLimit = 50
	inputs[1].Width = 60
	
	// Tags input (we'll handle this specially)
	inputs[2] = textinput.New()
	inputs[2].Placeholder = "Add tag (Enter to skip, 'q' when done adding tags)"
	inputs[2].CharLimit = 50
	inputs[2].Width = 60
	
	// Priority input
	inputs[3] = textinput.New()
	inputs[3].Placeholder = "low/medium/high or 1/2/3 (Enter to skip - no priority)"
	inputs[3].CharLimit = 10
	inputs[3].Width = 60
	
	// JIRA input
	inputs[4] = textinput.New()
	inputs[4].Placeholder = "JIRA ticket like APP-42 (Enter to skip)"
	inputs[4].CharLimit = 20
	inputs[4].Width = 60
	
	// Due date input
	inputs[5] = textinput.New()
	inputs[5].Placeholder = "Due: dd/mm/yyyy, 3 days, 24 hours, 2 weeks (Enter to skip)"
	inputs[5].CharLimit = 50
	inputs[5].Width = 60
	
	// Notes input
	inputs[6] = textinput.New()
	inputs[6].Placeholder = "Additional notes (Enter to skip)"
	inputs[6].CharLimit = 500
	inputs[6].Width = 60

	m := AddTaskModel{
		currentStep: StepTitle,
		inputs:      inputs,
		prefilled:   prefilled,
		tags:        []string{},
	}
	
	// Set pre-filled values
	if title, ok := prefilled["title"]; ok {
		m.inputs[0].SetValue(title)
		m.title = title
	}
	if project, ok := prefilled["project"]; ok {
		m.inputs[1].SetValue(project)
		m.project = project
	}
	if tags, ok := prefilled["tags"]; ok {
		m.inputs[2].SetValue(tags)
	}
	if priority, ok := prefilled["priority"]; ok {
		m.inputs[3].SetValue(priority)
		m.priority = priority
	}
	if jira, ok := prefilled["jira"]; ok {
		m.inputs[4].SetValue(jira)
		m.jiraID = jira
	}
	if dueDate, ok := prefilled["due_date"]; ok {
		m.inputs[5].SetValue(dueDate)
		m.dueDate = dueDate
	}
	if notes, ok := prefilled["notes"]; ok {
		m.inputs[6].SetValue(notes)
		m.notes = notes
	}

	return m
}

// Init initializes the model
func (m AddTaskModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles messages
func (m AddTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Update input field widths based on available space
		maxInputWidth := (m.width * 2 / 3) - 10 // Left panel width minus borders/padding
		if maxInputWidth < 30 {
			maxInputWidth = 30
		}
		if maxInputWidth > 80 {
			maxInputWidth = 80
		}
		
		for i := range m.inputs {
			m.inputs[i].Width = maxInputWidth
		}
		
		return m, nil
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.cancelled = true
			return m, tea.Quit
			
		case "enter":
			return m.handleEnter()
			
		case "tab", "down":
			// Don't allow skipping required title field
			if m.currentStep == StepTitle && strings.TrimSpace(m.title) == "" {
				m.validationErr = "Task title is required"
				return m, nil
			}
			return m.nextStep()
			
		case "shift+tab", "up":
			return m.prevStep()
		}
	}
	
	// Update the current input
	var cmd tea.Cmd
	m.inputs[m.currentStep], cmd = m.inputs[m.currentStep].Update(msg)
	
	// Update the corresponding field
	m.updateCurrentField()
	
	return m, cmd
}

// View renders the TUI
func (m AddTaskModel) View() string {
	if m.cancelled {
		return ""  // Don't show anything, let TUI handle exit message
	}
	
	if m.completed {
		return ""  // Don't show anything, let TUI handle exit message
	}
	
	// Split screen styles - left takes 2/3, right takes 1/3
	// Add some margin between panels
	leftWidth := (m.width * 2 / 3) - 3
	rightWidth := (m.width / 3) - 3
	
	// Ensure minimum widths
	if leftWidth < 40 {
		leftWidth = 40
	}
	if rightWidth < 25 {
		rightWidth = 25
	}
	
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1)
		
	rightStyle := lipgloss.NewStyle().
		Width(rightWidth).
		Height(m.height - 2).
		Padding(1)
	
	// Left side: Step-by-step wizard
	left := m.renderWizard()
	
	// Right side: Live preview
	right := m.renderPreview()
	
	// Render each panel separately
	leftPanel := leftStyle.Render(left)
	rightPanel := rightStyle.Render(right)
	
	// Combine with explicit spacing
	return lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		" ", // Add explicit separator
		rightPanel,
	)
}

// renderWizard renders the step-by-step wizard
func (m AddTaskModel) renderWizard() string {
	var b strings.Builder
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("62")).
		MarginBottom(1)
	b.WriteString(titleStyle.Render("📝 Create New Task"))
	b.WriteString("\n\n")
	
	// Current step indicator
	stepLabels := []string{"Title", "Project", "Tags", "Priority", "JIRA", "Due Date", "Notes"}
	for i, label := range stepLabels {
		if Step(i) == m.currentStep {
			b.WriteString(fmt.Sprintf("▶ %s\n", label))
		} else if Step(i) < m.currentStep {
			b.WriteString(fmt.Sprintf("✓ %s\n", label))
		} else {
			b.WriteString(fmt.Sprintf("  %s\n", label))
		}
	}
	b.WriteString("\n")
	
	// Current input field
	switch m.currentStep {
	case StepTitle:
		b.WriteString("📋 Task Title\n")
		b.WriteString(m.inputs[0].View())
		
	case StepProject:
		b.WriteString("📁 Project\n") 
		b.WriteString(m.inputs[1].View())
		
	case StepTags:
		b.WriteString("🔖 Tags\n")
		if len(m.tags) > 0 {
			b.WriteString(fmt.Sprintf("Added: %s\n", strings.Join(m.tags, ", ")))
		}
		b.WriteString(m.inputs[2].View())
		
	case StepPriority:
		b.WriteString("⚡ Priority\n")
		b.WriteString(m.inputs[3].View())
		
	case StepJira:
		b.WriteString("🎫 JIRA Ticket\n")
		b.WriteString(m.inputs[4].View())
		
	case StepDueDate:
		b.WriteString("📅 Due Date\n")
		b.WriteString(m.inputs[5].View())
		
	case StepNotes:
		b.WriteString("📝 Notes\n")
		b.WriteString(m.inputs[6].View())
	}
	
	// Show validation error if any
	if m.validationErr != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("196")). // Red
			Bold(true).
			MarginTop(1)
		b.WriteString(errorStyle.Render("❌ " + m.validationErr))
	}
	
	b.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("240")).
		Italic(true)
	b.WriteString(helpStyle.Render("Enter: Next | Tab/↓: Next | Shift+Tab/↑: Back | Esc: Cancel"))
	
	return b.String()
}

// renderPreview renders the live task preview
func (m AddTaskModel) renderPreview() string {
	var b strings.Builder
	
	// Calculate available width and height for the right panel
	rightPanelWidth := (m.width / 3) - 3
	if rightPanelWidth < 30 {
		rightPanelWidth = 30
	}
	
	// Calculate vertical centering
	availableHeight := m.height - 8
	if availableHeight < 1 {
		availableHeight = 1
	}

	// Card dimensions
	cardWidth := 36
	if rightPanelWidth > 45 {
		cardWidth = 42
	}

	// Add vertical spacing for centering
	verticalPadding := (availableHeight - 18) / 2 // Approximate card height
	if verticalPadding < 0 {
		verticalPadding = 0
	}
	
	for i := 0; i < verticalPadding; i++ {
		b.WriteString("\n")
	}

	// Build card content first
	var cardContent strings.Builder
	
	// WROK ASCII Logo with spinning O
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")). // Darker purple instead of light pink
		Bold(true).
		Align(lipgloss.Center)
	
	logo := `
  ██╗    ██╗██████╗  ██████╗ ██╗  ██╗
  ██║    ██║██╔══██╗██╔═══██╗██║ ██╔╝
  ██║ █╗ ██║██████╔╝██║   ██║█████╔╝ 
  ██║███╗██║██╔══██╗██║   ██║██╔═██╗ 
  ╚███╔███╔╝██║  ██║╚██████╔╝██║  ██╗
   ╚══╝╚══╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝`
	
	cardContent.WriteString(logoStyle.Render(logo))
	cardContent.WriteString("\n")
	
	// Separator line
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("99")). // Same darker purple
		Align(lipgloss.Center)
	cardContent.WriteString(separatorStyle.Render("═══════════════════════════════════"))
	cardContent.WriteString("\n")
	
	// Title section with BIG text
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("255")). // White
		Align(lipgloss.Center)
		
	// Add emoji first
	cardContent.WriteString(titleStyle.Render("🎯"))
	cardContent.WriteString("\n")
	
	// Then add the big title text
	var titleText string
	if m.title != "" {
		titleText = m.title
	} else {
		titleText = "TASK"
	}
	
	bigTitle := createBigText(titleText)
	cardContent.WriteString(titleStyle.Render(bigTitle))
	cardContent.WriteString("\n")
	
	// Status with nice styling
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("46")). // Green
		Background(lipgloss.Color("22")). // Dark green bg
		Bold(true).
		Padding(0, 1).
		Align(lipgloss.Center)
	cardContent.WriteString(statusStyle.Render("● TODO"))
	cardContent.WriteString("\n")
	
	// Separator (reuse animated style)
	cardContent.WriteString(separatorStyle.Render("─────────────────────────────────────"))
	cardContent.WriteString("\n")
	
	// Metadata section with proper styling
	metadataStyle := lipgloss.NewStyle().
		Align(lipgloss.Left).
		Padding(0, 1)
	
	var metadata strings.Builder
	
	// Project
	if m.project != "" {
		metadata.WriteString(fmt.Sprintf("📁 Project: %s\n", m.project))
	}
	
	// Tags with purple styling
	if len(m.tags) > 0 {
		tagStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("135")).
			Bold(true)
		var styledTags []string
		for _, tag := range m.tags {
			styledTags = append(styledTags, tagStyle.Render("#"+tag))
		}
		metadata.WriteString(fmt.Sprintf("🔖 Tags: %s\n", strings.Join(styledTags, " ")))
	}
	
	// Priority with visual indicators
	if m.priority != "" {
		normalizedPriority := parser.NormalizePriority(m.priority)
		var priorityDisplay string
		
		switch normalizedPriority {
		case "high":
			priorityDisplay = lipgloss.NewStyle().
				Foreground(lipgloss.Color("196")).
				Bold(true).
				Render("🔥 HIGH")
		case "medium":
			priorityDisplay = lipgloss.NewStyle().
				Foreground(lipgloss.Color("214")).
				Bold(true).
				Render("⚡ MEDIUM")
		case "low":
			priorityDisplay = lipgloss.NewStyle().
				Foreground(lipgloss.Color("46")).
				Bold(true).
				Render("🟢 LOW")
		}
		metadata.WriteString(fmt.Sprintf("⚡ Priority: %s\n", priorityDisplay))
	}
	
	// JIRA with validation
	if m.jiraID != "" {
		displayJira := m.jiraID
		warningText := ""
		
		if parser.IsValidJiraFormat(m.jiraID) {
			normalized, _ := parser.NormalizeJiraID(m.jiraID)
			displayJira = normalized
		} else {
			warningText = " ⚠️ Expected XXX-123"
		}
		
		jiraStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("33")).
			Background(lipgloss.Color("17")).
			Bold(true).
			Padding(0, 1)
		
		metadata.WriteString(fmt.Sprintf("🎫 JIRA: %s%s\n", jiraStyle.Render(displayJira), warningText))
	}
	
	// Due date
	if m.dueDate != "" {
		parsedDate, err := parser.ParseDueDate(m.dueDate)
		if err == nil && parsedDate != nil {
			dueDateDisplay := parser.FormatDueDate(parsedDate)
			metadata.WriteString(fmt.Sprintf("%s\n", dueDateDisplay))
		} else {
			metadata.WriteString(fmt.Sprintf("📅 Due: %s\n", m.dueDate))
		}
	}
	
	// Notes
	if m.notes != "" {
		noteStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("180")).
			Italic(true)
		metadata.WriteString(fmt.Sprintf("📝 Notes: %s\n", noteStyle.Render(m.notes)))
	}
	
	// Add metadata to card
	cardContent.WriteString(metadataStyle.Render(metadata.String()))
	
	// Create the card with static purple border
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("99")). // Darker purple border
		Width(cardWidth).
		Padding(1).
		Align(lipgloss.Center)
	
	// Center the card within the right panel
	cardContainer := lipgloss.NewStyle().
		Width(rightPanelWidth).
		Align(lipgloss.Center)
	
	card := cardStyle.Render(cardContent.String())
	b.WriteString(cardContainer.Render(card))
	
	return b.String()
}

// handleEnter processes the Enter key
func (m AddTaskModel) handleEnter() (AddTaskModel, tea.Cmd) {
	m.validationErr = "" // Clear any previous validation error
	
	// Handle different steps
	switch m.currentStep {
	case StepTitle:
		if strings.TrimSpace(m.title) == "" {
			m.validationErr = "Task title is required"
			return m, nil
		}
		return m.nextStep()
		
	case StepProject:
		// Project is optional, just move on
		return m.nextStep()
		
	case StepTags:
		// Handle tag input
		currentTag := strings.TrimSpace(m.inputs[2].Value())
		if currentTag == "q" || currentTag == "Q" {
			// User wants to stop adding tags
			return m.nextStep()
		} else if currentTag == "" {
			if len(m.tags) == 0 {
				// No tags added, move to next step
				return m.nextStep()
			} else {
				// Already have tags, ask for more or skip
				return m.nextStep()
			}
		} else {
			// Add the tag and clear input for next tag
			m.tags = append(m.tags, currentTag)
			m.inputs[2].SetValue("")
			m.inputs[2].Placeholder = fmt.Sprintf("Add another tag (%d added so far, Enter to finish, 'q' to stop)", len(m.tags))
			return m, nil
		}
		
	case StepPriority:
		// Priority is optional
		priorityInput := strings.TrimSpace(m.inputs[3].Value())
		if priorityInput == "" {
			// No priority set
			m.priority = ""
			return m.nextStep()
		} else {
			// Validate the priority
			normalizedInput := strings.ToLower(priorityInput)
			if normalizedInput != "low" && normalizedInput != "medium" && normalizedInput != "med" && normalizedInput != "high" &&
			   normalizedInput != "1" && normalizedInput != "2" && normalizedInput != "3" {
				m.validationErr = "Invalid priority. Use: low, medium, high, 1, 2, or 3"
				return m, nil
			}
			m.priority = priorityInput
			return m.nextStep()
		}
		
	case StepJira:
		// JIRA is optional, just move on
		return m.nextStep()
		
	case StepDueDate:
		// Due date is optional, validate if provided
		dueDateInput := strings.TrimSpace(m.inputs[5].Value())
		if dueDateInput == "" {
			// No due date
			m.dueDate = ""
			return m.nextStep()
		} else {
			// Validate due date format
			_, err := parser.ParseDueDate(dueDateInput)
			if err != nil {
				m.validationErr = "Invalid due date: " + err.Error()
				return m, nil
			}
			m.dueDate = dueDateInput
			return m.nextStep()
		}
		
	case StepNotes:
		// Complete the task creation
		return m.createTask()
	}
	
	return m, nil
}

// nextStep moves to the next step
func (m AddTaskModel) nextStep() (AddTaskModel, tea.Cmd) {
	if m.currentStep < StepNotes {
		m.inputs[m.currentStep].Blur()
		m.currentStep++
		m.inputs[m.currentStep].Focus()
	}
	return m, textinput.Blink
}

// prevStep moves to the previous step
func (m AddTaskModel) prevStep() (AddTaskModel, tea.Cmd) {
	if m.currentStep > StepTitle {
		m.inputs[m.currentStep].Blur()
		m.currentStep--
		m.inputs[m.currentStep].Focus()
	}
	return m, textinput.Blink
}

// updateCurrentField updates the model field based on current input
func (m *AddTaskModel) updateCurrentField() {
	switch m.currentStep {
	case StepTitle:
		m.title = m.inputs[0].Value()
	case StepProject:
		m.project = m.inputs[1].Value()
	case StepTags:
		// Don't auto-update tags here, we handle them manually in handleEnter
		// This prevents interference with our multi-tag input logic
	case StepPriority:
		// Don't auto-update priority to avoid validation issues
		// We handle this in handleEnter with proper validation
	case StepJira:
		m.jiraID = m.inputs[4].Value()
	case StepDueDate:
		m.dueDate = m.inputs[5].Value()
	case StepNotes:
		m.notes = m.inputs[6].Value()
	}
}

// createTask creates the task in the database
func (m AddTaskModel) createTask() (AddTaskModel, tea.Cmd) {
	// Parse due date if provided
	var dueDate *time.Time
	if m.dueDate != "" {
		parsedDate, err := parser.ParseDueDate(m.dueDate)
		if err != nil {
			m.err = fmt.Errorf("invalid due date: %w", err)
			return m, nil
		}
		dueDate = parsedDate
	}
	
	req := db.CreateTaskRequest{
		Title:    m.title,
		Project:  m.project,
		Tags:     m.tags,
		Priority: m.priority,
		JiraID:   m.jiraID,
		Note:     m.notes,
		DueDate:  dueDate,
	}
	
	task, err := db.CreateTask(req)
	if err != nil {
		m.err = err
		return m, nil
	}
	
	m.completed = true
	m.createdTaskID = task.ID
	m.createdTaskTitle = task.Title
	
	return m, tea.Quit
}