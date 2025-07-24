package context

import (
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"flint-cli/internal/config"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available contexts",
	Long: `List all configured Stone-Age.io contexts with their status and configuration details.

The currently active context is highlighted with an asterisk (*).

Each context is stored in its own directory within the flint configuration directory,
containing the context configuration file and any related files like NATS credentials.

Examples:
  flint context list
  flint context ls`,
	Aliases: []string{"ls"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateConfigManager(); err != nil {
			return err
		}

		// Get all contexts
		contexts, err := configManager.ListContexts()
		if err != nil {
			return fmt.Errorf("failed to list contexts: %w", err)
		}

		if len(contexts) == 0 {
			fmt.Printf("No contexts configured in %s.\n", configManager.GetConfigDir())
			fmt.Printf("\nCreate your first context:\n  %s\n", 
				color.New(color.FgCyan).Sprint("flint context create <n> --pb-url <url> --nats-servers <servers>"))
			return nil
		}

		// Get active context
		globalConfig, err := configManager.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		// Process contexts and display
		displayContextsTable(contexts, globalConfig.ActiveContext)

		// Show active context summary
		if globalConfig.ActiveContext != "" {
			fmt.Printf("\nActive context: %s\n", 
				color.New(color.FgCyan).Sprint(globalConfig.ActiveContext))
		} else {
			fmt.Printf("\nNo active context set. Use %s to select one.\n", 
				color.New(color.FgCyan).Sprint("flint context select <n>"))
		}

		return nil
	},
}

// ContextDisplayInfo holds processed context information for display
type ContextDisplayInfo struct {
	Name          string
	Status        string
	PocketBaseURL string
	Organization  string
	NATSServers   string
	NATSAuth      string
	LastAuth      string
	IsActive      bool
	HasError      bool
}

// displayContextsTable processes contexts and displays them in a properly formatted table
func displayContextsTable(contextNames []string, activeContext string) {
	// Process all contexts first
	var contexts []ContextDisplayInfo
	for _, name := range contextNames {
		ctx := processContextForDisplay(name, activeContext)
		contexts = append(contexts, ctx)
	}

	// Create and configure table
	table := createContextTable()
	
	// Add rows to table
	for _, ctx := range contexts {
		table.Append([]string{
			ctx.Name,
			ctx.Status,
			ctx.PocketBaseURL,
			ctx.Organization,
			ctx.NATSServers,
			ctx.NATSAuth,
			ctx.LastAuth,
		})
	}

	fmt.Printf("Stone-Age.io Contexts (stored in %s):\n", configManager.GetConfigDir())
	table.Render()
}

// processContextForDisplay loads and processes a single context for display
func processContextForDisplay(contextName, activeContext string) ContextDisplayInfo {
	ctx, err := configManager.LoadContext(contextName)
	if err != nil {
		return ContextDisplayInfo{
			Name:          contextName,
			Status:        color.New(color.FgRed).Sprint("ERROR"),
			PocketBaseURL: "N/A",
			Organization:  "N/A",
			NATSServers:   "N/A",
			NATSAuth:      "N/A",
			LastAuth:      "N/A",
			HasError:      true,
		}
	}

	isActive := activeContext == contextName
	
	return ContextDisplayInfo{
		Name:          formatContextName(contextName, isActive),
		Status:        formatContextStatus(ctx, isActive),
		PocketBaseURL: formatPocketBaseURL(ctx.PocketBase.URL),
		Organization:  formatOrganization(ctx),
		NATSServers:   formatNATSServers(ctx.NATS.Servers),
		NATSAuth:      formatNATSAuth(ctx.NATS.AuthMethod),
		LastAuth:      formatLastAuth(ctx),
		IsActive:      isActive,
		HasError:      false,
	}
}

// createContextTable creates and configures the table with proper column settings
func createContextTable() *tablewriter.Table {
	table := tablewriter.NewWriter(os.Stdout)
	
	// Set headers with NATS servers column
	table.SetHeader([]string{"NAME", "STATUS", "POCKETBASE URL", "ORGANIZATION", "NATS SERVERS", "NATS AUTH", "LAST AUTH"})
	
	// Configure table appearance - no borders for clean look
	table.SetBorder(false)
	table.SetHeaderLine(false)
	table.SetRowSeparator("")
	table.SetCenterSeparator("")
	table.SetColumnSeparator("  ")
	table.SetTablePadding("  ")
	table.SetNoWhiteSpace(false)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	
	// Prevent text wrapping and set wider table width
	table.SetColWidth(150) // Increased from 120 to accommodate NATS servers
	table.SetAutoWrapText(false) // Critical: disable auto-wrapping
	
	// Set minimum column widths for better formatting
	table.SetColMinWidth(0, 12)  // NAME column
	table.SetColMinWidth(1, 18)  // STATUS column
	table.SetColMinWidth(2, 30)  // POCKETBASE URL column
	table.SetColMinWidth(3, 20)  // ORGANIZATION column
	table.SetColMinWidth(4, 25)  // NATS SERVERS column
	table.SetColMinWidth(5, 10)  // NATS AUTH column
	table.SetColMinWidth(6, 12)  // LAST AUTH column
	
	return table
}

