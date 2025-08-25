package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/balkashynov/wrok/internal/models"
)

// ListModel represents the TUI model for listing tasks
type ListModel struct {
	width  int
	height int
	
	// Task data
	tasks        []models.Task
	selectedTask int // index in tasks slice
	
	// UI state
	focus        Focus
	searchActive bool
	searchQuery  string
	
	// Shimmer effect for selected task title
	shimmer *ShimmerState
	
	// Pagination
	currentPage int
	tasksPerPage int
}

// Focus represents what UI element has focus
type Focus int

const (
	FocusTable Focus = iota
	FocusSearch
	FocusModal
)

// NewListModel creates a new list TUI model
func NewListModel(tasks []models.Task) ListModel {
	// Initialize shimmer effect
	shimmerConfig := DefaultShimmerConfig()
	shimmer := NewShimmerState(shimmerConfig)

	model := ListModel{
		tasks:        tasks,
		selectedTask: 0,
		focus:        FocusTable,
		shimmer:      shimmer,
		currentPage:  0,
	}
	
	// Pre-select first task if available
	if len(tasks) > 0 {
		model.selectedTask = 0
	}
	
	return model
}

// Init initializes the model
func (m ListModel) Init() tea.Cmd {
	cmds := []tea.Cmd{}
	
	// Start shimmer ticking if enabled
	if m.shimmer.ShouldTick() {
		cmds = append(cmds, tea.Tick(m.shimmer.GetTickInterval(), func(time.Time) tea.Msg {
			return shimmerTickMsg{}
		}))
	}
	
	return tea.Batch(cmds...)
}

// Update handles messages
func (m ListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case shimmerTickMsg:
		// Continue shimmer animation if focused on table
		if m.focus == FocusTable && m.shimmer.ShouldTick() {
			return m, tea.Tick(m.shimmer.GetTickInterval(), func(time.Time) tea.Msg {
				return shimmerTickMsg{}
			})
		}
		return m, nil
		
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		
		// Calculate tasks per page based on available height
		// Height - header(2) - pagination(1) - help(1) - borders(4) - top/bottom margins(4) = content height  
		availableHeight := m.height - 12
		if availableHeight < 3 {
			availableHeight = 3
		}
		m.tasksPerPage = availableHeight
		
		return m, nil
		
	case tea.KeyMsg:
		if m.focus == FocusSearch {
			return m.handleSearchKeys(msg)
		}
		
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			// Handle escape key - exit search mode first if active, otherwise quit
			if msg.String() == "esc" && m.searchActive {
				m.focus = FocusTable
				m.searchActive = false
				m.searchQuery = ""
				m.shimmer.SetActive(true) // Resume shimmer
				return m, nil
			}
			return m, tea.Quit
			
		case "up", "k":
			return m.moveSelectionUp(), nil
			
		case "down", "j":
			return m.moveSelectionDown(), nil
			
		case "left", "h":
			return m.prevPage(), nil
			
		case "right", "l":
			return m.nextPage(), nil
			
		case "/":
			// Enter search mode
			m.focus = FocusSearch
			m.searchActive = true
			m.shimmer.SetActive(false) // Stop shimmer when not focused on table
			return m, nil
			
			
		// TODO: Add other hotkeys (e, d, a, s, F)
		}
	}
	
	return m, nil
}

// handleSearchKeys handles key input when in search mode
func (m ListModel) handleSearchKeys(msg tea.KeyMsg) (ListModel, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Exit search
		m.focus = FocusTable
		m.searchActive = false
		m.searchQuery = ""
		m.shimmer.SetActive(true)
		return m, nil
		
	case "enter":
		// Apply search and return to table
		m.focus = FocusTable
		m.searchActive = false
		m.shimmer.SetActive(true)
		// TODO: Apply search filter
		return m, nil
		
	case "backspace":
		if len(m.searchQuery) > 0 {
			m.searchQuery = m.searchQuery[:len(m.searchQuery)-1]
		}
		return m, nil
		
	default:
		// Add character to search query
		m.searchQuery += msg.String()
		return m, nil
	}
}

