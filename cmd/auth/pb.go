package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"golang.org/x/term"
	"flint-cli/internal/config"
	"flint-cli/internal/pocketbase"
	"flint-cli/internal/utils"
)

var (
	pbEmail      string
	pbPassword   string
	pbCollection string
	pbOrgID      string
)

var pbCmd = &cobra.Command{
	Use:   "pb",
	Short: "Authenticate with PocketBase",
	Long: `Authenticate with the Stone-Age.io PocketBase instance.

PocketBase handles all database operations and API access for Stone-Age.io.
You can authenticate using different collections depending on your role:

Collections:
  users        Human administrators (default)
  clients      NATS client entities
  edges        Edge device authentication  
  things       Individual IoT device authentication
  service_users System service accounts

The authentication will:
1. Validate your credentials with PocketBase
2. Store the session token securely in your context
3. Validate organization membership (for user accounts)
4. Update your current organization if specified

Examples:
  # Interactive authentication (prompts for credentials)
  flint auth pb

  # Authenticate with specific credentials
  flint auth pb --email admin@company.com --password secret

  # Authenticate as an edge device
  flint auth pb --collection edges --email edge001@company.com

  # Authenticate and set organization
  flint auth pb --email admin@company.com --organization org_abc123def456`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, err := validateActiveContext()
		if err != nil {
			return err
		}

		// Use collection from context if not specified
		if pbCollection == "" {
			pbCollection = ctx.PocketBase.AuthCollection
		}

		// Validate collection
		if err := utils.ValidateAuthCollection(pbCollection); err != nil {
			return err
		}

		// Get credentials if not provided
		if pbEmail == "" {
			pbEmail, err = promptForEmail()
			if err != nil {
				return fmt.Errorf("failed to get email: %w", err)
			}
		}

		if pbPassword == "" {
			pbPassword, err = promptForPassword()
			if err != nil {
				return fmt.Errorf("failed to get password: %w", err)
			}
		}

		// Validate email format
		if err := utils.ValidateEmail(pbEmail); err != nil {
			return fmt.Errorf("invalid email format: %w", err)
		}

		// Create PocketBase client
		client := pocketbase.NewClient(ctx.PocketBase.URL)

		// Test connection first
		utils.PrintInfo("Testing connection to PocketBase...")
		if err := client.GetHealth(); err != nil {
			return fmt.Errorf("failed to connect to PocketBase at %s: %w", ctx.PocketBase.URL, err)
		}

		// Perform authentication
		utils.PrintInfo(fmt.Sprintf("Authenticating with collection '%s'...", pbCollection))
		
		authResp, err := client.Authenticate(pbCollection, pbEmail, pbPassword)
		if err != nil {
			if pbErr, ok := err.(*pocketbase.PocketBaseError); ok {
				utils.PrintError(fmt.Errorf("%s", pbErr.GetFriendlyMessage()))
				if suggestion := pbErr.GetSuggestion(); suggestion != "" {
					fmt.Printf("\nSuggestion: %s\n", suggestion)
				}
				return fmt.Errorf("authentication failed")
			}
			return fmt.Errorf("authentication failed: %w", err)
		}

		// Handle organization validation for user accounts
		if pbCollection == config.AuthCollectionUsers {
			orgID, err := handleUserOrganization(client, authResp)
			if err != nil {
				return err
			}
			if orgID != "" {
				pbOrgID = orgID
			}
		}

		// Update context with authentication data
		if err := pocketbase.UpdateAuthContextFromResponse(ctx, authResp, pbOrgID); err != nil {
			return fmt.Errorf("failed to update context: %w", err)
		}

		// Save updated context
		if err := configManager.SaveContext(ctx); err != nil {
			return fmt.Errorf("failed to save authentication: %w", err)
		}

		// Update PocketBase current_organization_id if we're a user and have an org
		if pbCollection == config.AuthCollectionUsers && pbOrgID != "" {
			utils.PrintInfo("Updating current organization in PocketBase...")
			if err := client.UpdateCurrentOrganization(pbOrgID); err != nil {
				utils.PrintWarning(fmt.Sprintf("Failed to update current organization in PocketBase: %v", err))
				// Don't fail the authentication for this
			}
		}

		// Display success message
		green := color.New(color.FgGreen).SprintFunc()
		cyan := color.New(color.FgCyan).SprintFunc()
		
		fmt.Printf("\n%s Authentication successful!\n", green("✓"))
		
		// Show authentication details
		fmt.Printf("\nAuthentication Details:\n")
		fmt.Printf("  Collection: %s\n", pocketbase.GetCollectionDisplayName(pbCollection))
		fmt.Printf("  Identity: %s\n", pbEmail)
		fmt.Printf("  Context: %s\n", cyan(ctx.Name))
		
		if authResp.Record != nil {
			if name := getRecordDisplayName(authResp.Record, pbCollection); name != "" {
				fmt.Printf("  Name: %s\n", name)
			}
		}

		if pbOrgID != "" {
			fmt.Printf("  Organization: %s\n", cyan(pbOrgID))
		}

		// Show available next steps
		fmt.Printf("\nNext steps:\n")
		fmt.Printf("  View your profile: %s\n", 
			color.New(color.FgCyan).Sprint("flint collections get users $(flint context show --output json | jq -r '.pocketbase.auth_record.id')"))
		
		if pbOrgID != "" {
			fmt.Printf("  List organization resources: %s\n", 
				color.New(color.FgCyan).Sprint("flint collections list edges"))
			fmt.Printf("  View organization details: %s\n", 
				color.New(color.FgCyan).Sprintf("flint collections get organizations %s", pbOrgID))
		}

		if pbCollection == config.AuthCollectionUsers && pbOrgID == "" {
			fmt.Printf("  Set organization: %s\n", 
				color.New(color.FgCyan).Sprint("flint context organization <org_id>"))
		}

		return nil
	},
}

