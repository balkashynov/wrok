package tui

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/parser"
)



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
	StepSave
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
	
	// Edit mode
	isEditMode    bool
	editTaskID    uint
	
	// State
	err           error
	completed     bool
	cancelled     bool
	validationErr string
	createdTaskID uint
	createdTaskTitle string
	
	// Tag input state
	isAddingTags bool
	
	// Shimmer effect for field labels
	shimmer *ShimmerState
	
	// Save confirmation modal
	showSaveModal bool
	saveModalChoice bool // true for Yes, false for No
}

// NewAddTaskModel creates a new add task TUI model
func NewAddTaskModel(prefilled map[string]string) AddTaskModel {
	inputs := make([]textinput.Model, 7)
	
	// Apply color theme to all inputs
	for i := 0; i < 7; i++ {
		inputs[i] = textinput.New()
		inputs[i].Width = 60
		
		// Apply color scheme
		inputs[i].TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPrimaryText))
		inputs[i].PlaceholderStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorPlaceholder))
		inputs[i].Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBright))
	}
	
	// Title input
	inputs[0].Placeholder = "Enter task title... (required)"
	inputs[0].Focus()
	inputs[0].CharLimit = 200
	
	// Project input
	inputs[1].Placeholder = "Project name (Enter to skip)"
	inputs[1].CharLimit = 50
	
	// Tags input
	inputs[2].Placeholder = "Add tag (Enter to skip, 'q' when done adding tags)"
	inputs[2].CharLimit = 50
	
	// Priority input
	inputs[3].Placeholder = "low/medium/high or 1/2/3 (Enter to skip - no priority)"
	inputs[3].CharLimit = 10
	
	// JIRA input
	inputs[4].Placeholder = "JIRA ticket like APP-42 (Enter to skip)"
	inputs[4].CharLimit = 20
	
	// Due date input
	inputs[5].Placeholder = "Due: dd/mm/yyyy, 3 days, 24 hours, 2 weeks (Enter to skip)"
	inputs[5].CharLimit = 50
	
	// Notes input
	inputs[6].Placeholder = "Additional notes (Enter to skip)"
	inputs[6].CharLimit = 500

	// Initialize shimmer effect
	shimmerConfig := DefaultShimmerConfig()
	shimmer := NewShimmerState(shimmerConfig)

	m := AddTaskModel{
		currentStep: StepTitle,
		inputs:      inputs,
		prefilled:   prefilled,
		tags:        []string{},
		shimmer:     shimmer,
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

// NewEditTaskModel creates a new edit task TUI model with existing task data
func NewEditTaskModel(taskID uint, prefilled map[string]string) AddTaskModel {
	// Create model using the same logic as NewAddTaskModel
	m := NewAddTaskModel(prefilled)
	
	// Set edit mode
	m.isEditMode = true
	m.editTaskID = taskID
	
	return m
}

// Init initializes the model
func (m AddTaskModel) Init() tea.Cmd {
	cmds := []tea.Cmd{textinput.Blink}
	
	// Start shimmer ticking if enabled
	if m.shimmer.ShouldTick() {
		cmds = append(cmds, tea.Tick(m.shimmer.GetTickInterval(), func(time.Time) tea.Msg {
			return shimmerTickMsg{}
		}))
	}
	
	return tea.Batch(cmds...)
}

// shimmerTickMsg is sent when shimmer should update
type shimmerTickMsg struct{}

// Update handles messages
func (m AddTaskModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case shimmerTickMsg:
		// Continue shimmer animation
		if m.shimmer.ShouldTick() {
			return m, tea.Tick(m.shimmer.GetTickInterval(), func(time.Time) tea.Msg {
				return shimmerTickMsg{}
			})
		}
		return m, nil
		
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
		// Handle save modal keys if modal is shown
		if m.showSaveModal {
			switch msg.String() {
			case "left", "right":
				m.saveModalChoice = !m.saveModalChoice
				return m, nil
			case "y", "Y":
				m.saveModalChoice = true
				return m.handleSaveChoice()
			case "n", "N":
				m.saveModalChoice = false
				return m.handleSaveChoice()
			case "enter":
				return m.handleSaveChoice()
			case "esc":
				// Close modal and go back to editing
				m.showSaveModal = false
				return m, nil
			case "ctrl+c":
				m.cancelled = true
				return m, tea.Quit
			}
			return m, nil
		}
		
		switch msg.String() {
		case "ctrl+c":
			m.cancelled = true
			return m, tea.Quit
		
		case "esc":
			// If on Save step, go back to previous step instead of showing modal
			if m.currentStep == StepSave {
				return m.prevStep()
			}
			
			// Check if there are any changes before showing save modal
			if !m.hasChanges() {
				// No changes, exit immediately
				m.cancelled = true
				return m, tea.Quit
			}
			
			// Show save confirmation modal for unsaved changes
			m.showSaveModal = true
			m.saveModalChoice = true // Default to "Yes"
			return m, nil
			
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
	
	// Update the current input (only for input steps, not Save step)
	var cmd tea.Cmd
	if m.currentStep < StepSave {
		m.inputs[m.currentStep], cmd = m.inputs[m.currentStep].Update(msg)
		
		// Update the corresponding field
		m.updateCurrentField()
	}
	
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
	
	// Handle very small terminals
	if m.width < 85 {
		return m.renderSmallLayout()
	}
	
	// Calculate adaptive column widths
	rightWidth := (m.width * 30) / 100 // Start with 30%
	if rightWidth < 50 {
		// Need more space, take up to 70%
		maxRightWidth := (m.width * 70) / 100
		if maxRightWidth >= 50 {
			rightWidth = 50
		} else {
			// Fallback to small layout
			return m.renderSmallLayout()
		}
	}
	
	leftWidth := m.width - rightWidth - 4 // Account for margins
	
	// Ensure minimum left width
	if leftWidth < 30 {
		leftWidth = 30
		rightWidth = m.width - leftWidth - 4
	}
	
	leftStyle := lipgloss.NewStyle().
		Width(leftWidth).
		Height(m.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
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
	mainView := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		" ", // Add explicit separator
		rightPanel,
	)
	
	// Add save modal overlay if shown
	if m.showSaveModal {
		return m.renderSaveModal(mainView)
	}
	
	return mainView
}

// renderWizard renders the step-by-step wizard
func (m AddTaskModel) renderWizard() string {
	var b strings.Builder
	
	
	// Title
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccentBright)).
		MarginBottom(1)
	
	titleText := "📝 Create New Task"
	if m.isEditMode {
		titleText = fmt.Sprintf("📝 Edit Task #%d", m.editTaskID)
	}
	
	b.WriteString(titleStyle.Render(titleText))
	b.WriteString("\n\n")
	
	// Current step indicator with dynamic coloring (with fallback)
	stepLabels := []string{"Title", "Project", "Tags", "Priority", "JIRA", "Due Date", "Notes", "Save"}
	
	// Check if terminal supports colors
	supportsColor := m.terminalSupportsColor()
	
	if supportsColor {
		// Define color codes for direct terminal output
		purpleColor := "\033[38;2;167;139;250m"   // ColorAccentBright - current step (purple)
		greenColor := "\033[38;2;34;197;94m"      // ColorSuccess - completed steps (green)
		darkGreyPurpleColor := "\033[38;2;109;115;131m" // ColorDisabledText - skipped steps (darker grey-purple)
		lightGreyColor := "\033[38;2;177;184;199m"      // ColorSecondaryText - future steps (lighter default grey)
		resetColor := "\033[0m"
		
		for i, label := range stepLabels {
			hasValue := m.stepHasValue(Step(i))
			
			// Add extra spacing before Save step to distinguish it
			if Step(i) == StepSave {
				b.WriteString("\n") // Extra line before Save step
			}
			
			if Step(i) == m.currentStep {
				// Current step - purple arrow
				if Step(i) == StepSave {
					// Special styling for Save step when current
					b.WriteString(fmt.Sprintf("%s▶ 💾 %s%s\n", purpleColor, label, resetColor))
				} else {
					b.WriteString(fmt.Sprintf("%s▶ %s%s\n", purpleColor, label, resetColor))
				}
			} else if m.isEditMode && hasValue {
				// Edit mode: all populated steps show as completed (green)
				b.WriteString(fmt.Sprintf("%s✓ %s%s\n", greenColor, label, resetColor))
			} else if Step(i) < m.currentStep {
				// Check if step was actually completed or skipped
				if hasValue {
					// Completed with value - green checkmark and text
					b.WriteString(fmt.Sprintf("%s✓ %s%s\n", greenColor, label, resetColor))
				} else {
					// Skipped - no icon, darker grey-purple text
					b.WriteString(fmt.Sprintf("%s  %s%s\n", darkGreyPurpleColor, label, resetColor))
				}
			} else {
				// Future step - lighter default grey (or grey in edit mode if no value)
				color := lightGreyColor
				if m.isEditMode && !hasValue {
					color = darkGreyPurpleColor // Darker grey for empty fields in edit mode
				}
				
				if Step(i) == StepSave {
					// Special styling for Save step when not current
					b.WriteString(fmt.Sprintf("%s  💾 %s%s\n", color, label, resetColor))
				} else {
					b.WriteString(fmt.Sprintf("%s  %s%s\n", color, label, resetColor))
				}
			}
		}
	} else {
		// Fallback for terminals that don't support colors - plain text
		for i, label := range stepLabels {
			hasValue := m.stepHasValue(Step(i))
			
			// Add extra spacing before Save step to distinguish it
			if Step(i) == StepSave {
				b.WriteString("\n") // Extra line before Save step
			}
			
			if Step(i) == m.currentStep {
				// Current step - arrow
				if Step(i) == StepSave {
					b.WriteString(fmt.Sprintf("▶ 💾 %s\n", label))
				} else {
					b.WriteString(fmt.Sprintf("▶ %s\n", label))
				}
			} else if m.isEditMode && hasValue {
				// Edit mode: all populated steps show as completed (checkmark)
				b.WriteString(fmt.Sprintf("✓ %s\n", label))
			} else if Step(i) < m.currentStep {
				// Check if step was actually completed or skipped
				if hasValue {
					// Completed with value - checkmark
					b.WriteString(fmt.Sprintf("✓ %s\n", label))
				} else {
					// Skipped - no icon
					b.WriteString(fmt.Sprintf("  %s\n", label))
				}
			} else {
				// Future step
				if Step(i) == StepSave {
					b.WriteString(fmt.Sprintf("  💾 %s\n", label))
				} else {
					b.WriteString(fmt.Sprintf("  %s\n", label))
				}
			}
		}
	}
	b.WriteString("\n")
	
	// Current input field - simple text without styling boxes
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
		
	case StepSave:
		b.WriteString("💾 Save Task\n")
		b.WriteString("Press Enter to save task")
	}
	
	// Show validation error if any
	if m.validationErr != "" {
		errorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorError)).
			Bold(true).
			MarginTop(1)
		b.WriteString(errorStyle.Render("❌ " + m.validationErr))
	}
	
	b.WriteString("\n\n")
	
	// Help text
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHelpText)).
		Italic(true)
	b.WriteString(helpStyle.Render("Enter: Next | Tab/↓: Next | Shift+Tab/↑: Back | Esc: Cancel"))
	
	return b.String()
}

