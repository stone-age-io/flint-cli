package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"flint-cli/internal/config"
	"gopkg.in/yaml.v3"
)

// OutputData formats and prints data according to the specified format
func OutputData(data interface{}, format string) error {
	switch strings.ToLower(format) {
	case config.OutputFormatJSON, "":
		return outputJSON(data)
	case config.OutputFormatYAML:
		return outputYAML(data)
	case config.OutputFormatTable:
		return outputTable(data)
	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// outputJSON prints data in JSON format
func outputJSON(data interface{}) error {
	output, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}
	fmt.Println(string(output))
	return nil
}

// outputYAML prints data in YAML format
func outputYAML(data interface{}) error {
	output, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal YAML: %w", err)
	}
	fmt.Print(string(output))
	return nil
}

// outputTable prints data in table format
func outputTable(data interface{}) error {
	// This is a simplified table output - will be enhanced in future phases
	// when we implement collection-specific formatting
	switch v := data.(type) {
	case []map[string]interface{}:
		return outputMapSliceTable(v)
	case map[string]interface{}:
		return outputMapTable(v)
	default:
		// Fallback to JSON for complex types
		return outputJSON(data)
	}
}

// outputMapSliceTable outputs a slice of maps as a table
func outputMapSliceTable(data []map[string]interface{}) error {
	if len(data) == 0 {
		fmt.Println("No data found.")
		return nil
	}

	// Extract headers from first item
	var headers []string
	for key := range data[0] {
		headers = append(headers, key)
	}

	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)

	// Add rows
	for _, item := range data {
		var row []string
		for _, header := range headers {
			value := ""
			if v, ok := item[header]; ok && v != nil {
				value = fmt.Sprintf("%v", v)
			}
			row = append(row, value)
		}
		table.Append(row)
	}

	table.Render()
	return nil
}

// outputMapTable outputs a single map as a vertical table
func outputMapTable(data map[string]interface{}) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader([]string{"Field", "Value"})

	for key, value := range data {
		valueStr := ""
		if value != nil {
			valueStr = fmt.Sprintf("%v", value)
		}
		table.Append([]string{key, valueStr})
	}

	table.Render()
	return nil
}

// PrintError prints an error message with consistent formatting
func PrintError(err error) {
	if !config.Global.ColorsEnabled {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return
	}

	red := color.New(color.FgRed).SprintFunc()
	fmt.Fprintf(os.Stderr, "%s %v\n", red("Error:"), err)
}

// PrintWarning prints a warning message with consistent formatting
func PrintWarning(message string) {
	if !config.Global.ColorsEnabled {
		fmt.Fprintf(os.Stderr, "Warning: %s\n", message)
		return
	}

	yellow := color.New(color.FgYellow).SprintFunc()
	fmt.Fprintf(os.Stderr, "%s %s\n", yellow("Warning:"), message)
}

// PrintSuccess prints a success message with consistent formatting
func PrintSuccess(message string) {
	if !config.Global.ColorsEnabled {
		fmt.Printf("Success: %s\n", message)
		return
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s %s\n", green("✓"), message)
}

// PrintInfo prints an info message with consistent formatting
func PrintInfo(message string) {
	if !config.Global.ColorsEnabled {
		fmt.Printf("Info: %s\n", message)
		return
	}

	cyan := color.New(color.FgCyan).SprintFunc()
	fmt.Printf("%s %s\n", cyan("ℹ"), message)
}

// PrintDebug prints a debug message if debug mode is enabled
func PrintDebug(message string) {
	if !config.Global.Debug {
		return
	}

	if !config.Global.ColorsEnabled {
		fmt.Fprintf(os.Stderr, "Debug: %s\n", message)
		return
	}

	gray := color.New(color.FgHiBlack).SprintFunc()
	fmt.Fprintf(os.Stderr, "%s %s\n", gray("Debug:"), message)
}

// TruncateString truncates a string to the specified length with ellipsis
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// FormatDuration formats a duration string in a human-readable way
func FormatDuration(seconds int64) string {
	if seconds < 60 {
		return fmt.Sprintf("%ds", seconds)
	}
	if seconds < 3600 {
		return fmt.Sprintf("%dm %ds", seconds/60, seconds%60)
	}
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	return fmt.Sprintf("%dh %dm", hours, minutes)
}
