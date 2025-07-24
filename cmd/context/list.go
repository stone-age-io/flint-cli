package context

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all available contexts",
	Long: `List all configured Stone-Age.io contexts with their status and configuration details.

The currently active context is highlighted with an asterisk (*).

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
			fmt.Println("No contexts configured.")
			fmt.Printf("\nCreate your first context:\n  %s\n", 
				color.New(color.FgCyan).Sprint("flint context create <name> --pb-url <url> --nats-servers <servers>"))
			return nil
		}

		// Get active context
		globalConfig, err := configManager.LoadGlobalConfig()
		if err != nil {
			return fmt.Errorf("failed to load global config: %w", err)
		}

		// Prepare table
		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Status", "PocketBase URL", "Organization", "NATS Auth", "Last Auth"})
		table.SetBorder(false)
		table.SetHeaderLine(false)
		table.SetRowSeparator("")
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetTablePadding("\t")
		table.SetNoWhiteSpace(true)

		// Colors
		green := color.New(color.FgGreen).SprintFunc()
		yellow := color.New(color.FgYellow).SprintFunc()
		red := color.New(color.FgRed).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()

		// Process each context
		for _, contextName := range contexts {
			ctx, err := configManager.LoadContext(contextName)
			if err != nil {
				// Handle error loading context
				status := red("ERROR")
				table.Append([]string{contextName, status, "N/A", "N/A", "N/A", "N/A"})
				continue
			}

			// Determine status
			var status string
			var nameDisplay string
			
			if globalConfig.ActiveContext == contextName {
				nameDisplay = cyan("* " + contextName)
				if ctx.PocketBase.AuthToken != "" {
					status = green("Active & Authenticated")
				} else {
					status = yellow("Active (Not Authenticated)")
				}
			} else {
				nameDisplay = contextName
				if ctx.PocketBase.AuthToken != "" {
					status = green("Authenticated")
				} else {
					status = yellow("Not Authenticated")
				}
			}

			// Organization info
			orgDisplay := "Not Set"
			if ctx.PocketBase.OrganizationID != "" {
				orgDisplay = ctx.PocketBase.OrganizationID
				// If we have auth record with organization info, use organization name
				if authRecord, ok := ctx.PocketBase.AuthRecord["organizations"]; ok {
					if orgs, ok := authRecord.([]interface{}); ok && len(orgs) > 0 {
						if org, ok := orgs[0].(map[string]interface{}); ok {
							if name, ok := org["name"].(string); ok && name != "" {
								orgDisplay = fmt.Sprintf("%s (%s)", name, ctx.PocketBase.OrganizationID)
							}
						}
					}
				}
			}

			// Last authentication time
			lastAuth := "Never"
			if ctx.PocketBase.AuthExpires != nil {
				lastAuth = ctx.PocketBase.AuthExpires.Format("2006-01-02 15:04")
			}

			table.Append([]string{
				nameDisplay,
				status,
				ctx.PocketBase.URL,
				orgDisplay,
				ctx.NATS.AuthMethod,
				lastAuth,
			})
		}

		fmt.Println("Stone-Age.io Contexts:")
		table.Render()

		// Show active context summary
		if globalConfig.ActiveContext != "" {
			fmt.Printf("\nActive context: %s\n", 
				cyan(globalConfig.ActiveContext))
		} else {
			fmt.Printf("\nNo active context set. Use %s to select one.\n", 
				color.New(color.FgCyan).Sprint("flint context select <name>"))
		}

		return nil
	},
}
