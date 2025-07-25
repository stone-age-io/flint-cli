package auth

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"flint-cli/internal/config"
	natsClient "flint-cli/internal/nats"
	"flint-cli/internal/utils"
)

var (
	natsUsername   string
	natsPassword   string
	natsToken      string
	natsCredsFile  string
	natsAuthMethod string
	natsTestConn   bool
)

var natsCmd = &cobra.Command{
	Use:   "nats",
	Short: "Configure NATS authentication",
	Long: `Configure NATS authentication credentials for the active context.

NATS supports multiple authentication methods:
  user_pass  Username and password authentication
  token      JWT token authentication
  creds      JWT credentials file authentication (recommended)

The authentication method is configured in your context. This command updates
the credentials for the configured method, or allows you to change the method.

Authentication Method Details:

user_pass:
  Simple username and password authentication. Credentials are stored
  in the context configuration file.

token:
  JWT token-based authentication. The token is stored in the context
  configuration file. Tokens may have expiration times.

creds:
  JWT credentials file authentication (recommended for production).
  Uses a .creds file containing JWT and NKey for secure authentication.
  The file path is stored in the context, with support for relative paths.

Examples:
  # Configure username/password authentication
  flint auth nats --method user_pass --username client001 --password secret

  # Configure token authentication
  flint auth nats --method token --token eyJhbGciOiJSUzI1NiIs...

  # Configure credentials file authentication
  flint auth nats --method creds --creds-file /path/to/client.creds

  # Interactive configuration (prompts for credentials)
  flint auth nats

  # Test connection after configuration
  flint auth nats --test`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := validateActiveContext()
		if err != nil {
			return err
		}

		// Determine authentication method
		authMethod := natsAuthMethod
		if authMethod == "" {
			authMethod = ctx.NATS.AuthMethod
		}

		// Validate or prompt for auth method if needed
		if authMethod == "" {
			authMethod, err = promptForAuthMethod()
			if err != nil {
				return fmt.Errorf("failed to get authentication method: %w", err)
			}
		}

		// Validate auth method
		if err := utils.ValidateNATSAuthMethod(authMethod); err != nil {
			return err
		}

		// Configure credentials based on method
		switch authMethod {
		case config.NATSAuthUserPass:
			err = configureUserPassAuth(ctx)
		case config.NATSAuthToken:
			err = configureTokenAuth(ctx)
		case config.NATSAuthCreds:
			err = configureCredsAuth(ctx)
		default:
			return fmt.Errorf("unsupported authentication method: %s", authMethod)
		}

		if err != nil {
			return err
		}

		// Update auth method in context
		ctx.NATS.AuthMethod = authMethod

		// Save updated context
		if err := configManager.SaveContext(ctx); err != nil {
			return fmt.Errorf("failed to save NATS authentication: %w", err)
		}

		// Display success message
		green := color.New(color.FgGreen).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		
		fmt.Printf("\n%s NATS authentication configured successfully!\n", green("✓"))
		fmt.Printf("\nAuthentication Details:\n")
		fmt.Printf("  Method: %s\n", authMethod)
		fmt.Printf("  Context: %s\n", cyan(ctx.Name))
		
		switch authMethod {
		case config.NATSAuthUserPass:
			fmt.Printf("  Username: %s\n", ctx.NATS.Username)
		case config.NATSAuthToken:
			fmt.Printf("  Token: %s\n", truncateToken(ctx.NATS.Token))
		case config.NATSAuthCreds:
			fmt.Printf("  Credentials File: %s\n", ctx.NATS.CredsFile)
		}

		// Test connection if requested
		if natsTestConn {
			fmt.Printf("\nTesting NATS connection...\n")
			if err := testNATSConnection(ctx); err != nil {
				utils.PrintWarning(fmt.Sprintf("Connection test failed: %v", err))
				fmt.Printf("\nThe credentials were saved, but connection testing failed.\n")
				fmt.Printf("Please verify your NATS server configuration and credentials.\n")
			} else {
				fmt.Printf("%s Connection test successful!\n", green("✓"))
			}
		}

		// Show next steps
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  Test connection: %s\n", cyan("flint auth nats --test"))
		fmt.Printf("  Publish message: %s\n", cyan("flint nats publish test.subject \"hello world\""))
		fmt.Printf("  Subscribe to messages: %s\n", cyan("flint nats subscribe \"test.>\""))

		return nil
	},
}

