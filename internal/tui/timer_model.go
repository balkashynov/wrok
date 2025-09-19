package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/balkashynov/wrok/internal/db"
	"github.com/balkashynov/wrok/internal/models"
)

// TimerModel represents the TUI model for time tracking
type TimerModel struct {
	width   int
	height  int
	session *models.Session
	task    *models.Task

	// Timer state
	elapsedTime time.Duration
	lastUpdate  time.Time

	// Animation state
	timerAnimation int // For animated timer display

	// UI state
	stopping bool // True when user pressed S and we're stopping
	exiting  bool // True when user pressed ESC/Q and we're exiting without stopping
}

// timerTickMsg is sent every second to update the timer
type timerTickMsg struct{}

// animationTickMsg is sent for faster animations
type animationTickMsg struct{}

// NewTimerModel creates a new timer TUI model
func NewTimerModel(session *models.Session) TimerModel {
	return TimerModel{
		session:        session,
		task:           &session.Task,
		elapsedTime:    time.Since(session.StartedAt),
		lastUpdate:     time.Now(),
		timerAnimation: 0,
		stopping:       false,
		exiting:        false,
	}
}

// Init initializes the timer model
func (m TimerModel) Init() tea.Cmd {
	// Start both timer and animation tickers
	return tea.Batch(
		tea.Tick(time.Second, func(t time.Time) tea.Msg {
			return timerTickMsg{}
		}),
		tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
			return animationTickMsg{}
		}),
	)
}

// Update handles messages
func (m TimerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case timerTickMsg:
		// Update elapsed time
		now := time.Now()
		m.elapsedTime = now.Sub(m.session.StartedAt)
		m.lastUpdate = now

		// Continue ticking if not stopping or exiting
		if !m.stopping && !m.exiting {
			return m, tea.Tick(time.Second, func(t time.Time) tea.Msg {
				return timerTickMsg{}
			})
		}
		return m, nil

	case animationTickMsg:
		// Update animation states
		m.timerAnimation = (m.timerAnimation + 1) % 4

		// Continue animating if not stopping or exiting
		if !m.stopping && !m.exiting {
			return m, tea.Tick(250*time.Millisecond, func(t time.Time) tea.Msg {
				return animationTickMsg{}
			})
		}
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "s", "S":
			// Stop the timer and save
			m.stopping = true
			return m, tea.Quit
		case "ctrl+c", "esc", "q":
			// Exit without stopping
			m.exiting = true
			return m, tea.Quit
		}
	}

	return m, nil
}

// View renders the timer TUI
func (m TimerModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Help bar at bottom
	helpBar := m.renderHelpBar()
	helpBarHeight := 1

	// Available height for content (total minus help bar and gap)
	contentHeight := m.height - helpBarHeight - 1

	// Check if screen is too narrow for split view
	if m.width < 90 {
		// Narrow view: just timer panel, full width
		timerPanel := m.renderTimerPanel(m.width, contentHeight)

		return lipgloss.JoinVertical(
			lipgloss.Left,
			timerPanel,
			helpBar,
		)
	}

	// Wide view: split screen
	leftWidth := m.width / 2
	rightWidth := m.width - leftWidth - 2 // -2 for gap

	// Left side: timer (full height)
	leftPanel := m.renderTimerPanel(leftWidth, contentHeight)

	// Right side: task details (full height)
	rightPanel := m.renderTaskDetailsPanel(rightWidth, contentHeight)

	// Main content
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		"  ", // Gap
		rightPanel,
	)

	// Final layout
	return lipgloss.JoinVertical(
		lipgloss.Left,
		content,
		helpBar,
	)
}

// renderTimerPanel renders the left timer panel
func (m TimerModel) renderTimerPanel(width, height int) string {
	// Build all content components first
	var components []string

	// Animated header
	animChars := []string{"⏱", "⏲", "⏱", "⏲"}
	animChar := animChars[m.timerAnimation]
	headerText := fmt.Sprintf("%s  TRACKING TIME  %s", animChar, animChar)

	headerStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccentBright)).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)

	components = append(components, headerStyle.Render(headerText))

	// Task ID and title
	idStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccentMain)).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)

	taskIdText := fmt.Sprintf("#%d", m.task.ID)
	components = append(components, idStyle.Render(taskIdText))

	// Task title
	titleStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimaryText)).
		Bold(true).
		Align(lipgloss.Center).
		Width(width)

	titleText := m.task.Title
	if len(titleText) > width-4 {
		titleText = titleText[:width-7] + "..."
	}
	components = append(components, titleStyle.Render(titleText))

	// Big clock display
	clockDisplay := m.renderBigClock()
	clockLines := strings.Split(clockDisplay, "\n")
	clockContent := ""
	for _, line := range clockLines {
		centeredLine := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(width).
			Render(line)
		clockContent += centeredLine + "\n"
	}
	components = append(components, strings.TrimRight(clockContent, "\n"))

	// Session start time
	sessionInfo := fmt.Sprintf("Started at %s", m.session.StartedAt.Format("15:04:05"))
	sessionStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorSecondaryText)).
		Italic(true).
		Align(lipgloss.Center).
		Width(width)
	components = append(components, sessionStyle.Render(sessionInfo))

	// Join all components with spacing and center vertically
	content := strings.Join(components, "\n\n")

	// Use lipgloss to center content vertically and fill the full height
	panelStyle := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Align(lipgloss.Center, lipgloss.Center)

	return panelStyle.Render(content)
}

