package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"flint-cli/cmd/auth"
	"flint-cli/cmd/collections"
	"flint-cli/cmd/context"
	"flint-cli/cmd/nats"
	"flint-cli/internal/config"
	"flint-cli/internal/resolver"
)

var (
	configManager *config.Manager
	cmdResolver   *resolver.CommandResolver
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flint",
	Short: "CLI tool for managing Stone-Age.io IoT platform",
	Long: `Flint is a command-line interface for managing your Stone-Age.io IoT platform.
It provides comprehensive tools for managing contexts, authenticating with PocketBase,
performing CRUD operations on collections, and interacting with NATS messaging.

Features:
- Multi-environment context management with organization support
- PocketBase authentication and collection operations
- NATS messaging with multiple authentication methods
- Partial command matching (Cisco-style)
- Generic file upload/download operations`,
	Version: "0.1.0",
	RunE: func(cmd *cobra.Command, args []string) error {
		// Show usage instead of full help when no subcommand provided
		return fmt.Errorf("missing subcommand. See 'flint --help' for available commands")
	},
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Initialize configuration manager
		var err error
		configManager, err = config.NewManager()
		if err != nil {
			return fmt.Errorf("failed to initialize configuration: %w", err)
		}

		// Initialize command resolver for partial matching
		cmdResolver = resolver.NewCommandResolver()

		// Pass config manager and resolver to command groups
		context.SetConfigManager(configManager)
		auth.SetConfigManager(configManager)
		collections.SetConfigManager(configManager)
		collections.SetCommandResolver(cmdResolver)
		nats.SetConfigManager(configManager)

		return nil
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	// Global flags
	rootCmd.PersistentFlags().StringVar(&config.Global.OutputFormat, "output", "json", "Output format (json|yaml|table)")
	rootCmd.PersistentFlags().BoolVar(&config.Global.ColorsEnabled, "colors", true, "Enable colored output")
	rootCmd.PersistentFlags().BoolVar(&config.Global.Debug, "debug", false, "Enable debug output")

	// Bind flags to viper
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("colors", rootCmd.PersistentFlags().Lookup("colors"))
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))

	// Add command groups
	addCommands()
}

// addCommands adds all command groups to the root command
func addCommands() {
	// Context management commands
	rootCmd.AddCommand(context.ContextCmd)
	
	// Authentication commands
	rootCmd.AddCommand(auth.AuthCmd)
	
	// Collections CRUD commands
	rootCmd.AddCommand(collections.CollectionsCmd)
	
	// NATS messaging commands
	rootCmd.AddCommand(nats.NATSCmd)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	// Set config file type
	viper.SetConfigType("yaml")
	viper.SetConfigName("config")
	
	// Look for config in XDG config directory
	if configManager != nil {
		viper.AddConfigPath(configManager.GetConfigDir())
	}

	// Read in environment variables that match
	viper.SetEnvPrefix("FLINT")
	viper.AutomaticEnv()

	// If a config file is found, read it in
	if err := viper.ReadInConfig(); err == nil && config.Global.Debug {
		fmt.Fprintf(os.Stderr, "Using config file: %s\n", viper.ConfigFileUsed())
	}
}

// GetConfigManager returns the global configuration manager instance
func GetConfigManager() *config.Manager {
	return configManager
}

// GetCommandResolver returns the global command resolver instance
func GetCommandResolver() *resolver.CommandResolver {
	return cmdResolver
}