// moveSelectionUp moves the selection up
func (m ListModel) moveSelectionUp() ListModel {
	if m.selectedTask > 0 {
		m.selectedTask--
		m.shimmer.Reset() // Reset shimmer for new selection
		
		// Auto-pagination: if we scrolled above current page, go to previous page
		currentPageStart := m.currentPage * m.tasksPerPage
		if m.selectedTask < currentPageStart && m.currentPage > 0 {
			m.currentPage--
		}
	}
	return m
}

// moveSelectionDown moves the selection down
func (m ListModel) moveSelectionDown() ListModel {
	if m.selectedTask < len(m.tasks)-1 {
		m.selectedTask++
		m.shimmer.Reset() // Reset shimmer for new selection
		
		// Auto-pagination: if we scrolled below current page, go to next page
		currentPageEnd := min((m.currentPage+1)*m.tasksPerPage-1, len(m.tasks)-1)
		maxPages := (len(m.tasks) + m.tasksPerPage - 1) / m.tasksPerPage
		if m.selectedTask > currentPageEnd && m.currentPage < maxPages-1 {
			m.currentPage++
		}
	}
	return m
}

// prevPage goes to previous page
func (m ListModel) prevPage() ListModel {
	if m.currentPage > 0 {
		m.currentPage--
		// Adjust selection to be within the new page
		maxIndex := min((m.currentPage+1)*m.tasksPerPage-1, len(m.tasks)-1)
		if m.selectedTask > maxIndex {
			m.selectedTask = maxIndex
		}
		minIndex := m.currentPage * m.tasksPerPage
		if m.selectedTask < minIndex {
			m.selectedTask = minIndex
		}
		m.shimmer.Reset()
	}
	return m
}

// nextPage goes to next page
func (m ListModel) nextPage() ListModel {
	maxPages := (len(m.tasks) + m.tasksPerPage - 1) / m.tasksPerPage
	if m.currentPage < maxPages-1 {
		m.currentPage++
		// Adjust selection to be within the new page
		minIndex := m.currentPage * m.tasksPerPage
		if m.selectedTask < minIndex {
			m.selectedTask = minIndex
		}
		maxIndex := min((m.currentPage+1)*m.tasksPerPage-1, len(m.tasks)-1)
		if m.selectedTask > maxIndex {
			m.selectedTask = maxIndex
		}
		m.shimmer.Reset()
	}
	return m
}

// View renders the TUI
func (m ListModel) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}
	
	// Calculate layout
	leftWidth := m.width * 60 / 100  // 60% for table
	rightWidth := m.width - leftWidth - 1 // Rest for details
	
	// Left panel: Task table
	leftPanel := m.renderTaskTable(leftWidth)
	
	// Right panel: Task details
	rightPanel := m.renderTaskDetails(rightWidth)
	
	// Main content
	content := lipgloss.JoinHorizontal(
		lipgloss.Top,
		leftPanel,
		" ",
		rightPanel,
	)
	
	// Search bar (if active)
	var searchBar string
	if m.searchActive {
		searchBar = m.renderSearchBar()
	} else {
		searchBar = m.renderHelpBar()
	}
	
	// Add small margin at top and bottom
	return lipgloss.JoinVertical(
		lipgloss.Left,
		"", // Small top margin to show border
		content,
		"", // Small bottom spacing
		searchBar,
	)
}

