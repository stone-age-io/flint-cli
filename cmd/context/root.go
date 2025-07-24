package context

import (
	"fmt"

	"github.com/spf13/cobra"
	"flint-cli/internal/config"
)

// contextCmd represents the context command
var ContextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage Stone-Age.io environment contexts",
	Long: `Context management allows you to work with multiple Stone-Age.io environments.
Each context contains PocketBase configuration, NATS settings, and organization information.

Examples:
  flint context create production --pb-url https://api.stone-age.io
  flint context select production
  flint context organization org_abc123
  flint context list
  flint context show production`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show usage instead of full help when no subcommand provided
		return fmt.Errorf("missing subcommand. See 'flint context --help' for available commands")
	},
}

var configManager *config.Manager

func init() {
	// Add subcommands
	ContextCmd.AddCommand(createCmd)
	ContextCmd.AddCommand(listCmd)
	ContextCmd.AddCommand(selectCmd)
	ContextCmd.AddCommand(showCmd)
	ContextCmd.AddCommand(deleteCmd)
	ContextCmd.AddCommand(organizationCmd)
}

// SetConfigManager sets the configuration manager for the context commands
func SetConfigManager(cm *config.Manager) {
	configManager = cm
}

// validateConfigManager ensures the config manager is available
func validateConfigManager() error {
	if configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return nil
}
