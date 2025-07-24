package utils

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
)

// DisplayEmptyState shows a consistent empty state message with suggestions
func DisplayEmptyState(resource, suggestion string) {
	gray := color.New(color.FgHiBlack).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	
	fmt.Printf("%s No %s found.\n", gray("ℹ"), resource)
	
	if suggestion != "" {
		fmt.Printf("\nSuggestion: %s\n", cyan(suggestion))
	}
}

// DisplaySuccessWithDetails shows a consistent success message with details
func DisplaySuccessWithDetails(action, resource, id, name string) {
	green := color.New(color.FgGreen).SprintFunc()
	
	fmt.Printf("%s %s %s successfully!\n", green("✓"), 
		strings.Title(resource), strings.ToLower(action))
	
	if id != "" {
		fmt.Printf("  ID: %s\n", id)
	}
	if name != "" {
		fmt.Printf("  Name: %s\n", name)
	}
}

// DisplayWarningBanner shows a consistent warning banner
func DisplayWarningBanner(title, message string) {
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()
	
	fmt.Printf("%s %s\n", yellow("⚠"), bold(title))
	if message != "" {
		// Indent the message
		lines := strings.Split(message, "\n")
		for _, line := range lines {
			if strings.TrimSpace(line) != "" {
				fmt.Printf("  %s\n", line)
			}
		}
	}
}

// FormatTableTitle formats a title for table displays
func FormatTableTitle(title string, current, total int) string {
	if total == 0 {
		return fmt.Sprintf("%s (empty)", TitleCase(title))
	}
	
	if current == total {
		return fmt.Sprintf("%s (%d total)", TitleCase(title), total)
	}
	
	return fmt.Sprintf("%s (%d of %d)", TitleCase(title), current, total)
}

// FormatStatusBadge formats a status with appropriate coloring
func FormatStatusBadge(status string, isPositive bool) string {
	if isPositive {
		return color.New(color.FgGreen).Sprint(status)
	}
	return color.New(color.FgYellow).Sprint(status)
}

// FormatRecordIdentifier formats a record identifier for display
func FormatRecordIdentifier(id, name, email string) string {
	var parts []string
	
	if name != "" {
		parts = append(parts, name)
	}
	if email != "" {
		parts = append(parts, fmt.Sprintf("<%s>", email))
	}
	if len(parts) == 0 && id != "" {
		parts = append(parts, TruncateString(id, 12))
	}
	
	return strings.Join(parts, " ")
}