// formatContextName formats the context name with active indicator
func formatContextName(name string, isActive bool) string {
	if isActive {
		return color.New(color.FgCyan).Sprint("* " + name)
	}
	return name
}

// formatContextStatus formats the authentication status
func formatContextStatus(ctx *config.Context, isActive bool) string {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	
	hasAuth := ctx.PocketBase.AuthToken != ""
	
	if isActive && hasAuth {
		return green("Active & Authenticated")
	} else if isActive && !hasAuth {
		return yellow("Active (Not Authenticated)")
	} else if !isActive && hasAuth {
		return green("Authenticated")
	} else {
		return yellow("Not Authenticated")
	}
}

// formatPocketBaseURL formats the PocketBase URL for display
func formatPocketBaseURL(url string) string {
	// Increased truncation length for better readability
	if len(url) > 40 {
		return url[:37] + "..."
	}
	return url
}

// formatOrganization formats organization information for display
func formatOrganization(ctx *config.Context) string {
	if ctx.PocketBase.OrganizationID == "" {
		return color.New(color.FgYellow).Sprint("Not Set")
	}

	orgID := ctx.PocketBase.OrganizationID
	
	// Try to get organization name from auth record
	orgName := extractOrganizationName(ctx.PocketBase.AuthRecord)
	
	if orgName != "" {
		// Show name with abbreviated ID: "ACME Corp (abc123...)"
		shortID := truncateString(orgID, 8)
		return fmt.Sprintf("%s (%s...)", orgName, shortID)
	}

	// Just show abbreviated ID if no name available
	return truncateString(orgID, 18) // Increased from 15 for better readability
}

// formatNATSServers formats NATS server information for display
func formatNATSServers(servers []string) string {
	if len(servers) == 0 {
		return color.New(color.FgHiBlack).Sprint("not configured")
	}

	if len(servers) == 1 {
		// Single server - show with reasonable truncation
		server := servers[0]
		if len(server) > 30 {
			return server[:27] + "..."
		}
		return server
	}

	// Multiple servers - show first one with count indicator
	firstServer := servers[0]
	
	// Extract just the hostname part for cleaner display
	cleanServer := cleanServerURL(firstServer)
	
	if len(cleanServer) > 20 {
		cleanServer = cleanServer[:17] + "..."
	}
	
	return fmt.Sprintf("%s (+%d more)", cleanServer, len(servers)-1)
}

// formatNATSAuth formats NATS authentication method
func formatNATSAuth(authMethod string) string {
	if authMethod == "" {
		return color.New(color.FgHiBlack).Sprint("not set")
	}
	return authMethod
}

// formatLastAuth formats the last authentication time
func formatLastAuth(ctx *config.Context) string {
	if ctx.PocketBase.AuthExpires == nil {
		return color.New(color.FgHiBlack).Sprint("Never")
	}
	return ctx.PocketBase.AuthExpires.Format("01-02 15:04")
}

// extractOrganizationName extracts organization name from auth record
func extractOrganizationName(authRecord map[string]interface{}) string {
	if authRecord == nil {
		return ""
	}

	orgData, exists := authRecord["organizations"]
	if !exists {
		return ""
	}

	orgs, ok := orgData.([]interface{})
	if !ok || len(orgs) == 0 {
		return ""
	}

	org, ok := orgs[0].(map[string]interface{})
	if !ok {
		return ""
	}

	name, ok := org["name"].(string)
	if !ok {
		return ""
	}

	return name
}

// cleanServerURL extracts hostname from NATS server URL for cleaner display
func cleanServerURL(serverURL string) string {
	// Remove protocol prefix
	url := serverURL
	if strings.HasPrefix(url, "nats://") {
		url = url[7:]
	} else if strings.HasPrefix(url, "tls://") {
		url = url[6:]
	}
	
	// Remove port if it's the default NATS port
	if strings.HasSuffix(url, ":4222") {
		url = url[:len(url)-5]
	}
	
	return url
}

// truncateString truncates a string to the specified length
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}
