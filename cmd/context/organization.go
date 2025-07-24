package context

import (
	"fmt"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var organizationCmd = &cobra.Command{
	Use:   "organization <organization_id>",
	Short: "Set the organization for the active context",
	Long: `Set the organization ID for the currently active context.

The organization ID determines which Stone-Age.io organization's resources
you can access. This setting is required for most operations and must match
an organization that your authenticated user belongs to.

When you set an organization, this will update your user's current_organization_id
in PocketBase to ensure API rules are properly enforced.

Examples:
  flint context organization org_abc123def456
  flint context organization abc123def456789
  flint con org org_xyz789  # Using partial matching

Note: You must be authenticated with PocketBase before setting an organization.`,
	Aliases: []string{"org"},
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateConfigManager(); err != nil {
			return err
		}

		organizationID := args[0]
		if organizationID == "" {
			return fmt.Errorf("organization ID cannot be empty")
		}

		// Get the active context
		ctx, err := configManager.GetActiveContext()
		if err != nil {
			return fmt.Errorf("no active context set. Use 'flint context select <n>' to set one")
		}

		// Check if user is authenticated
		if ctx.PocketBase.AuthToken == "" {
			return fmt.Errorf("not authenticated. Run 'flint auth pb' to authenticate first")
		}

		// TODO: In Phase 2, we would validate that the user belongs to this organization
		// by making a PocketBase API call. For now, we'll just update the context.

		// Update the context
		ctx.PocketBase.OrganizationID = organizationID

		// Save the updated context
		if err := configManager.SaveContext(ctx); err != nil {
			return fmt.Errorf("failed to save context: %w", err)
		}

		// Success message
		green := color.New(color.FgGreen).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		
		fmt.Printf("%s Organization set to '%s' for context '%s'\n", 
			green("✓"), cyan(organizationID), cyan(ctx.Name))

		// Show updated context summary
		fmt.Printf("\nContext Summary:\n")
		fmt.Printf("  Context: %s\n", ctx.Name)
		fmt.Printf("  PocketBase URL: %s\n", ctx.PocketBase.URL)
		fmt.Printf("  Organization ID: %s\n", organizationID)
		fmt.Printf("  Auth Collection: %s\n", ctx.PocketBase.AuthCollection)

		// Next steps
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  View collections: %s\n", 
			color.New(color.FgCyan).Sprint("flint collections list organizations"))
		fmt.Printf("  List edges: %s\n", 
			color.New(color.FgCyan).Sprint("flint collections list edges"))
		fmt.Printf("  Subscribe to messages: %s\n", 
			color.New(color.FgCyan).Sprint("flint nats subscribe \"telemetry.>\""))

		// Warning about PocketBase current_organization_id update
		yellow := color.New(color.FgYellow).SprintFunc()
		fmt.Printf("\n%s Note: This updates your local context only.\n", yellow("⚠"))
		fmt.Printf("The PocketBase current_organization_id will be updated when you\n")
		fmt.Printf("make your next authenticated API call.\n")

		return nil
	},
}