func init() {
	pbCmd.Flags().StringVarP(&pbEmail, "email", "e", "", "Email address for authentication")
	pbCmd.Flags().StringVarP(&pbPassword, "password", "p", "", "Password for authentication (will prompt if not provided)")
	pbCmd.Flags().StringVarP(&pbCollection, "collection", "c", "", "Authentication collection (users|clients|edges|things|service_users)")
	pbCmd.Flags().StringVarP(&pbOrgID, "organization", "o", "", "Organization ID to set after authentication")
}

// promptForEmail prompts the user for their email address
func promptForEmail() (string, error) {
	fmt.Print("Email: ")
	reader := bufio.NewReader(os.Stdin)
	email, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(email), nil
}

// promptForPassword prompts the user for their password (hidden input)
func promptForPassword() (string, error) {
	fmt.Print("Password: ")
	
	// Hide password input
	passwordBytes, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", err
	}
	
	fmt.Println() // New line after hidden input
	return string(passwordBytes), nil
}

// handleUserOrganization handles organization selection and validation for user accounts
func handleUserOrganization(client *pocketbase.Client, authResp *pocketbase.AuthResponse) (string, error) {
	// If organization was specified via flag, validate it
	if pbOrgID != "" {
		if err := client.ValidateOrganizationAccess(pbOrgID); err != nil {
			return "", fmt.Errorf("organization validation failed: %w", err)
		}
		utils.PrintInfo(fmt.Sprintf("Organization '%s' validated successfully", pbOrgID))
		return pbOrgID, nil
	}

	// Get user's organizations
	orgs, err := client.GetUserOrganizations()
	if err != nil {
		utils.PrintWarning(fmt.Sprintf("Could not retrieve user organizations: %v", err))
		return "", nil // Don't fail authentication for this
	}

	if len(orgs) == 0 {
		utils.PrintWarning("User is not a member of any organizations")
		fmt.Printf("\nYou can set an organization later using:\n")
		fmt.Printf("  %s\n", color.New(color.FgCyan).Sprint("flint context organization <org_id>"))
		return "", nil
	}

	// If user belongs to only one organization, use it automatically
	if len(orgs) == 1 {
		orgID := orgs[0]["id"].(string)
		orgName := ""
		if name, ok := orgs[0]["name"].(string); ok {
			orgName = name
		}
		
		utils.PrintInfo(fmt.Sprintf("Automatically selected organization: %s (%s)", orgName, orgID))
		return orgID, nil
	}

	// Multiple organizations - let user choose or use existing context setting
	ctx, _ := configManager.GetActiveContext()
	if ctx != nil && ctx.PocketBase.OrganizationID != "" {
		// Validate existing organization setting
		if err := client.ValidateOrganizationAccess(ctx.PocketBase.OrganizationID); err == nil {
			utils.PrintInfo(fmt.Sprintf("Using existing organization from context: %s", ctx.PocketBase.OrganizationID))
			return ctx.PocketBase.OrganizationID, nil
		}
	}

	// Prompt user to select organization
	fmt.Printf("\nYou belong to multiple organizations:\n")
	for i, org := range orgs {
		orgID := org["id"].(string)
		orgName := ""
		if name, ok := org["name"].(string); ok {
			orgName = name
		}
		fmt.Printf("  %d. %s (%s)\n", i+1, orgName, orgID)
	}

	fmt.Printf("\nSelect organization (1-%d), or press Enter to set later: ", len(orgs))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read organization selection: %w", err)
	}

	input = strings.TrimSpace(input)
	if input == "" {
		fmt.Printf("You can set an organization later using:\n")
		fmt.Printf("  %s\n", color.New(color.FgCyan).Sprint("flint context organization <org_id>"))
		return "", nil
	}

	// Parse selection
	var selection int
	if _, err := fmt.Sscanf(input, "%d", &selection); err != nil || selection < 1 || selection > len(orgs) {
		return "", fmt.Errorf("invalid organization selection")
	}

	selectedOrg := orgs[selection-1]
	orgID := selectedOrg["id"].(string)
	
	return orgID, nil
}

// getRecordDisplayName returns a human-readable display name for a record
func getRecordDisplayName(record map[string]interface{}, collection string) string {
	switch collection {
	case config.AuthCollectionUsers:
		userRecord := pocketbase.UserRecord{Record: record}
		return userRecord.GetFullName()
	case config.AuthCollectionClients:
		if username, ok := record["nats_username"].(string); ok {
			return username
		}
	case config.AuthCollectionEdges, config.AuthCollectionThings:
		if name, ok := record["name"].(string); ok {
			return name
		}
	case config.AuthCollectionServiceUsers:
		if username, ok := record["username"].(string); ok {
			return username
		}
	}
	
	// Fallback to email or ID
	if email, ok := record["email"].(string); ok && email != "" {
		return email
	}
	if id, ok := record["id"].(string); ok {
		return id
	}
	
	return ""
}
