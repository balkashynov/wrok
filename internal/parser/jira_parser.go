package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// NormalizeJiraID normalizes JIRA ticket IDs to uppercase XXX-111 format
// Accepts formats like:
// - "APP-123", "app-123" -> "APP-123"
// - "PROJ-42", "proj-42" -> "PROJ-42"
// Returns error if format is invalid
func NormalizeJiraID(jiraID string) (string, error) {
	if jiraID == "" {
		return "", nil
	}
	
	// Remove whitespace and convert to uppercase
	jiraID = strings.ToUpper(strings.TrimSpace(jiraID))
	
	// Validate format: letters, dash, numbers
	jiraRegex := regexp.MustCompile(`^([A-Z]+)-(\d+)$`)
	if !jiraRegex.MatchString(jiraID) {
		return "", fmt.Errorf("invalid JIRA format. Use: XXX-111 (letters-numbers)")
	}
	
	return jiraID, nil
}

// IsValidJiraFormat checks if a string matches JIRA ID format
func IsValidJiraFormat(jiraID string) bool {
	if jiraID == "" {
		return true // Empty is valid (optional field)
	}
	
	// Convert to uppercase for validation
	jiraID = strings.ToUpper(strings.TrimSpace(jiraID))
	
	jiraRegex := regexp.MustCompile(`^([A-Z]+)-(\d+)$`)
	return jiraRegex.MatchString(jiraID)
}