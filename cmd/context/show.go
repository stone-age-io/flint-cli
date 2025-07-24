package context

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"flint-cli/internal/config"
	"gopkg.in/yaml.v3"
)

var showOutputFormat string

var showCmd = &cobra.Command{
	Use:   "show [n]",
	Short: "Show detailed context configuration",
	Long: `Display detailed configuration for a specific context or the active context.

If no context name is provided, shows the currently active context.
The output format can be controlled with the --output flag.

Examples:
  flint context show                    # Show active context
  flint context show production         # Show specific context
  flint context show prod --output yaml # Show in YAML format`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateConfigManager(); err != nil {
			return err
		}

		var contextName string
		var ctx *config.Context
		var err error

		// Determine which context to show
		if len(args) == 0 {
			// Show active context
			ctx, err = configManager.GetActiveContext()
			if err != nil {
				return fmt.Errorf("no active context set. Use 'flint context select <n>' to set one")
			}
			contextName = ctx.Name
		} else {
			// Show specified context
			contextName = args[0]
			ctx, err = configManager.LoadContext(contextName)
			if err != nil {
				// Try to provide helpful suggestions
				contexts, listErr := configManager.ListContexts()
				if listErr == nil && len(contexts) > 0 {
					return fmt.Errorf("context '%s' not found. Available contexts: %v", 
						contextName, contexts)
				}
				return fmt.Errorf("context '%s' not found", contextName)
			}
		}

		// Check if it's the active context
		globalConfig, err := configManager.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		isActive := globalConfig.ActiveContext == contextName

		// Create a display version of the context (hide sensitive data)
		displayCtx := *ctx
		if displayCtx.PocketBase.AuthToken != "" {
			displayCtx.PocketBase.AuthToken = "***HIDDEN***"
		}
		if displayCtx.NATS.Password != "" {
			displayCtx.NATS.Password = "***HIDDEN***"
		}
		if displayCtx.NATS.Token != "" {
			displayCtx.NATS.Token = "***HIDDEN***"
		}

		// Output based on format
		switch strings.ToLower(showOutputFormat) {
		case "json":
			output, err := json.MarshalIndent(displayCtx, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal context to JSON: %w", err)
			}
			fmt.Println(string(output))

		case "yaml":
			output, err := yaml.Marshal(displayCtx)
			if err != nil {
				return fmt.Errorf("failed to marshal context to YAML: %w", err)
			}
			fmt.Print(string(output))

		case "table", "":
			// Default table format
			showContextTable(ctx, isActive)

		default:
			return fmt.Errorf("invalid output format '%s'. Valid formats: json, yaml, table", 
				showOutputFormat)
		}

		return nil
	},
}

