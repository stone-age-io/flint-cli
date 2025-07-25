package nats

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"flint-cli/internal/config"
	natsClient "flint-cli/internal/nats"
)

// NATSCmd represents the nats command group
var NATSCmd = &cobra.Command{
	Use:   "nats",
	Short: "NATS messaging operations for Stone-Age.io",
	Long: `Interact with the Stone-Age.io NATS messaging system for real-time communication.

NATS provides the messaging backbone for Stone-Age.io IoT platform, enabling
real-time communication between edges, things, and the control plane.

The NATS commands support multiple authentication methods:
  user_pass  Username and password authentication
  creds      JWT credentials file authentication  
  token      Token-based authentication

Authentication is configured per context and uses the settings from your
active context configuration.

Stone-Age.io NATS Subject Examples:
  telemetry.>          All telemetry data
  telemetry.edge.*     Edge device telemetry
  telemetry.thing.*    Individual thing telemetry
  command.>            All command messages
  command.edge.*       Edge device commands
  command.thing.*      Thing-specific commands
  system.>             System messages and alerts
  events.>             Platform events and notifications

Examples:
  # Publish telemetry data from an edge
  flint nats publish "telemetry.edge.edge_123" '{"temperature": 22.5, "timestamp": "2025-01-15T10:30:00Z"}'
  
  # Subscribe to all telemetry messages
  flint nats subscribe "telemetry.>"
  
  # Subscribe to edge commands with queue group
  flint nats subscribe "command.edge.*" --queue edge_processors
  
  # Publish with headers
  flint nats publish "system.alerts.critical" '{"alert": "High temperature"}' \\
    --header source=admin_console --header priority=high

Authentication will be performed automatically using your context configuration.
Ensure you have authenticated with PocketBase and configured NATS settings in your context.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show usage when no subcommand provided
		return fmt.Errorf("missing subcommand. See 'flint nats --help' for available commands")
	},
}

var configManager *config.Manager

func init() {
	// Add subcommands
	NATSCmd.AddCommand(publishCmd)
	NATSCmd.AddCommand(subscribeCmd)
	// Note: request command will be added in a future enhancement
}

// SetConfigManager sets the configuration manager for the NATS commands
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

// validateActiveContext ensures there's an active context with NATS configuration
func validateActiveContext() (*config.Context, error) {
	if err := validateConfigManager(); err != nil {
		return nil, err
	}

	ctx, err := configManager.GetActiveContext()
	if err != nil {
		return nil, fmt.Errorf("no active context set. Use 'flint context select <name>' to set one")
	}

	// Validate NATS configuration
	if err := validateNATSConfig(&ctx.NATS); err != nil {
		return nil, fmt.Errorf("invalid NATS configuration: %w", err)
	}

	return ctx, nil
}

// validateNATSConfig validates the NATS configuration in a context
func validateNATSConfig(natsConfig *config.NATSConfig) error {
	if len(natsConfig.Servers) == 0 {
		return fmt.Errorf("no NATS servers configured. Add servers to your context with 'flint context create' or update your context configuration")
	}

	// Validate auth method
	validMethods := []string{config.NATSAuthUserPass, config.NATSAuthToken, config.NATSAuthCreds}
	validMethod := false
	for _, method := range validMethods {
		if natsConfig.AuthMethod == method {
			validMethod = true
			break
		}
	}
	
	if !validMethod {
		return fmt.Errorf("invalid NATS auth method '%s'. Valid methods: user_pass, token, creds", natsConfig.AuthMethod)
	}

	// Validate auth-specific configuration
	switch natsConfig.AuthMethod {
	case config.NATSAuthUserPass:
		if natsConfig.Username == "" || natsConfig.Password == "" {
			return fmt.Errorf("username and password are required for user_pass authentication. Configure with 'flint auth nats'")
		}
	case config.NATSAuthToken:
		if natsConfig.Token == "" {
			return fmt.Errorf("token is required for token authentication. Configure with 'flint auth nats'")
		}
	case config.NATSAuthCreds:
		if natsConfig.CredsFile == "" {
			return fmt.Errorf("credentials file is required for creds authentication. Configure with 'flint auth nats' or place file in context directory")
		}
	}

	return nil
}

// createNATSClient creates a NATS client from the active context
func createNATSClient() (*natsClient.Client, error) {
	ctx, err := validateActiveContext()
	if err != nil {
		return nil, err
	}

	return createNATSClientFromContext(ctx), nil
}

// createNATSClientFromContext creates a NATS client from a context
func createNATSClientFromContext(ctx *config.Context) *natsClient.Client {
	// We need to resolve the relative creds file path if present
	natsConfig := ctx.NATS
	
	// Convert relative creds file path to absolute if needed
	if natsConfig.CredsFile != "" && strings.HasPrefix(natsConfig.CredsFile, "./") {
		contextDir := configManager.GetContextDir(ctx.Name)
		// Replace "./" with the context directory path
		natsConfig.CredsFile = strings.Replace(natsConfig.CredsFile, "./", contextDir+"/", 1)
	}
	
	return natsClient.NewClient(&natsConfig)
}

// Common flag variables that will be used across NATS commands
var (
	outputFormat string
	timeout      string
	verbose      bool
)

func init() {
	// Add persistent flags that apply to all NATS commands
	NATSCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "text", "Output format (text|json|yaml)")
	NATSCmd.PersistentFlags().StringVar(&timeout, "timeout", "30s", "Operation timeout")
	NATSCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose output")
}