// terminalSupportsColor checks if the terminal supports ANSI truecolor
func (m AddTaskModel) terminalSupportsColor() bool {
	// Only support terminals with truecolor capability
	colorTerm := os.Getenv("COLORTERM")
	
	// Only enable colors for truecolor terminals
	return colorTerm == "truecolor"
}

// stepHasValue checks if a step has been filled with a value (not skipped)
func (m AddTaskModel) stepHasValue(step Step) bool {
	switch step {
	case StepTitle:
		return strings.TrimSpace(m.title) != ""
	case StepProject:
		return strings.TrimSpace(m.project) != ""
	case StepTags:
		return len(m.tags) > 0
	case StepPriority:
		return strings.TrimSpace(m.priority) != ""
	case StepJira:
		return strings.TrimSpace(m.jiraID) != ""
	case StepDueDate:
		return strings.TrimSpace(m.dueDate) != ""
	case StepNotes:
		return strings.TrimSpace(m.notes) != ""
	case StepSave:
		return false // Save step doesn't have a value, it's an action
	default:
		return false
	}
}

// hasChanges checks if there are any changes made to the task
func (m AddTaskModel) hasChanges() bool {
	if m.isEditMode {
		// In edit mode, check if any field was changed from original prefilled values
		if m.prefilled == nil {
			return true // If no prefilled data, assume changes
		}
		
		// Compare each field with prefilled values
		if strings.TrimSpace(m.title) != strings.TrimSpace(m.prefilled["title"]) {
			return true
		}
		if strings.TrimSpace(m.project) != strings.TrimSpace(m.prefilled["project"]) {
			return true
		}
		if strings.TrimSpace(m.priority) != strings.TrimSpace(m.prefilled["priority"]) {
			return true
		}
		if strings.TrimSpace(m.jiraID) != strings.TrimSpace(m.prefilled["jira"]) {
			return true
		}
		if strings.TrimSpace(m.dueDate) != strings.TrimSpace(m.prefilled["due_date"]) {
			return true
		}
		if strings.TrimSpace(m.notes) != strings.TrimSpace(m.prefilled["notes"]) {
			return true
		}
		
		// Compare tags (more complex comparison)
		prefilledTags := strings.TrimSpace(m.prefilled["tags"])
		currentTagsStr := ""
		if len(m.tags) > 0 {
			var tagNames []string
			for _, tag := range m.tags {
				tagNames = append(tagNames, "#"+tag)
			}
			currentTagsStr = strings.Join(tagNames, ", ")
		}
		if currentTagsStr != prefilledTags {
			return true
		}
		
		return false // No changes detected
	} else {
		// In add mode, check if any field has content
		return strings.TrimSpace(m.title) != "" || 
		       strings.TrimSpace(m.project) != "" ||
		       len(m.tags) > 0 ||
		       strings.TrimSpace(m.priority) != "" ||
		       strings.TrimSpace(m.jiraID) != "" ||
		       strings.TrimSpace(m.dueDate) != "" ||
		       strings.TrimSpace(m.notes) != ""
	}
}

