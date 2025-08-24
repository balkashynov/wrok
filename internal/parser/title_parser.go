package parser

import (
	"regexp"
	"strings"
	"time"
)

// ParsedTask represents a task parsed from natural language
type ParsedTask struct {
	Title    string
	Project  string
	Tags     []string
	Priority string
	JiraID   string
	DueDate  *time.Time
	Errors   []string
}

// ParseTitle extracts metadata from a task title using natural syntax
// Syntax: "Task title #tag1,tag2 @project +priority JIRA-123 due:3days"
func ParseTitle(input string) ParsedTask {
	result := ParsedTask{
		Title:  input,
		Tags:   []string{},
		Errors: []string{},
	}

	// Extract JIRA tickets (pattern: XXX-123)
	jiraRegex := regexp.MustCompile(`\b([A-Za-z]+)-(\d+)\b`)
	jiraMatches := jiraRegex.FindAllString(input, -1)
	if len(jiraMatches) > 0 {
		// Normalize JIRA ID to uppercase
		normalizedJira, err := NormalizeJiraID(jiraMatches[0])
		if err != nil {
			result.Errors = append(result.Errors, "Invalid JIRA ID format: "+jiraMatches[0])
		} else {
			result.JiraID = normalizedJira
		}
		// Remove from title
		input = jiraRegex.ReplaceAllString(input, "")
	}

	// Extract tags (#tag1,tag2 or #tag1 #tag2)
	tagRegex := regexp.MustCompile(`#([a-zA-Z0-9_,-]+)`)
	tagMatches := tagRegex.FindAllStringSubmatch(input, -1)
	for _, match := range tagMatches {
		if len(match) > 1 {
			// Split by comma in case of #tag1,tag2
			tagGroup := strings.Split(match[1], ",")
			for _, tag := range tagGroup {
				tag = strings.TrimSpace(tag)
				if tag != "" {
					result.Tags = append(result.Tags, tag)
				}
			}
		}
	}
	// Remove from title
	input = tagRegex.ReplaceAllString(input, "")

	// Extract project (@project-name)
	projectRegex := regexp.MustCompile(`@([a-zA-Z0-9_-]+)`)
	projectMatches := projectRegex.FindStringSubmatch(input)
	if len(projectMatches) > 1 {
		result.Project = projectMatches[1]
		// Remove from title
		input = projectRegex.ReplaceAllString(input, "")
	}

	// Extract priority (+high, +3, +medium, etc.)
	priorityRegex := regexp.MustCompile(`\+([a-zA-Z0-9]+)`)
	priorityMatches := priorityRegex.FindStringSubmatch(input)
	if len(priorityMatches) > 1 {
		priority := strings.ToLower(priorityMatches[1])
		if isValidPriority(priority) {
			result.Priority = priority
		} else {
			result.Errors = append(result.Errors, "Invalid priority '"+priorityMatches[1]+"'. Use: low, medium, high, 1, 2, or 3")
		}
		// Remove from title
		input = priorityRegex.ReplaceAllString(input, "")
	}

	// Extract due date (due:3days, due:15/12/2024, etc.)
	dueRegex := regexp.MustCompile(`due:([^\s]+)`)
	dueMatches := dueRegex.FindStringSubmatch(input)
	if len(dueMatches) > 1 {
		dueDate, err := ParseDueDate(dueMatches[1])
		if err != nil {
			result.Errors = append(result.Errors, "Invalid due date '"+dueMatches[1]+"': "+err.Error())
		} else {
			result.DueDate = dueDate
		}
		// Remove from title
		input = dueRegex.ReplaceAllString(input, "")
	}

	// Clean up the title (remove extra spaces)
	result.Title = strings.Join(strings.Fields(input), " ")
	result.Title = strings.TrimSpace(result.Title)

	return result
}

// isValidPriority checks if a priority value is valid
func isValidPriority(priority string) bool {
	validPriorities := map[string]bool{
		"low":    true,
		"medium": true,
		"med":    true,
		"high":   true,
		"1":      true,
		"2":      true,
		"3":      true,
	}
	return validPriorities[priority]
}

// NormalizePriority converts priority to standard form
func NormalizePriority(priority string) string {
	priority = strings.ToLower(strings.TrimSpace(priority))
	switch priority {
	case "1", "low":
		return "low"
	case "2", "medium", "med":
		return "medium"
	case "3", "high":
		return "high"
	default:
		return "low"
	}
}

// PriorityToInt converts priority string to integer
func PriorityToInt(priority string) int {
	switch NormalizePriority(priority) {
	case "low":
		return 1
	case "medium":
		return 2
	case "high":
		return 3
	default:
		return 1
	}
}