// renderBigClock renders ASCII art clock
func (m TimerModel) renderBigClock() string {
	duration := m.elapsedTime
	hours := int(duration.Hours())
	minutes := int(duration.Minutes()) % 60
	seconds := int(duration.Seconds()) % 60

	// ASCII art for digits (3x5 characters each)
	digits := map[rune][][]string{
		'0': {
			{" ███ "},
			{"█   █"},
			{"█   █"},
			{"█   █"},
			{" ███ "},
		},
		'1': {
			{"  █  "},
			{" ██  "},
			{"  █  "},
			{"  █  "},
			{"█████"},
		},
		'2': {
			{" ███ "},
			{"█   █"},
			{"   █ "},
			{"  █  "},
			{"█████"},
		},
		'3': {
			{" ███ "},
			{"█   █"},
			{"  ██ "},
			{"█   █"},
			{" ███ "},
		},
		'4': {
			{"█   █"},
			{"█   █"},
			{"█████"},
			{"    █"},
			{"    █"},
		},
		'5': {
			{"█████"},
			{"█    "},
			{"████ "},
			{"    █"},
			{"████ "},
		},
		'6': {
			{" ███ "},
			{"█    "},
			{"████ "},
			{"█   █"},
			{" ███ "},
		},
		'7': {
			{"█████"},
			{"    █"},
			{"   █ "},
			{"  █  "},
			{" █   "},
		},
		'8': {
			{" ███ "},
			{"█   █"},
			{" ███ "},
			{"█   █"},
			{" ███ "},
		},
		'9': {
			{" ███ "},
			{"█   █"},
			{" ████"},
			{"    █"},
			{" ███ "},
		},
		':': {
			{"     "},
			{"  █  "},
			{"     "},
			{"  █  "},
			{"     "},
		},
	}

	// Format time string
	timeStr := ""
	if hours > 0 {
		timeStr = fmt.Sprintf("%02d:%02d:%02d", hours, minutes, seconds)
	} else {
		timeStr = fmt.Sprintf("%02d:%02d", minutes, seconds)
	}

	// Build the big clock display
	var lines [5]strings.Builder

	for _, char := range timeStr {
		if digitArt, ok := digits[char]; ok {
			for i := 0; i < 5; i++ {
				lines[i].WriteString(digitArt[i][0])
				lines[i].WriteString(" ") // Space between digits
			}
		}
	}

	// Apply consistent color (no blinking)
	clockStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccentBright)).
		Bold(true)

	var result strings.Builder
	for i := 0; i < 5; i++ {
		result.WriteString(clockStyle.Render(lines[i].String()))
		if i < 4 {
			result.WriteString("\n")
		}
	}

	return result.String()
}