// renderPreview renders the live task preview
func (m AddTaskModel) renderPreview() string {
	var b strings.Builder
	
	// Handle very small terminals with fallback
	if m.width < 85 {
		return m.renderSmallPreview()
	}
	
	// Calculate adaptive width for right panel
	// Start with 30% but ensure minimum 50px and allow up to 70%
	rightPanelWidth := (m.width * 30) / 100 // Start with 30%
	if rightPanelWidth < 50 {
		// If too small, take more space up to 70%
		maxRightWidth := (m.width * 70) / 100
		if maxRightWidth >= 50 {
			rightPanelWidth = 50
		} else {
			// Terminal too small for proper layout
			return m.renderSmallPreview()
		}
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
	
	// WROK ASCII Logo
	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccentMain)).
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
		Foreground(lipgloss.Color(ColorAccentMain)).
		Align(lipgloss.Center)
	cardContent.WriteString(separatorStyle.Render("═══════════════════════════════════"))
	cardContent.WriteString("\n")
	
	// Title section with nice border box and shimmer effect
	var titleText string
	if m.title != "" {
		titleText = m.title
	} else {
		titleText = "Untitled Task"
	}
	
	// Apply shimmer to the title text
	shimmerTitle := m.shimmer.RenderShimmerText(titleText, cardWidth - 6) // Account for border and padding
	
	// Create a fancy title box with double border
	titleBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color(ColorAccentMain)).
		Bold(true).
		Padding(0, 1).
		Align(lipgloss.Center).
		Width(cardWidth - 4) // Fit within card
	
	// Add emoji and shimmered title together (reset ANSI codes so border displays properly)
	titleWithEmoji := fmt.Sprintf("🎯 %s\033[0m", shimmerTitle)
	cardContent.WriteString(titleBoxStyle.Render(titleWithEmoji))
	cardContent.WriteString("\n")
	
	// Status with nice styling - placeholder purple and lowercase
	statusStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPlaceholder)).
		Bold(true).
		Padding(0, 1).
		Align(lipgloss.Center)
	cardContent.WriteString(statusStyle.Render("● todo"))
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
			Foreground(lipgloss.Color(ColorAccentBright)).
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
				Foreground(lipgloss.Color(ColorError)).
				Bold(true).
				Render("🔥 HIGH")
		case "medium":
			priorityDisplay = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorWarning)).
				Bold(true).
				Render("⚡ MEDIUM")
		case "low":
			priorityDisplay = lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSuccess)).
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
			warningText = " (Expected XXX-123)"
		}
		
		jiraStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorPrimaryText)).
			Background(lipgloss.Color(ColorBorder)).
			Bold(true).
			Padding(0, 1)
		
		if warningText != "" {
			// Show JIRA with darker grey hint text
			warningStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorHelpText)).
				Italic(true)
			metadata.WriteString(fmt.Sprintf("🎫 JIRA: %s %s\n", jiraStyle.Render(displayJira), warningStyle.Render(warningText)))
		} else {
			metadata.WriteString(fmt.Sprintf("🎫 JIRA: %s\n", jiraStyle.Render(displayJira)))
		}
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
			Foreground(lipgloss.Color(ColorSecondaryText)).
			Italic(true)
		metadata.WriteString(fmt.Sprintf("📝 Notes: %s\n", noteStyle.Render(m.notes)))
	}
	
	// Add metadata to card
	cardContent.WriteString(metadataStyle.Render(metadata.String()))
	
	// Create the card with static purple border
	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccentMain)).
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