func init() {
	natsCmd.Flags().StringVar(&natsAuthMethod, "method", "", 
		"Authentication method (user_pass|token|creds)")
	natsCmd.Flags().StringVarP(&natsUsername, "username", "u", "", 
		"Username for user_pass authentication")
	natsCmd.Flags().StringVarP(&natsPassword, "password", "p", "", 
		"Password for user_pass authentication (will prompt if not provided)")
	natsCmd.Flags().StringVar(&natsToken, "token", "", 
		"JWT token for token authentication")
	natsCmd.Flags().StringVar(&natsCredsFile, "creds-file", "", 
		"Path to JWT credentials file for creds authentication")
	natsCmd.Flags().BoolVar(&natsTestConn, "test", false, 
		"Test NATS connection after configuration")
}

// promptForAuthMethod prompts the user to select an authentication method
func promptForAuthMethod() (string, error) {
	fmt.Printf("Select NATS authentication method:\n")
	fmt.Printf("  1. user_pass - Username and password\n")
	fmt.Printf("  2. token - JWT token\n")
	fmt.Printf("  3. creds - JWT credentials file (recommended)\n")
	fmt.Printf("Enter choice (1-3): ")

	reader := bufio.NewReader(os.Stdin)
	choice, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	choice = strings.TrimSpace(choice)
	switch choice {
	case "1":
		return config.NATSAuthUserPass, nil
	case "2":
		return config.NATSAuthToken, nil
	case "3":
		return config.NATSAuthCreds, nil
	default:
		return "", fmt.Errorf("invalid choice: %s", choice)
	}
}

// configureUserPassAuth configures username/password authentication
func configureUserPassAuth(ctx *config.Context) error {
	username := natsUsername
	password := natsPassword

	// Get username if not provided
	if username == "" {
		if ctx.NATS.Username != "" {
			fmt.Printf("Current username: %s\n", ctx.NATS.Username)
			fmt.Print("New username (press Enter to keep current): ")
		} else {
			fmt.Print("Username: ")
		}

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read username: %w", err)
		}

		input = strings.TrimSpace(input)
		if input != "" {
			username = input
		} else if ctx.NATS.Username != "" {
			username = ctx.NATS.Username
		} else {
			return fmt.Errorf("username is required")
		}
	}

	// Get password if not provided
	if password == "" {
		fmt.Print("Password: ")
		passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
		if err != nil {
			return fmt.Errorf("failed to read password: %w", err)
		}
		fmt.Println() // New line after hidden input
		password = string(passwordBytes)
	}

	if username == "" || password == "" {
		return fmt.Errorf("both username and password are required")
	}

	// Update context
	ctx.NATS.Username = username
	ctx.NATS.Password = password
	// Clear other auth fields
	ctx.NATS.Token = ""
	ctx.NATS.CredsFile = ""

	return nil
}

// configureTokenAuth configures JWT token authentication
func configureTokenAuth(ctx *config.Context) error {
	token := natsToken

	// Get token if not provided
	if token == "" {
		if ctx.NATS.Token != "" {
			fmt.Printf("Current token: %s\n", truncateToken(ctx.NATS.Token))
			fmt.Print("New token (press Enter to keep current): ")
		} else {
			fmt.Print("JWT Token: ")
		}

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read token: %w", err)
		}

		input = strings.TrimSpace(input)
		if input != "" {
			token = input
		} else if ctx.NATS.Token != "" {
			token = ctx.NATS.Token
		} else {
			return fmt.Errorf("token is required")
		}
	}

	if token == "" {
		return fmt.Errorf("JWT token is required")
	}

	// Basic token validation
	if !strings.HasPrefix(token, "eyJ") {
		utils.PrintWarning("Token does not appear to be a valid JWT (should start with 'eyJ')")
	}

	// Update context
	ctx.NATS.Token = token
	// Clear other auth fields
	ctx.NATS.Username = ""
	ctx.NATS.Password = ""
	ctx.NATS.CredsFile = ""

	return nil
}

