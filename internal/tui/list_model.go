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
	contentHeight := m.height - 6 // Reserve space for search/help bar
	
	// Left panel: Task table
	leftPanel := m.renderTaskTable(leftWidth)
	
	// Right panel: Task details
	rightPanel := m.renderTaskDetails(rightWidth, contentHeight)
	
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
		PaddingLeft(2)
		// Adjusted left padding for perfect alignment
	
	// Calculate fixed widths for right-aligned columns
	availableWidth := width - 4 // Account for borders
	idWidth := 4
	statusWidth := 8     // For "✓ done" / "○ todo"
	priorityWidth := 8   // For "high" / "med" / "low" / "-"
	jiraWidth := 9       // For "ABC-123" / "-" - reduced to 9 chars max
	dueWidth := 9        // For "TOMORROW" / "OVERDUE"
	
	// Responsive layout: hide priority, jira, due when terminal < 105 chars
	// The 'width' here is the left panel width (60% of terminal)
	// For 105px terminal breakpoint: 105 * 0.6 ≈ 63px table width
	showExtraColumns := width >= 63
	
	var headerLeft, headerRight string
	var rightSideWidth int
	var titleWidth int // Declare titleWidth at proper scope
	
	if showExtraColumns {
		// Full layout with all columns
		rightSideWidth = statusWidth + priorityWidth + jiraWidth + dueWidth + 3 // +3 for single spaces between 4 columns
		titleWidth = availableWidth - idWidth - rightSideWidth - 2 // -2 for spacing around title
		
		// Ensure minimum widths
		if titleWidth < 15 {
			titleWidth = 15
		}
		
		headerLeft = fmt.Sprintf("%-*s %-*s", idWidth, "ID", titleWidth, "TITLE")
		headerRight = fmt.Sprintf("%-*s %-*s %-*s %-*s", 
			statusWidth, "STATUS",
			priorityWidth, "PRIORITY",
			jiraWidth, "JIRA",
			dueWidth, "DUE")
	} else {
		// Compact layout: only ID, TITLE, STATUS
		rightSideWidth = statusWidth
		titleWidth = availableWidth - idWidth - rightSideWidth - 2 // -2 for spacing around title
		
		// Ensure minimum widths
		if titleWidth < 20 {
			titleWidth = 20 // More space for title in compact mode
		}
		
		headerLeft = fmt.Sprintf("%-*s %-*s", idWidth, "ID", titleWidth, "TITLE")
		headerRight = fmt.Sprintf("%-*s", statusWidth, "STATUS")
	}
	
	// Calculate spacing to push right side to the right
	spacingNeeded := availableWidth - len(headerLeft) - len(headerRight)
	if spacingNeeded < 1 {
		spacingNeeded = 1
	}
	
	headers := headerLeft + strings.Repeat(" ", spacingNeeded) + headerRight
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
		
		// Truncate ID if too long
		if len(id) > idWidth {
			if idWidth > 3 {
				id = id[:idWidth-3] + "..."
			} else {
				id = id[:idWidth]
			}
		}
		
		title := task.Title
		// Title truncation and shimmer will be applied later
		
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
		
		// Format and color priority
		var priorityText, coloredPriorityText string
		if task.Priority > 0 && task.Priority <= 3 {
			priorities := []string{"", "low", "med", "high"}
			priorityText = priorities[task.Priority]
			
			// Color coding: high=red, med=yellow, low=dim
			switch task.Priority {
			case 3: // high
				coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(priorityText)
			case 2: // medium
				coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(priorityText)
			case 1: // low
				coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondaryText)).Render(priorityText)
			}
		} else {
			priorityText = "-"
			coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDisabledText)).Render(priorityText)
		}
		
		// Format and color JIRA - ensure consistent styling
		var jiraText, coloredJiraText string
		if task.JiraID != "" {
			jiraText = task.JiraID
			// Apply consistent purple color to entire JIRA ID
			coloredJiraText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentMain)).Bold(true).Render(jiraText)
		} else {
			jiraText = "-"
			coloredJiraText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDisabledText)).Render(jiraText)
		}
		
		// Apply truncation AFTER creating colored versions to avoid overwriting
		// This will be done after the color assignments
		
		// TITLE: More conservative truncation to prevent layout breaking
		// Truncate to a safe maximum that won't overflow
		maxTitleLen := titleWidth - 2 // Leave some buffer to prevent overflow
		if maxTitleLen < 10 {
			maxTitleLen = 10 // Minimum reasonable title length
		}
		
		if !isSelected && len(title) > maxTitleLen {
			title = title[:maxTitleLen-3] + "..."
		} else if isSelected {
			// For selected items, truncate the original title first, then apply shimmer
			originalTitle := task.Title
			if len(originalTitle) > maxTitleLen {
				originalTitle = originalTitle[:maxTitleLen-3] + "..."
			}
			title = m.shimmer.RenderShimmerText(originalTitle, titleWidth)
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
		
		// NOW apply truncation to all fields (after colors are applied)
		
		// ID: 4 chars max
		if len(id) > 4 {
			id = id[:1] + "..."
		}
		
		// DUE: 9 chars max
		if len(dueText) > 9 {
			dueText = dueText[:6] + "..."
			// Re-apply color after truncation
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
		}
		
		// JIRA: 9 chars max (as requested)
		if len(jiraText) > 9 {
			jiraText = jiraText[:6] + "..."
			// Re-apply color after truncation
			if task.JiraID != "" {
				coloredJiraText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentMain)).Bold(true).Render(jiraText)
			} else {
				coloredJiraText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDisabledText)).Render(jiraText)
			}
		}
		
		// PRIORITY: 8 chars max
		if len(priorityText) > 8 {
			priorityText = priorityText[:5] + "..."
			// Re-apply color after truncation
			if task.Priority > 0 && task.Priority <= 3 {
				switch task.Priority {
				case 3: // high
					coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(priorityText)
				case 2: // medium
					coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(priorityText)
				case 1: // low
					coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondaryText)).Render(priorityText)
				}
			} else {
				coloredPriorityText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorDisabledText)).Render(priorityText)
			}
		}
		
		// STATUS: 8 chars max
		if len(statusText) > 8 {
			statusText = statusText[:5] + "..."
			// Re-apply color after truncation
			if task.Status == "done" {
				coloredStatusText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSuccess)).Render(statusText)
			} else {
				coloredStatusText = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondaryText)).Render(statusText)
			}
		}
		
		// Create row content with exact column alignment (responsive)
		// Add extra spaces to align values with headers
		rowLeft := fmt.Sprintf(" %-*s %-*s", idWidth, id, titleWidth, title)  // Added leading space
		
		var rowRight string
		if showExtraColumns {
			// Full layout
			rowRight = fmt.Sprintf("%-*s %-*s %-*s %-*s", 
				statusWidth, statusText,
				priorityWidth, priorityText,
				jiraWidth, jiraText,
				dueWidth, dueText)
		} else {
			// Compact layout: only status
			rowRight = fmt.Sprintf("%-*s", statusWidth, statusText)
		}
		
		// Calculate spacing to align right side (account for the extra space we added)
		spacingNeeded := availableWidth - len(rowLeft) - len(rowRight)
		if spacingNeeded < 1 {
			spacingNeeded = 1
		}
		
		// Combine with spacing
		plainRowContent := rowLeft + strings.Repeat(" ", spacingNeeded) + rowRight
		
		// Replace plain text with colored versions (responsive)
		rowContent := plainRowContent
		rowContent = strings.Replace(rowContent, statusText, coloredStatusText, 1)
		if showExtraColumns {
			// Only apply these replacements if columns are shown
			rowContent = strings.Replace(rowContent, priorityText, coloredPriorityText, 1)
			rowContent = strings.Replace(rowContent, jiraText, coloredJiraText, 1)
			rowContent = strings.Replace(rowContent, dueText, coloredDueText, 1)
		}
		
		if isSelected {
			// Selected row: custom text with ID, title, and non-null fields
			var customParts []string
			
			// Build parts with proper styling
			customParts = append(customParts, id)
			
			// Add title with shimmer effect (give it plenty of width)
			shimmeredTitle := m.shimmer.RenderShimmerText(task.Title, len(task.Title)+20) // Extra width for shimmer effect
			customParts = append(customParts, shimmeredTitle)
			
			// Add priority with same colors as table
			if task.Priority > 0 && task.Priority <= 3 {
				priorities := []string{"", "low", "med", "high"}
				priorityText := priorities[task.Priority]
				
				// Apply same color coding as table
				var coloredPriority string
				switch task.Priority {
				case 3: // high
					coloredPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(priorityText)
				case 2: // medium
					coloredPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(priorityText)
				case 1: // low
					coloredPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondaryText)).Render(priorityText)
				}
				customParts = append(customParts, coloredPriority)
			}
			
			// Add JIRA with same color as table
			if task.JiraID != "" {
				coloredJira := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentMain)).Bold(true).Render(task.JiraID)
				customParts = append(customParts, coloredJira)
			}
			
			// Add due date with same colors as table
			if task.Due != nil {
				now := time.Now()
				days := int(task.Due.Sub(now).Hours() / 24)
				var dueDisplay string
				var coloredDue string
				
				if days < 0 {
					dueDisplay = "OVERDUE"
					coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(dueDisplay)
				} else if days == 0 {
					dueDisplay = "TODAY"
					coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(dueDisplay)
				} else if days == 1 {
					dueDisplay = "TOMORROW"
					coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(dueDisplay)
				} else if days <= 7 {
					dueDisplay = fmt.Sprintf("%dd", days)
					coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBright)).Render(dueDisplay)
				} else {
					dueDisplay = task.Due.Format("02/01")
					coloredDue = dueDisplay // No special color for far dates
				}
				customParts = append(customParts, coloredDue)
			}
			
			// Smart truncation: JIRA first, then title
			maxWidth := availableWidth - 4 // Account for border + padding
			
			// Helper function to calculate visual length (without ANSI codes)
			visualLength := func(text string) int {
				// Count visible characters by removing ANSI escape sequences
				visibleLen := 0
				inEscape := false
				for _, r := range text {
					if r == '\033' { // ESC character
						inEscape = true
					} else if inEscape && r == 'm' {
						inEscape = false
					} else if !inEscape {
						visibleLen++
					}
				}
				return visibleLen
			}
			
			// Try different truncation strategies
			customText := ""
			
			// Strategy 1: No truncation
			customText = strings.Join(customParts, "   ")
			if visualLength(customText) <= maxWidth {
				// Perfect fit, done
			} else if task.JiraID != "" && len(task.JiraID) > 9 {
				// Strategy 2: Truncate JIRA first
				truncatedJira := task.JiraID[:6] + "..."
				truncatedParts := make([]string, 0)
				
				// Rebuild parts with truncated JIRA (with styling)
				truncatedParts = append(truncatedParts, id)
				
				// Add shimmered title
				shimmeredTitle := m.shimmer.RenderShimmerText(task.Title, len(task.Title)+20)
				truncatedParts = append(truncatedParts, shimmeredTitle)
				
				if task.Priority > 0 && task.Priority <= 3 {
					priorities := []string{"", "low", "med", "high"}
					priorityText := priorities[task.Priority]
					
					// Apply same color coding
					var coloredPriority string
					switch task.Priority {
					case 3: // high
						coloredPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(priorityText)
					case 2: // medium
						coloredPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(priorityText)
					case 1: // low
						coloredPriority = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorSecondaryText)).Render(priorityText)
					}
					truncatedParts = append(truncatedParts, coloredPriority)
				}
				
				// Add styled truncated JIRA
				coloredTruncatedJira := lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentMain)).Bold(true).Render(truncatedJira)
				truncatedParts = append(truncatedParts, coloredTruncatedJira)
				
				if task.Due != nil {
					now := time.Now()
					days := int(task.Due.Sub(now).Hours() / 24)
					var dueDisplay string
					var coloredDue string
					
					if days < 0 {
						dueDisplay = "OVERDUE"
						coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorError)).Render(dueDisplay)
					} else if days == 0 {
						dueDisplay = "TODAY"
						coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(dueDisplay)
					} else if days == 1 {
						dueDisplay = "TOMORROW"
						coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorWarning)).Render(dueDisplay)
					} else if days <= 7 {
						dueDisplay = fmt.Sprintf("%dd", days)
						coloredDue = lipgloss.NewStyle().Foreground(lipgloss.Color(ColorAccentBright)).Render(dueDisplay)
					} else {
						dueDisplay = task.Due.Format("02/01")
						coloredDue = dueDisplay // No special color for far dates
					}
					truncatedParts = append(truncatedParts, coloredDue)
				}
				
				customText = strings.Join(truncatedParts, "   ")
				if visualLength(customText) > maxWidth {
					// Strategy 3: Truncate title too
					overflow := visualLength(customText) - maxWidth + 3 // +3 for "..."
					if len(task.Title) > overflow + 10 { // Keep at least 10 chars
						truncatedTitle := task.Title[:len(task.Title)-overflow] + "..."
						// Apply shimmer to truncated title with proper width
					shimmeredTruncatedTitle := m.shimmer.RenderShimmerText(truncatedTitle, len(truncatedTitle))
					truncatedParts[1] = shimmeredTruncatedTitle // Title is at index 1 with shimmer
						customText = strings.Join(truncatedParts, "   ")
					} else {
						// Fallback: truncate entire string
						customText = customText[:maxWidth-3] + "..."
					}
				}
			} else {
				// Strategy 3: No JIRA to truncate, truncate title directly
				overflow := visualLength(customText) - maxWidth + 3 // +3 for "..."
				if len(task.Title) > overflow + 10 { // Keep at least 10 chars
					truncatedTitle := task.Title[:len(task.Title)-overflow] + "..."
					truncatedParts := make([]string, 0)
					
					// Rebuild with truncated title
					truncatedParts = append(truncatedParts, id)
					// Apply shimmer to truncated title with proper width
				shimmeredTruncatedTitle := m.shimmer.RenderShimmerText(truncatedTitle, len(truncatedTitle))
				truncatedParts = append(truncatedParts, shimmeredTruncatedTitle)
					
					if task.Priority > 0 && task.Priority <= 3 {
						priorities := []string{"", "low", "med", "high"}
						truncatedParts = append(truncatedParts, priorities[task.Priority])
					}
					
					if task.JiraID != "" {
						truncatedParts = append(truncatedParts, task.JiraID)
					}
					
					if task.Due != nil {
						now := time.Now()
						days := int(task.Due.Sub(now).Hours() / 24)
						var dueDisplay string
						if days < 0 {
							dueDisplay = "OVERDUE"
						} else if days == 0 {
							dueDisplay = "TODAY"
						} else if days == 1 {
							dueDisplay = "TOMORROW"
						} else if days <= 7 {
							dueDisplay = fmt.Sprintf("%dd", days)
						} else {
							dueDisplay = task.Due.Format("02/01")
						}
						truncatedParts = append(truncatedParts, dueDisplay)
					}
					
					customText = strings.Join(truncatedParts, "   ")
				} else {
					// Fallback: truncate entire string
					customText = customText[:maxWidth-3] + "..."
				}
			}
			
			shimmerBorder := lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorAccentMain)).
				Bold(true).
				Padding(0, 1)
			
			b.WriteString(shimmerBorder.Render(customText))
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

