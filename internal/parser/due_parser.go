package parser

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// ParseDueDate parses various due date formats
// Supported formats:
// - dd/mm/yyyy (e.g., "15/12/2024")
// - X days (e.g., "3 days", "1 day")
// - X hours (e.g., "24 hours", "1 hour")
// - X weeks (e.g., "2 weeks", "1 week")
func ParseDueDate(input string) (*time.Time, error) {
	if input == "" {
		return nil, nil
	}
	
	input = strings.TrimSpace(input)
	
	// Try dd/mm/yyyy format first
	if dueDate, err := parseDateFormat(input); err == nil {
		return dueDate, nil
	}
	
	// Try relative time formats
	if dueDate, err := parseRelativeTime(input); err == nil {
		return dueDate, nil
	}
	
	return nil, fmt.Errorf("invalid date format. Use: dd/mm/yyyy, X days, X hours, or X weeks")
}

// parseDateFormat parses dd/mm/yyyy format
func parseDateFormat(input string) (*time.Time, error) {
	dateRegex := regexp.MustCompile(`^(\d{1,2})/(\d{1,2})/(\d{4})$`)
	matches := dateRegex.FindStringSubmatch(input)
	
	if len(matches) != 4 {
		return nil, fmt.Errorf("invalid date format")
	}
	
	day, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid day")
	}
	
	month, err := strconv.Atoi(matches[2])
	if err != nil {
		return nil, fmt.Errorf("invalid month")
	}
	
	year, err := strconv.Atoi(matches[3])
	if err != nil {
		return nil, fmt.Errorf("invalid year")
	}
	
	// Validate date ranges
	if day < 1 || day > 31 {
		return nil, fmt.Errorf("day must be between 1 and 31")
	}
	if month < 1 || month > 12 {
		return nil, fmt.Errorf("month must be between 1 and 12")
	}
	if year < 2024 || year > 2100 {
		return nil, fmt.Errorf("year must be between 2024 and 2100")
	}
	
	dueDate := time.Date(year, time.Month(month), day, 23, 59, 59, 0, time.Local)
	
	// Check if date is valid (handles leap years, etc.)
	if dueDate.Day() != day || dueDate.Month() != time.Month(month) || dueDate.Year() != year {
		return nil, fmt.Errorf("invalid date")
	}
	
	return &dueDate, nil
}

// parseRelativeTime parses relative time formats like "3 days", "24 hours", etc.
func parseRelativeTime(input string) (*time.Time, error) {
	input = strings.ToLower(input)
	
	// Regex for "X unit" or "X units"
	relativeRegex := regexp.MustCompile(`^(\d+)\s+(hour|hours|day|days|week|weeks)$`)
	matches := relativeRegex.FindStringSubmatch(input)
	
	if len(matches) != 3 {
		return nil, fmt.Errorf("invalid relative time format")
	}
	
	amount, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil, fmt.Errorf("invalid number")
	}
	
	unit := matches[2]
	now := time.Now()
	
	switch unit {
	case "hour", "hours":
		if amount < 1 || amount > 8760 { // Max 1 year in hours
			return nil, fmt.Errorf("hours must be between 1 and 8760")
		}
		dueDate := now.Add(time.Duration(amount) * time.Hour)
		return &dueDate, nil
		
	case "day", "days":
		if amount < 1 || amount > 365 { // Max 1 year in days
			return nil, fmt.Errorf("days must be between 1 and 365")
		}
		// Set to end of day (23:59:59) for the target date
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dueDate := today.AddDate(0, 0, amount).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		return &dueDate, nil
		
	case "week", "weeks":
		if amount < 1 || amount > 52 { // Max 1 year in weeks
			return nil, fmt.Errorf("weeks must be between 1 and 52")
		}
		// Set to end of day (23:59:59) for the target date
		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
		dueDate := today.AddDate(0, 0, amount*7).Add(23*time.Hour + 59*time.Minute + 59*time.Second)
		return &dueDate, nil
		
	default:
		return nil, fmt.Errorf("unsupported time unit")
	}
}

// FormatDueDate formats a due date for display
func FormatDueDate(dueDate *time.Time) string {
	if dueDate == nil {
		return ""
	}
	
	now := time.Now()
	
	// Calculate calendar days difference
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	dueDay := time.Date(dueDate.Year(), dueDate.Month(), dueDate.Day(), 0, 0, 0, 0, dueDate.Location())
	daysDiff := int(dueDay.Sub(today).Hours() / 24)
	
	// Always show the actual date to avoid confusion
	dateStr := dueDate.Format("02/01/2006")
	
	if daysDiff < 0 {
		// Overdue
		return fmt.Sprintf("⚠️ OVERDUE (%s)", dateStr)
	} else if daysDiff == 0 {
		// Due today
		return fmt.Sprintf("🔥 Due today (%s)", dateStr)
	} else if daysDiff == 1 {
		// Due tomorrow
		return fmt.Sprintf("📅 Due tomorrow (%s)", dateStr)
	} else if daysDiff <= 7 {
		// Due within a week  
		return fmt.Sprintf("📅 Due %s (in %d days)", dateStr, daysDiff)
	} else {
		// Due later
		return fmt.Sprintf("📅 Due %s", dateStr)
	}
}