package context

import (
	"fmt"
	"strings"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"flint-cli/internal/config"
)

var (
	pbURL            string
	pbAuthCollection string
	organizationID   string
	natsServers      []string
	natsAuthMethod   string
)

var createCmd = &cobra.Command{
	Use:   "create <name>",
	Short: "Create a new Stone-Age.io context",
	Long: `Create a new context configuration for a Stone-Age.io environment.

A context contains all the connection information needed to work with a specific
Stone-Age.io deployment including PocketBase URL, authentication settings,
NATS server information, and organization details.

The context will be created as a directory containing:
- context.yaml: Main context configuration
- nats.creds: NATS credentials file (when using creds auth method)

Examples:
  flint context create production \\
    --pb-url https://api.stone-age.io \\
    --pb-auth-collection users \\
    --nats-servers nats://nats1.stone-age.io:4222,nats://nats2.stone-age.io:4222 \\
    --nats-auth-method creds

  flint context create development \\
    --pb-url http://localhost:8090 \\
    --nats-servers nats://localhost:4222 \\
    --nats-auth-method user_pass`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateConfigManager(); err != nil {
			return err
		}

		contextName := args[0]
		if contextName == "" {
			return fmt.Errorf("context name cannot be empty")
		}

		// Validate required flags
		if pbURL == "" {
			return fmt.Errorf("--pb-url is required")
		}

		if len(natsServers) == 0 {
			return fmt.Errorf("--nats-servers is required")
		}

		// Validate auth method
		validAuthMethods := []string{config.NATSAuthUserPass, config.NATSAuthToken, config.NATSAuthCreds}
		validMethod := false
		for _, method := range validAuthMethods {
			if natsAuthMethod == method {
				validMethod = true
				break
			}
		}
		if !validMethod {
			return fmt.Errorf("invalid NATS auth method '%s'. Valid options: %s", 
				natsAuthMethod, strings.Join(validAuthMethods, ", "))
		}

		// Validate auth collection
		validCollections := []string{
			config.AuthCollectionUsers,
			config.AuthCollectionClients,
			config.AuthCollectionEdges,
			config.AuthCollectionThings,
			config.AuthCollectionServiceUsers,
		}
		validCollection := false
		for _, collection := range validCollections {
			if pbAuthCollection == collection {
				validCollection = true
				break
			}
		}
		if !validCollection {
			return fmt.Errorf("invalid auth collection '%s'. Valid options: %s", 
				pbAuthCollection, strings.Join(validCollections, ", "))
		}

		// Check if context already exists
		if configManager.ContextExists(contextName) {
			return fmt.Errorf("context '%s' already exists", contextName)
		}

		// Determine NATS credentials file path based on auth method
		var credsFilePath string
		if natsAuthMethod == config.NATSAuthCreds {
			// Use relative path for creds file in the new structure
			credsFilePath = "./nats.creds"
		}

		// Create new context configuration
		newContext := &config.Context{
			Name: contextName,
			PocketBase: config.PocketBaseConfig{
				URL:                  pbURL,
				AuthCollection:       pbAuthCollection,
				OrganizationID:       organizationID,
				AvailableCollections: config.GetDefaultCollections(),
			},
			NATS: config.NATSConfig{
				Servers:    natsServers,
				AuthMethod: natsAuthMethod,
				CredsFile:  credsFilePath, // Relative path to context directory
				TLSEnabled: true,          // Default to secure
				TLSVerify:  true,          // Default to verified
			},
		}

		// Save the context (this will create the directory structure)
		if err := configManager.SaveContext(newContext); err != nil {
			return fmt.Errorf("failed to save context: %w", err)
		}

		// Print success message
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Context '%s' created successfully\n", 
			green("✓"), contextName)

		// Show context directory information
		contextDir := configManager.GetContextDir(contextName)
		fmt.Printf("\nContext directory: %s\n", contextDir)

		// Show configuration summary
		fmt.Printf("\nContext Configuration:\n")
		fmt.Printf("  Name: %s\n", contextName)
		fmt.Printf("  PocketBase URL: %s\n", pbURL)
		fmt.Printf("  Auth Collection: %s\n", pbAuthCollection)
		if organizationID != "" {
			fmt.Printf("  Organization ID: %s\n", organizationID)
		}
		fmt.Printf("  NATS Servers: %s\n", strings.Join(natsServers, ", "))
		fmt.Printf("  NATS Auth Method: %s\n", natsAuthMethod)

		// Show auth method specific information
		switch natsAuthMethod {
		case config.NATSAuthCreds:
			credsPath := configManager.GetContextCredsPath(contextName)
			fmt.Printf("  NATS Creds File: %s\n", credsPath)
			fmt.Printf("\n%s Place your NATS credentials file at: %s\n", 
				color.New(color.FgYellow).Sprint("Note:"), credsPath)
		case config.NATSAuthUserPass:
			fmt.Printf("\n%s Configure NATS username/password using: %s\n",
				color.New(color.FgYellow).Sprint("Note:"),
				color.New(color.FgCyan).Sprint("flint auth nats"))
		case config.NATSAuthToken:
			fmt.Printf("\n%s Configure NATS token using: %s\n",
				color.New(color.FgYellow).Sprint("Note:"),
				color.New(color.FgCyan).Sprint("flint auth nats"))
		}

		// Suggest next steps
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  1. Select this context: %s\n", 
			color.New(color.FgCyan).Sprintf("flint context select %s", contextName))
		fmt.Printf("  2. Authenticate with PocketBase: %s\n", 
			color.New(color.FgCyan).Sprint("flint auth pb"))
		if organizationID == "" {
			fmt.Printf("  3. Set organization: %s\n", 
				color.New(color.FgCyan).Sprint("flint context organization <org_id>"))
		}
		if natsAuthMethod != config.NATSAuthCreds {
			fmt.Printf("  4. Configure NATS authentication: %s\n",
				color.New(color.FgCyan).Sprint("flint auth nats"))
		}

		return nil
	},
}

func init() {
	createCmd.Flags().StringVar(&pbURL, "pb-url", "", "PocketBase server URL (required)")
	createCmd.Flags().StringVar(&pbAuthCollection, "pb-auth-collection", config.AuthCollectionUsers, 
		"PocketBase auth collection (users|clients|edges|things|service_users)")
	createCmd.Flags().StringVar(&organizationID, "organization-id", "", 
		"Organization ID (can be set later)")
	createCmd.Flags().StringSliceVar(&natsServers, "nats-servers", nil, 
		"NATS server URLs (comma-separated, required)")
	createCmd.Flags().StringVar(&natsAuthMethod, "nats-auth-method", config.NATSAuthCreds, 
		"NATS authentication method (user_pass|token|creds)")

	// Mark required flags
	createCmd.MarkFlagRequired("pb-url")
	createCmd.MarkFlagRequired("nats-servers")
}