// Helper function for min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// renderTaskTable renders the left panel with the task table
func (m ListModel) renderTaskTable(width int) string {
	var b strings.Builder
	
	// Table header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccentBright))
	
	b.WriteString(headerStyle.Render("📋 Tasks"))
	b.WriteString("\n\n")
	
	if len(m.tasks) == 0 {
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSecondaryText)).
			Italic(true)
		b.WriteString(emptyStyle.Render("No tasks found"))
		return b.String()
	}
	
	// Table column headers
	columnHeaderStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(ColorAccentBright)).
		Padding(0, 1)
	
	// Calculate column widths for the available space (subtract borders and padding)
	availableWidth := width - 4 // Account for borders
	idWidth := 4
	statusWidth := 8  // Increased for "✓ done"
	dueWidth := 10    // Keep at 10 for "TOMORROW"/"OVERDUE"
	titleWidth := availableWidth - idWidth - statusWidth - dueWidth - 6 // Account for spacing between columns
	
	// Ensure minimum widths
	if titleWidth < 20 {
		titleWidth = 20
	}
	if dueWidth < 10 {
		dueWidth = 10
	}
	
	// Column headers - simple, no borders
	headers := fmt.Sprintf("%-*s %-*s %-*s %-*s", 
		idWidth, "ID",
		titleWidth, "TITLE", 
		statusWidth, "STATUS",
		dueWidth, "DUE")
	b.WriteString(columnHeaderStyle.Render(headers))
	b.WriteString("\n\n")
	
	// Calculate visible tasks for current page  
	startIndex := m.currentPage * m.tasksPerPage
	endIndex := min(startIndex+m.tasksPerPage, len(m.tasks))
	
	// Render task rows
	for i := startIndex; i < endIndex; i++ {
		task := m.tasks[i]
		isSelected := i == m.selectedTask
		
		// Format columns
		id := fmt.Sprintf("#%d", task.ID)
		
		title := task.Title
		if len(title) > titleWidth-1 {
			if titleWidth > 4 {
				title = title[:titleWidth-4] + "..."
			} else {
				title = title[:titleWidth-1]
			}
		}
		
		// Apply shimmer to selected task title
		if isSelected {
			title = m.shimmer.RenderShimmerText(title, titleWidth)
		}
		
		// Format status text (always plain text for consistent column alignment)
		var statusText string
		if task.Status == "done" {
			statusText = "✓ done"
		} else {
			statusText = "○ todo"
		}
		
		// Format due date text (always plain text for consistent column alignment)
		var dueText string
		if task.Due != nil {
			now := time.Now()
			days := int(task.Due.Sub(now).Hours() / 24)
			if days < 0 {
				dueText = "OVERDUE"
			} else if days == 0 {
				dueText = "TODAY"
			} else if days == 1 {
				dueText = "TOMORROW"
			} else if days <= 7 {
				dueText = fmt.Sprintf("%dd", days)
			} else {
				dueText = task.Due.Format("02/01")
			}
		} else {
			dueText = "-"
		}
		
		// Ensure due text fits in column
		if len(dueText) > dueWidth {
			if dueWidth > 3 {
				dueText = dueText[:dueWidth-3] + "..."
			} else {
				dueText = dueText[:dueWidth]
			}
		}
		
		// Apply colors to status and due date
		var coloredStatusText string
		if task.Status == "done" {
			coloredStatusText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Render(statusText)
		} else {
			coloredStatusText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondaryText)).Render(statusText)
		}
		
		var coloredDueText string
		if task.Due != nil {
			now := time.Now()
			days := int(task.Due.Sub(now).Hours() / 24)
			if days < 0 {
				coloredDueText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(dueText)
			} else if days == 0 {
				coloredDueText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(dueText)
			} else if days == 1 {
				coloredDueText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(dueText)
			} else if days <= 7 {
				coloredDueText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBright)).Render(dueText)
			} else {
				coloredDueText = dueText // No special color for far dates
			}
		} else {
			coloredDueText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDisabledText)).Render(dueText)
		}
		
		// Create row content
		rowContent := fmt.Sprintf("%-*s %-*s %-*s %-*s", 
			idWidth, id,
			titleWidth, title,
			statusWidth, coloredStatusText,
			dueWidth, coloredDueText)
		
		if isSelected {
			// Selected row: use shining purple border
			shimmerBorder := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccentMain)).
				Bold(true).
				Padding(0, 1)
			
			b.WriteString(shimmerBorder.Render(rowContent))
		} else {
			// Regular row: no borders, just content
			b.WriteString(" " + rowContent)
		}
		b.WriteString("\n")
	}
	
	
	// Pagination info
	if m.tasksPerPage < len(m.tasks) {
		totalPages := (len(m.tasks) + m.tasksPerPage - 1) / m.tasksPerPage
		pageInfo := fmt.Sprintf("Page %d/%d (%d tasks)", m.currentPage+1, totalPages, len(m.tasks))
		pageStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHelpText)).
			Align(lipgloss.Center).
			Width(width-2).
			MarginTop(1)
		b.WriteString(pageStyle.Render(pageInfo))
	}
	
	// Apply outer border
	outerBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Width(width)
		// Remove fixed height to let content determine size
	
	return outerBorderStyle.Render(b.String())
}

