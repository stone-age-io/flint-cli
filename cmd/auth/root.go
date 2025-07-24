package auth

import (
	"fmt"

	"github.com/spf13/cobra"
	"flint-cli/internal/config"
)

// AuthCmd represents the auth command
var AuthCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Stone-Age.io services",
	Long: `Authentication commands for Stone-Age.io platform services.

The auth command group provides authentication for both PocketBase (database/API)
and NATS (messaging) services. Each service has its own authentication methods
and requirements.

Stone-Age.io supports multiple authentication collections in PocketBase:
- users (default): Human administrators with organization membership
- clients: NATS client authentication entities  
- edges: Edge device authentication
- things: Individual IoT device authentication
- service_users: System service accounts

Examples:
  # Authenticate with PocketBase as a user
  flint auth pb --email admin@company.com --password secret

  # Authenticate with PocketBase as an edge device
  flint auth pb --collection edges --email edge001@company.com --password secret
  
  # Authenticate with NATS (configured per context)
  flint auth nats --username client001 --password secret`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show usage instead of full help when no subcommand provided
		return fmt.Errorf("missing subcommand. See 'flint auth --help' for available commands")
	},
}

var configManager *config.Manager

func init() {
	// Add subcommands
	AuthCmd.AddCommand(pbCmd)
	// NATS auth command will be added in Phase 4
}

// SetConfigManager sets the configuration manager for the auth commands
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

// validateActiveContext ensures there's an active context
func validateActiveContext() (*config.Context, error) {
	if err := validateConfigManager(); err != nil {
		return nil, err
	}

	ctx, err := configManager.GetActiveContext()
	if err != nil {
		return nil, fmt.Errorf("no active context set. Use 'flint context select <name>' to set one")
	}

	return ctx, nil
}