// configureCredsAuth configures JWT credentials file authentication
func configureCredsAuth(ctx *config.Context) error {
	credsFile := natsCredsFile

	// Get credentials file path if not provided
	if credsFile == "" {
		if ctx.NATS.CredsFile != "" {
			displayPath := ctx.NATS.CredsFile
			if strings.HasPrefix(displayPath, "./") {
				displayPath = configManager.GetContextCredsPath(ctx.Name)
			}
			fmt.Printf("Current credentials file: %s\n", displayPath)
			fmt.Print("New credentials file path (press Enter to keep current): ")
		} else {
			fmt.Print("Credentials file path: ")
		}

		reader := bufio.NewReader(os.Stdin)
		input, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read credentials file path: %w", err)
		}

		input = strings.TrimSpace(input)
		if input != "" {
			credsFile = input
		} else if ctx.NATS.CredsFile != "" {
			credsFile = ctx.NATS.CredsFile
		} else {
			return fmt.Errorf("credentials file path is required")
		}
	}

	if credsFile == "" {
		return fmt.Errorf("credentials file path is required")
	}

	// Determine if we should use relative path
	var finalCredsFile string
	contextDir := configManager.GetContextDir(ctx.Name)
	
	// Check if file is within context directory
	if strings.HasPrefix(credsFile, contextDir) {
		// Convert to relative path
		relPath, err := filepath.Rel(contextDir, credsFile)
		if err == nil && !strings.HasPrefix(relPath, "..") {
			finalCredsFile = "./" + relPath
			utils.PrintInfo(fmt.Sprintf("Using relative path: %s", finalCredsFile))
		} else {
			finalCredsFile = credsFile
		}
	} else {
		finalCredsFile = credsFile
	}

	// Validate file exists (if absolute path or relative from current dir)
	testPath := finalCredsFile
	if strings.HasPrefix(finalCredsFile, "./") {
		testPath = filepath.Join(contextDir, finalCredsFile[2:])
	}

	if _, err := os.Stat(testPath); os.IsNotExist(err) {
		utils.PrintWarning(fmt.Sprintf("Credentials file does not exist: %s", testPath))
		fmt.Print("Continue anyway? (y/N): ")
		
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("failed to read confirmation: %w", err)
		}
		
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			return fmt.Errorf("credentials file configuration cancelled")
		}
	}

	// Update context
	ctx.NATS.CredsFile = finalCredsFile
	// Clear other auth fields
	ctx.NATS.Username = ""
	ctx.NATS.Password = ""
	ctx.NATS.Token = ""

	return nil
}

// testNATSConnection tests the NATS connection with the configured credentials
func testNATSConnection(ctx *config.Context) error {
	client := natsClient.NewClientFromContext(ctx)
	
	// Resolve credentials file path if relative
	if ctx.NATS.CredsFile != "" && strings.HasPrefix(ctx.NATS.CredsFile, "./") {
		contextDir := configManager.GetContextDir(ctx.Name)
		actualCredsFile := filepath.Join(contextDir, ctx.NATS.CredsFile[2:])
		
		// Temporarily update the client's creds file for testing
		tempConfig := ctx.NATS
		tempConfig.CredsFile = actualCredsFile
		client = natsClient.NewClient(&tempConfig)
	}
	
	// Test connection
	if err := client.Connect(); err != nil {
		return fmt.Errorf("failed to connect: %w", err)
	}
	defer func() {
		if disconnectErr := client.Disconnect(); disconnectErr != nil {
			utils.PrintDebug(fmt.Sprintf("Error disconnecting test connection: %v", disconnectErr))
		}
	}()

	// Get connection status
	status := client.GetConnectionStatus()
	if !status.Connected {
		return fmt.Errorf("connection established but not healthy")
	}

	utils.PrintInfo(fmt.Sprintf("Connected to: %s", status.URL))
	utils.PrintInfo(fmt.Sprintf("Server: %s", status.ServerName))

	return nil
}

// truncateToken truncates a JWT token for display
func truncateToken(token string) string {
	if len(token) <= 50 {
		return token
	}
	return token[:20] + "..." + token[len(token)-20:]
}