// renderTaskDetails renders the right panel with task details
func (m ListModel) renderTaskDetails(width int) string {
	var b strings.Builder
	
	if len(m.tasks) == 0 || m.selectedTask >= len(m.tasks) {
		// Empty state with logo
		logoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccentMain)).
			Bold(true).
			Align(lipgloss.Center).
			Width(width)
		b.WriteString(logoStyle.Render("wrok"))
		
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSecondaryText)).
			Italic(true).
			Align(lipgloss.Center).
			Width(width).
			MarginTop(2)
		b.WriteString("\n")
		b.WriteString(emptyStyle.Render("Select a task to view details"))
	} else {
		// Show selected task details
		task := m.tasks[m.selectedTask]
		
		// Title
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorPrimaryText)).
			Width(width)
		b.WriteString(titleStyle.Render("📋 " + task.Title))
		b.WriteString("\n\n")
		
		// Status
		statusColor := ColorSecondaryText
		if task.Status == "done" {
			statusColor = ColorSuccess
		}
		statusStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(statusColor)).
			Bold(true)
		b.WriteString("Status: ")
		b.WriteString(statusStyle.Render(task.Status))
		b.WriteString("\n")
		
		// Project
		if task.Project != "" {
			b.WriteString("Project: ")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBright)).Render(task.Project))
			b.WriteString("\n")
		}
		
		// Priority
		if task.Priority > 0 {
			priorities := []string{"", "low", "medium", "high"}
			priorityStr := priorities[task.Priority]
			priorityColor := ColorSecondaryText
			if task.Priority == 3 {
				priorityColor = ColorError
			} else if task.Priority == 2 {
				priorityColor = ColorWarning
			}
			b.WriteString("Priority: ")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor)).Render(priorityStr))
			b.WriteString("\n")
		}
		
		// Tags
		if len(task.Tags) > 0 {
			var tagNames []string
			for _, tag := range task.Tags {
				tagNames = append(tagNames, tag.Name)
			}
			b.WriteString("Tags: ")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBright)).Render(strings.Join(tagNames, ", ")))
			b.WriteString("\n")
		}
		
		// JIRA
		if task.JiraID != "" {
			b.WriteString("JIRA: ")
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentMain)).Render(task.JiraID))
			b.WriteString("\n")
		}
		
		// Due date
		if task.Due != nil {
			b.WriteString("Due: ")
			// TODO: Format due date nicely
			b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(task.Due.Format("02/01/2006")))
			b.WriteString("\n")
		}
		
		// Notes
		if task.Note != "" {
			b.WriteString("\nNotes:\n")
			noteStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSecondaryText)).
				Italic(true).
				Width(width - 2)
			b.WriteString(noteStyle.Render(task.Note))
		}
	}
	
	// Apply border
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorBorder)).
		Width(width)
		// Remove fixed height to let content determine size
	
	return borderStyle.Render(b.String())
}

// renderSearchBar renders the search bar when active
func (m ListModel) renderSearchBar() string {
	searchStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorPrimaryText)).
		Background(lipgloss.Color(ColorBorder)).
		Padding(0, 1).
		Width(m.width - 2)
		
	prompt := "Search: " + m.searchQuery + "█"
	return searchStyle.Render(prompt)
}

// renderHelpBar renders the help bar with hotkey hints
func (m ListModel) renderHelpBar() string {
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(ColorHelpText)).
		Italic(true).
		Align(lipgloss.Center).
		Width(m.width)
		
	helpText := "↑/↓ nav · ←/→ page · / search · e edit · d done · s start/stop · q/esc quit"
	return helpStyle.Render(helpText)
}