// renderSmallPreview renders a compact preview for small terminals
func (m AddTaskModel) renderSmallPreview() string {
	var b strings.Builder
	
	// Simple compact preview with terminal size hint
	b.WriteString("═══ PREVIEW ═══\n")
	b.WriteString("💡 Tip: Stretch terminal for better UI\n")
	
	if m.title != "" {
		b.WriteString(fmt.Sprintf("📋 %s\n", m.title))
	}
	
	if m.project != "" {
		b.WriteString(fmt.Sprintf("📁 %s\n", m.project))
	}
	
	if len(m.tags) > 0 {
		b.WriteString(fmt.Sprintf("🔖 %s\n", strings.Join(m.tags, ", ")))
	}
	
	if m.priority != "" {
		b.WriteString(fmt.Sprintf("⚡ %s\n", parser.NormalizePriority(m.priority)))
	}
	
	if m.jiraID != "" {
		b.WriteString(fmt.Sprintf("🎫 %s\n", m.jiraID))
	}
	
	if m.dueDate != "" {
		parsedDate, err := parser.ParseDueDate(m.dueDate)
		if err == nil && parsedDate != nil {
			b.WriteString(fmt.Sprintf("📅 %s\n", parser.FormatDueDate(parsedDate)))
		} else {
			b.WriteString(fmt.Sprintf("📅 %s\n", m.dueDate))
		}
	}
	
	if m.notes != "" {
		b.WriteString(fmt.Sprintf("📝 %s\n", m.notes))
	}
	
	b.WriteString("═══════════════\n")
	return b.String()
}