// renderTaskDetails renders the right panel with task details in purple box style
func (m ListModel) renderTaskDetails(width, height int) string {
	var b strings.Builder
	
	// Handle narrow terminals (< 110px total width) or short terminals (< 35px height) with simpler design
	if m.width < 110 || m.height < 35 {
		return m.renderNarrowTaskDetails(width, height)
	}
	
	if len(m.tasks) == 0 || m.selectedTask >= len(m.tasks) {
		// Empty state - center vertically with logo
		availableHeight := height - 4
		verticalPadding := availableHeight / 3
		if verticalPadding > 5 {
			verticalPadding = 5
		}
		
		// Add vertical spacing
		for i := 0; i < verticalPadding; i++ {
			b.WriteString("\n")
		}
		
		// ASCII logo
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
		
		emptyStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSecondaryText)).
			Italic(true).
			Align(lipgloss.Center).
			Width(width-8)
		b.WriteString(emptyStyle.Render("Select a task to view details"))
		
	} else {
		// Selected task - show details with some top padding
		task := m.tasks[m.selectedTask]
		
		b.WriteString("\n")
		
		// ASCII logo at top
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
		
		// Separator line
		separatorStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorBorder)).
			Align(lipgloss.Center).
			Width(width-8)
		separatorLine := strings.Repeat("─", min(width-12, 40))
		b.WriteString(separatorStyle.Render(separatorLine))
		b.WriteString("\n\n")
		
		// Title in bordered box
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
		
		// Task details in structured format
		// Status with emoji
		statusIcon := "○"
		statusColor := ColorSecondaryText
		statusText := "todo"
		if task.Status == "done" {
			statusIcon = "✅"
			statusColor = ColorSuccess
			statusText = "done"
		}
		
		statusStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(width-8)
		statusLine := fmt.Sprintf("%s Status: %s", statusIcon, 
			lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(statusText))
		b.WriteString(statusStyle.Render(statusLine))
		b.WriteString("\n")
		
		// Project with emoji
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
		
		// Priority with emoji
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
			lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor)).Bold(true).Render(priorityValue))
		b.WriteString(priorityStyle.Render(priorityLine))
		b.WriteString("\n")
		
		// Tags with emoji
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
		tagsLine := fmt.Sprintf("🔖  Tags: %s", 
			lipgloss.NewStyle().Foreground(lipgloss.Color(tagsColor)).Render(tagsValue))
		b.WriteString(tagsStyle.Render(tagsLine))
		b.WriteString("\n")
		
		// JIRA with emoji
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
			lipgloss.NewStyle().Foreground(lipgloss.Color(jiraColor)).Bold(true).Render(jiraValue))
		b.WriteString(jiraStyle.Render(jiraLine))
		b.WriteString("\n")
		
		// Due date with emoji
		dueStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(width-8)
		dueIcon := "📅"
		dueValue := "none"
		dueColor := ColorDisabledText
		if task.Due != nil {
			now := time.Now()
			days := int(task.Due.Sub(now).Hours() / 24)
			
			if days < 0 {
				dueIcon = "⚠️"
				dueValue = fmt.Sprintf("OVERDUE (%s)", task.Due.Format("02/01/2006"))
				dueColor = ColorError
			} else if days == 0 {
				dueIcon = "🚨"
				dueValue = fmt.Sprintf("TODAY (%s)", task.Due.Format("02/01/2006"))
				dueColor = ColorWarning
			} else if days == 1 {
				dueIcon = "📅"
				dueValue = fmt.Sprintf("TOMORROW (%s)", task.Due.Format("02/01/2006"))
				dueColor = ColorWarning
			} else if days <= 7 {
				dueIcon = "📅"
				dueValue = fmt.Sprintf("in %d days (%s)", days, task.Due.Format("02/01/2006"))
				dueColor = ColorAccentBright
			} else {
				dueIcon = "📅"
				dueValue = task.Due.Format("02/01/2006")
				dueColor = ColorSecondaryText
			}
		}
		dueLine := fmt.Sprintf("%s Due: %s", dueIcon, 
			lipgloss.NewStyle().Foreground(lipgloss.Color(dueColor)).Bold(true).Render(dueValue))
		b.WriteString(dueStyle.Render(dueLine))
		b.WriteString("\n\n")
		
		// Notes section with emoji
		notesStyle := lipgloss.NewStyle().
			Align(lipgloss.Center).
			Width(width-8)
		b.WriteString(notesStyle.Render("📝 Notes:"))
		b.WriteString("\n")
		
		if task.Note != "" {
			noteStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSecondaryText)).
				Width(width - 12).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Align(lipgloss.Left).
				Padding(1, 2)
			b.WriteString(noteStyle.Render(task.Note))
		} else {
			emptyNoteStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorDisabledText)).
				Italic(true).
				Width(width - 12).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Align(lipgloss.Center).
				Padding(1, 2)
			b.WriteString(emptyNoteStyle.Render("No notes"))
		}
		
		// Timestamps at bottom
		b.WriteString("\n\n")
		timestampStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHelpText)).
			Align(lipgloss.Center).
			Width(width-8)
		timestamps := fmt.Sprintf("Created: %s • Updated: %s", 
			task.CreatedAt.Format("02/01 15:04"), 
			task.UpdatedAt.Format("02/01 15:04"))
		b.WriteString(timestampStyle.Render(timestamps))
	}
	
	// Apply full-width purple border with space for the border
	purpleBoxStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccentMain)).
		Width(width - 4).  // Leave more space for the right border
		Padding(1, 2)
	
	return purpleBoxStyle.Render(b.String())
}