func showContextTable(ctx *config.Context, isActive bool) {
	green := color.New(color.FgGreen).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	cyan := color.New(color.FgCyan).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	// Header
	fmt.Printf("%s Context: %s", bold("Stone-Age.io"), cyan(ctx.Name))
	if isActive {
		fmt.Printf(" %s", green("(ACTIVE)"))
	}
	fmt.Println()
	fmt.Println(strings.Repeat("=", 50))

	// PocketBase Configuration
	fmt.Printf("%s\n", bold("PocketBase Configuration:"))
	fmt.Printf("  URL:                %s\n", ctx.PocketBase.URL)
	fmt.Printf("  Auth Collection:    %s\n", ctx.PocketBase.AuthCollection)
	
	if ctx.PocketBase.OrganizationID != "" {
		fmt.Printf("  Organization ID:    %s\n", ctx.PocketBase.OrganizationID)
	} else {
		fmt.Printf("  Organization ID:    %s\n", yellow("Not Set"))
	}

	// Authentication status
	if ctx.PocketBase.AuthToken != "" {
		if ctx.PocketBase.AuthExpires != nil {
			fmt.Printf("  Authentication:     %s (expires %s)\n", 
				green("Valid"), 
				ctx.PocketBase.AuthExpires.Format("2006-01-02 15:04:05"))
		} else {
			fmt.Printf("  Authentication:     %s\n", green("Valid"))
		}
	} else {
		fmt.Printf("  Authentication:     %s\n", yellow("Not Authenticated"))
	}

	// Available collections
	fmt.Printf("  Available Collections: %d\n", len(ctx.PocketBase.AvailableCollections))
	if len(ctx.PocketBase.AvailableCollections) > 0 {
		fmt.Printf("    %s\n", strings.Join(ctx.PocketBase.AvailableCollections, ", "))
	}

	fmt.Println()

	// NATS Configuration
	fmt.Printf("%s\n", bold("NATS Configuration:"))
	fmt.Printf("  Servers:            %s\n", strings.Join(ctx.NATS.Servers, ", "))
	fmt.Printf("  Auth Method:        %s\n", ctx.NATS.AuthMethod)
	fmt.Printf("  TLS Enabled:        %t\n", ctx.NATS.TLSEnabled)
	fmt.Printf("  TLS Verify:         %t\n", ctx.NATS.TLSVerify)

	// Auth-specific details
	switch ctx.NATS.AuthMethod {
	case config.NATSAuthUserPass:
		if ctx.NATS.Username != "" {
			fmt.Printf("  Username:           %s\n", ctx.NATS.Username)
			if ctx.NATS.Password != "" {
				fmt.Printf("  Password:           %s\n", green("Set"))
			} else {
				fmt.Printf("  Password:           %s\n", yellow("Not Set"))
			}
		} else {
			fmt.Printf("  Credentials:        %s\n", yellow("Not Configured"))
		}

	case config.NATSAuthToken:
		if ctx.NATS.Token != "" {
			fmt.Printf("  Token:              %s\n", green("Set"))
		} else {
			fmt.Printf("  Token:              %s\n", yellow("Not Set"))
		}

	case config.NATSAuthCreds:
		if ctx.NATS.CredsFile != "" {
			fmt.Printf("  Credentials File:   %s\n", ctx.NATS.CredsFile)
		} else {
			fmt.Printf("  Credentials File:   %s\n", yellow("Not Set"))
		}
	}

	fmt.Println()

	// Organization details from auth record
	if authRecord, ok := ctx.PocketBase.AuthRecord["organizations"]; ok {
		if orgs, ok := authRecord.([]interface{}); ok && len(orgs) > 0 {
			fmt.Printf("%s\n", bold("Organization Details:"))
			for i, orgInterface := range orgs {
				if org, ok := orgInterface.(map[string]interface{}); ok {
					fmt.Printf("  Organization %d:\n", i+1)
					if name, ok := org["name"].(string); ok {
						fmt.Printf("    Name:           %s\n", name)
					}
					if id, ok := org["id"].(string); ok {
						fmt.Printf("    ID:             %s\n", id)
					}
					if desc, ok := org["description"].(string); ok && desc != "" {
						fmt.Printf("    Description:    %s\n", desc)
					}
				}
			}
			fmt.Println()
		}
	}

	// Show helpful commands
	if !isActive {
		fmt.Printf("%s\n", bold("Commands:"))
		fmt.Printf("  Select this context: %s\n", 
			cyan(fmt.Sprintf("flint context select %s", ctx.Name)))
	} else if ctx.PocketBase.AuthToken == "" {
		fmt.Printf("%s\n", bold("Next Steps:"))
		fmt.Printf("  Authenticate: %s\n", cyan("flint auth pb"))
		if ctx.PocketBase.OrganizationID == "" {
			fmt.Printf("  Set organization: %s\n", cyan("flint context organization <org_id>"))
		}
	}
}

func init() {
	showCmd.Flags().StringVarP(&showOutputFormat, "output", "o", "table", 
		"Output format (table|json|yaml)")
}