// renderTaskDetailsPanel renders the right panel (stolen from ls)
func (m TimerModel) renderTaskDetailsPanel(width, height int) string {
	task := m.task
	var b strings.Builder

	b.WriteString("\n")

	// ASCII logo at top (from ls)
	logoLines := []string{
		"██╗    ██╗██████╗  ██████╗ ██╗  ██╗",
		"██║    ██║██╔══██╗██╔═══██╗██║ ██╔╝",
		"██║ █╗ ██║██████╔╝██║   ██║█████╔╝ ",
		"██║███╗██║██╔══██╗██║   ██║██╔═██╗ ",
		"╚███╔███╔╝██║  ██║╚██████╔╝██║  ██╗",
		" ╚══╝╚══╝ ╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═╝",
	}

	logoStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorAccentMain)).
		Bold(true).
		Align(lipgloss.Center).
		Width(width-8)

	b.WriteString(logoStyle.Render(strings.Join(logoLines, "\n")))
	b.WriteString("\n\n")

	// Separator line (from ls)
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorBorder)).
		Align(lipgloss.Center).
		Width(width-8)
	separatorLine := strings.Repeat("─", min(width-12, 40))
	b.WriteString(separatorStyle.Render(separatorLine))
	b.WriteString("\n\n")

	// Title in bordered box (from ls)
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorPrimaryText)).
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccentMain)).
		Width(width-12).
		Padding(0, 1)
	b.WriteString(titleStyle.Render(task.Title))
	b.WriteString("\n\n")

	// Task details in structured format (from ls)
	// Status with emoji
	statusIcon := "○"
	statusColor := ColorSecondaryText
	statusText := "todo"
	if task.Status == "done" {
		statusIcon = "✅"
		statusColor = ColorSuccess
		statusText = "done"
	} else if task.Status == "archived" {
		statusIcon = "▪"
		statusColor = ColorDisabledText
		statusText = "archived"
	}

	statusStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width-8)
	statusLine := fmt.Sprintf("%s Status: %s", statusIcon,
		lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(statusText))
	b.WriteString(statusStyle.Render(statusLine))
	b.WriteString("\n")

	// Project with emoji (from ls)
	projectStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width-8)
	projectValue := "none"
	projectColor := ColorDisabledText
	if task.Project != "" {
		projectValue = task.Project
		projectColor = ColorAccentBright
	}
	projectLine := fmt.Sprintf("📁 Project: %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color(projectColor)).Render(projectValue))
	b.WriteString(projectStyle.Render(projectLine))
	b.WriteString("\n")

	// Priority with emoji (from ls)
	priorityStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width-8)
	priorityIcon := "⚪"
	priorityValue := "none"
	priorityColor := ColorDisabledText
	if task.Priority > 0 && task.Priority <= 3 {
		priorities := []string{"", "low", "medium", "high"}
		priorityValue = priorities[task.Priority]
		switch task.Priority {
		case 3:
			priorityIcon = "🔴"
			priorityColor = ColorError
		case 2:
			priorityIcon = "🟡"
			priorityColor = ColorWarning
		case 1:
			priorityIcon = "🟢"
			priorityColor = ColorSecondaryText
		}
	}
	priorityLine := fmt.Sprintf("%s Priority: %s", priorityIcon,
		lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor)).Render(priorityValue))
	b.WriteString(priorityStyle.Render(priorityLine))
	b.WriteString("\n")

	// Tags (from ls style)
	tagsStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width-8)
	tagsValue := "none"
	tagsColor := ColorDisabledText
	if len(task.Tags) > 0 {
		var tagNames []string
		for _, tag := range task.Tags {
			tagNames = append(tagNames, "#"+tag.Name)
		}
		tagsValue = strings.Join(tagNames, " ")
		tagsColor = ColorAccentBright
	}
	tagsLine := fmt.Sprintf("🏷️  Tags: %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color(tagsColor)).Render(tagsValue))
	b.WriteString(tagsStyle.Render(tagsLine))
	b.WriteString("\n")

	// JIRA (from ls style)
	jiraStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width-8)
	jiraValue := "none"
	jiraColor := ColorDisabledText
	if task.JiraID != "" {
		jiraValue = task.JiraID
		jiraColor = ColorAccentMain
	}
	jiraLine := fmt.Sprintf("🎯 JIRA: %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color(jiraColor)).Render(jiraValue))
	b.WriteString(jiraStyle.Render(jiraLine))
	b.WriteString("\n")

	// Due date (from ls style)
	dueStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width-8)
	dueValue := "none"
	dueColor := ColorDisabledText
	if task.Due != nil && !task.Due.IsZero() {
		dueValue = task.Due.Format("Jan 02, 2006")
		dueColor = ColorWarning
	}
	dueLine := fmt.Sprintf("📅 Due: %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color(dueColor)).Render(dueValue))
	b.WriteString(dueStyle.Render(dueLine))
	b.WriteString("\n")

	// Created date (from ls style)
	createdStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width-8)
	createdValue := task.CreatedAt.Format("Jan 02, 2006")
	createdLine := fmt.Sprintf("📝 Created: %s",
		lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondaryText)).Render(createdValue))
	b.WriteString(createdStyle.Render(createdLine))

	return b.String()
}

// renderHelpBar renders the help bar at the bottom
func (m TimerModel) renderHelpBar() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHelpText)).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.width)

	helpText := "s stop & save · esc/q exit (keep running) · ctrl+c force quit"

	return helpStyle.Render(helpText)
}

// RunTimerTUI runs the timer TUI
func RunTimerTUI(session *models.Session) error {
	model := NewTimerModel(session)

	p := tea.NewProgram(model, tea.WithAltScreen())

	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if we need to stop the session
	timerModel := finalModel.(TimerModel)
	if timerModel.stopping {
		// Stop the active session
		stoppedSession, err := db.StopActiveSession()
		if err != nil {
			return fmt.Errorf("failed to stop session: %w", err)
		}

		// Show completion message
		duration := time.Duration(stoppedSession.DurationSeconds) * time.Second
		fmt.Printf("⏹️  Stopped tracking time for task #%d: %s\n", stoppedSession.TaskID, stoppedSession.Task.Title)
		fmt.Printf("📊 Session duration: %s\n", formatDuration(duration))
	} else if timerModel.exiting {
		// Just exiting without stopping
		fmt.Printf("\n💡 Timer is still running in the background for task #%d: %s\n", session.TaskID, session.Task.Title)
		fmt.Printf("   Use 'wrok status' to check current timer or 'wrok stop' to stop it.\n")
	}

	return nil
}

// formatDuration formats a duration in a human-readable way (copied from time.go)
func formatDuration(d time.Duration) string {
	if d.Hours() >= 1 {
		return fmt.Sprintf("%.1fh", d.Hours())
	} else if d.Minutes() >= 1 {
		return fmt.Sprintf("%.0fm", d.Minutes())
	} else {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
}