// renderNarrowTaskDetails renders a compact view for narrow terminals (< 110px)
func (m ListModel) renderNarrowTaskDetails(width, height int) string {
	var b strings.Builder
	
	if len(m.tasks) == 0 || m.selectedTask >= len(m.tasks) {
		// Empty state for narrow terminals
		availableHeight := height - 4
		verticalPadding := availableHeight / 3
		if verticalPadding > 3 {
			verticalPadding = 3
		}
		
		// Add vertical spacing
		for i := 0; i < verticalPadding; i++ {
			b.WriteString("\n")
		}
		
		// Simple compact logo
		compactLogoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccentMain)).
			Bold(true).
			Align(lipgloss.Center).
			Width(width-8)
		
		b.WriteString(compactLogoStyle.Render("◈ WROK ◈"))
		b.WriteString("\n\n")
		
		// Tip message
		tipStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSecondaryText)).
			Italic(true).
			Align(lipgloss.Center).
			Width(width-8)
		b.WriteString(tipStyle.Render("💡 Stretch terminal"))
		b.WriteString("\n")
		b.WriteString(tipStyle.Render("for better experience"))
		
	} else {
		// Selected task - compact view
		task := m.tasks[m.selectedTask]
		
		b.WriteString("\n")
		
		// Simple compact header
		headerStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorAccentMain)).
			Bold(true).
			Align(lipgloss.Center).
			Width(width-8)
		
		b.WriteString(headerStyle.Render("◈ TASK DETAILS ◈"))
		b.WriteString("\n\n")
		
		// Task title - compact
		titleStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(ColorPrimaryText)).
			Align(lipgloss.Center).
			Width(width-8)
		
		// Truncate title if too long for narrow view
		title := task.Title
		if len(title) > width-12 {
			title = title[:width-15] + "..."
		}
		b.WriteString(titleStyle.Render(title))
		b.WriteString("\n\n")
		
		// Compact field display - vertical layout
		fieldStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorSecondaryText)).
			Width(width-8)
		
		// Status
		statusIcon := "○"
		statusColor := ColorSecondaryText
		if task.Status == "done" {
			statusIcon = "✓"
			statusColor = ColorSuccess
		}
		statusLine := fmt.Sprintf("%s %s", statusIcon, 
			lipgloss.NewStyle().Foreground(lipgloss.Color(statusColor)).Bold(true).Render(task.Status))
		b.WriteString(fieldStyle.Render(statusLine))
		b.WriteString("\n")
		
		// Project (if exists)
		if task.Project != "" {
			projectLine := fmt.Sprintf("📁 %s", task.Project)
			if len(projectLine) > width-10 {
				projectLine = projectLine[:width-13] + "..."
			}
			b.WriteString(fieldStyle.Render(projectLine))
			b.WriteString("\n")
		}
		
		// Priority (if exists)
		if task.Priority > 0 && task.Priority <= 3 {
			priorities := []string{"", "low", "med", "high"}
			priorityText := priorities[task.Priority]
			var priorityColor string
			var priorityIcon string
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
			priorityLine := fmt.Sprintf("%s %s", priorityIcon, 
				lipgloss.NewStyle().Foreground(lipgloss.Color(priorityColor)).Bold(true).Render(priorityText))
			b.WriteString(fieldStyle.Render(priorityLine))
			b.WriteString("\n")
		}
		
		// Tags (if exist)
		if len(task.Tags) > 0 {
			var tagNames []string
			for _, tag := range task.Tags {
				tagNames = append(tagNames, "#"+tag.Name)
			}
			tagsText := strings.Join(tagNames, " ")
			if len(tagsText) > width-12 {
				tagsText = tagsText[:width-15] + "..."
			}
			tagsLine := fmt.Sprintf("🔖 %s", tagsText)
			b.WriteString(fieldStyle.Render(tagsLine))
			b.WriteString("\n")
		}
		
		// JIRA (if exists)
		if task.JiraID != "" {
			jiraLine := fmt.Sprintf("🎯 %s", task.JiraID)
			b.WriteString(fieldStyle.Render(jiraLine))
			b.WriteString("\n")
		}
		
		// Due date (if exists)
		if task.Due != nil {
			now := time.Now()
			days := int(task.Due.Sub(now).Hours() / 24)
			var dueText, dueColor string
			
			if days < 0 {
				dueText = "OVERDUE"
				dueColor = ColorError
			} else if days == 0 {
				dueText = "TODAY"
				dueColor = ColorWarning
			} else if days == 1 {
				dueText = "TOMORROW"
				dueColor = ColorWarning
			} else if days <= 7 {
				dueText = fmt.Sprintf("%dd", days)
				dueColor = ColorAccentBright
			} else {
				dueText = task.Due.Format("02/01")
				dueColor = ColorSecondaryText
			}
			
			dueLine := fmt.Sprintf("📅 %s", 
				lipgloss.NewStyle().Foreground(lipgloss.Color(dueColor)).Bold(true).Render(dueText))
			b.WriteString(fieldStyle.Render(dueLine))
			b.WriteString("\n")
		}
		
		// Notes (if exist) - very compact
		if task.Note != "" {
			b.WriteString("\n")
			noteHeaderStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSecondaryText)).
				Align(lipgloss.Center).
				Width(width-8)
			b.WriteString(noteHeaderStyle.Render("📝 Notes"))
			b.WriteString("\n")
			
			// Truncate notes for narrow view
			notes := task.Note
			maxNoteLen := (width - 10) * 3 // Allow up to 3 lines
			if len(notes) > maxNoteLen {
				notes = notes[:maxNoteLen-3] + "..."
			}
			
			compactNoteStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color(ColorSecondaryText)).
				Width(width - 10).
				Border(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color(ColorBorder)).
				Align(lipgloss.Left).
				Padding(0, 1)
			b.WriteString(compactNoteStyle.Render(notes))
		}
		
		// Tip at bottom
		b.WriteString("\n\n")
		tipStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(ColorHelpText)).
			Italic(true).
			Align(lipgloss.Center).
			Width(width-8)
		b.WriteString(tipStyle.Render("💡 Stretch for full view"))
	}
	
	// Apply compact purple border
	compactBorderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color(ColorAccentMain)).
		Width(width - 4).
		Padding(1, 2)
	
	return compactBorderStyle.Render(b.String())
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