// renderSmallLayout renders entire TUI for very small terminals
func (m AddTaskModel) renderSmallLayout() string {
	// Single column layout for small terminals
	style := lipgloss.NewStyle().
		Width(m.width - 2).
		Height(m.height - 2).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Padding(1)
	
	// Wizard content (already has logo at top)
	wizard := m.renderWizard()
	
	// Small preview
	preview := m.renderSmallPreview()
	
	// Combine vertically
	content := wizard + "\n" + preview
	
	return style.Render(content)
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
		// Notes is optional, move to Save step
		return m.nextStep()
		
	case StepSave:
		// Save the task
		return m.createTask()
	}
	
	return m, nil
}

// nextStep moves to the next step
func (m AddTaskModel) nextStep() (AddTaskModel, tea.Cmd) {
	if m.currentStep < StepSave {
		m.inputs[m.currentStep].Blur()
		m.currentStep++
		if m.currentStep < StepSave {
			// Only focus input fields, not the Save step
			m.inputs[m.currentStep].Focus()
		}
		// Reset shimmer for new field
		m.shimmer.Reset()
	}
	return m, textinput.Blink
}

// prevStep moves to the previous step
func (m AddTaskModel) prevStep() (AddTaskModel, tea.Cmd) {
	if m.currentStep > StepTitle {
		if m.currentStep <= StepNotes {
			m.inputs[m.currentStep].Blur()
		}
		m.currentStep--
		if m.currentStep <= StepNotes {
			m.inputs[m.currentStep].Focus()
		}
		// Reset shimmer for new field
		m.shimmer.Reset()
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
	
	if m.isEditMode {
		// Update existing task
		updateReq := db.UpdateTaskRequest{
			ID:       m.editTaskID,
			Title:    m.title,
			Project:  m.project,
			Tags:     m.tags,
			Priority: m.priority,
			JiraID:   m.jiraID,
			Note:     m.notes,
			DueDate:  dueDate,
		}
		
		task, err := db.UpdateTask(updateReq)
		if err != nil {
			m.err = err
			return m, nil
		}
		
		m.completed = true
		m.createdTaskID = task.ID
		m.createdTaskTitle = task.Title
	} else {
		// Create new task
		createReq := db.CreateTaskRequest{
			Title:    m.title,
			Project:  m.project,
			Tags:     m.tags,
			Priority: m.priority,
			JiraID:   m.jiraID,
			Note:     m.notes,
			DueDate:  dueDate,
		}
		
		task, err := db.CreateTask(createReq)
		if err != nil {
			m.err = err
			return m, nil
		}
		
		m.completed = true
		m.createdTaskID = task.ID
		m.createdTaskTitle = task.Title
	}
	
	return m, tea.Quit
}

// handleSaveChoice handles the save confirmation modal response
func (m AddTaskModel) handleSaveChoice() (AddTaskModel, tea.Cmd) {
	m.showSaveModal = false
	
	if m.saveModalChoice {
		// User chose "Yes", save the task
		return m.createTask()
	} else {
		// User chose "No", cancel without saving
		m.cancelled = true
		return m, tea.Quit
	}
}

// renderSaveModal renders the save confirmation modal overlay
func (m AddTaskModel) renderSaveModal(background string) string {
	// Modal dimensions
	modalWidth := 50
	modalHeight := 7
	
	// Modal content
	var modalContent strings.Builder
	modalContent.WriteString("Save changes?\n\n")
	
	// Yes/No options with highlighting
	yesStyle := lipgloss.NewStyle().Padding(0, 2)
	noStyle := lipgloss.NewStyle().Padding(0, 2)
	
	if m.saveModalChoice {
		// "Yes" is selected
		yesStyle = yesStyle.
			Background(lipgloss.Color(ColorAccentBright)).
			Foreground(lipgloss.Color("#000000")).
			Bold(true)
	} else {
		// "No" is selected  
		noStyle = noStyle.
			Background(lipgloss.Color(ColorError)).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)
	}
	
	yesButton := yesStyle.Render("Yes")
	noButton := noStyle.Render("No")
	
	modalContent.WriteString(lipgloss.JoinHorizontal(
		lipgloss.Center,
		yesButton,
		"   ",
		noButton,
	))
	modalContent.WriteString("\n\n")
	modalContent.WriteString("← → or Y/N to choose, Enter to confirm\nEsc to cancel")
	
	// Create modal box
	modalStyle := lipgloss.NewStyle().
		Width(modalWidth).
		Height(modalHeight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccentBright)).
		Background(lipgloss.Color(ColorCardBackground)).
		Padding(1).
		Align(lipgloss.Center)
	
	modal := modalStyle.Render(modalContent.String())
	
	// Position the modal with same background
	return lipgloss.Place(
		m.width, m.height,
		lipgloss.Center, lipgloss.Center,
		modal,
	)
}