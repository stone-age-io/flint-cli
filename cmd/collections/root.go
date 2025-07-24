package collections

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"flint-cli/internal/config"
	"flint-cli/internal/pocketbase"
	"flint-cli/internal/resolver"
)

// Global flag variables
var (
	// List flags
	offsetFlag int
	limitFlag  int
	filterFlag string
	sortFlag   string
	fieldsFlag []string
	expandFlag []string
	
	// Create/Update flags
	fileFlag string
	
	// Delete flags
	forceFlag bool
	quietFlag bool
	
	// Common flags
	outputFlag string
)

// CollectionsCmd represents the collections command
var CollectionsCmd = &cobra.Command{
	Use:   "collections <collection> <action> [args]",
	Short: "Manage Stone-Age.io collections",
	Long: `Perform CRUD operations on Stone-Age.io collections through PocketBase.

Collections are the data entities in your Stone-Age.io platform including organizations,
users, edges, things, locations, and more. This command provides full CRUD (Create,
Read, Update, Delete) operations for all available collections.

The available collections depend on your current context configuration. Each context
defines which collections are accessible based on your Stone-Age.io deployment.

Usage Pattern:
  flint collections <collection> <action> [args] [flags]

Examples:
  # List all edges in your organization
  flint collections edges list

  # List with filtering and custom fields
  flint collections edges list --filter 'active=true && region="us-west"' --fields name,code,region

  # Get a specific user by ID with expanded relations
  flint collections users get user_abc123def456 --expand organizations

  # Create a new organization from JSON
  flint collections organizations create '{"name":"My Organization","code":"myorg"}'

  # Create from file
  flint collections edges create --file edge-config.json

  # Update an edge device
  flint collections edges update edge_123 '{"name":"Updated Edge Name"}' --output table

  # Delete a thing with confirmation skip
  flint collections things delete thing_456 --force

Available Actions:
  list     List records from a collection with filtering and pagination
  get      Get a single record by ID with optional expansion
  create   Create a new record from JSON data or file
  update   Update an existing record with JSON data or file
  delete   Delete a record with confirmation

Available Collections (depends on your context):
  organizations    Multi-tenant organization management
  users           Human administrators with organization membership
  edges           Edge computing nodes managing local IoT devices  
  things          IoT devices (door controllers, sensors, etc.)
  locations       Physical locations with hierarchical structure
  clients         NATS client authentication entities
  edge_types      Edge device type definitions
  thing_types     IoT device type definitions
  location_types  Location type definitions
  edge_regions    Geographic/logical edge groupings
  audit_logs      System audit trail (read-only)
  topic_permissions NATS topic access control

Note: All operations are automatically scoped to your current organization
via PocketBase rules. You must be authenticated and have the appropriate
permissions for the target collection.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Need at least collection and action
		if len(args) < 2 {
			return fmt.Errorf("missing required arguments: <collection> <action>")
		}

		collection := args[0]
		action := args[1]
		actionArgs := args[2:] // Remaining args for the action

		// Validate collection against context
		ctx, err := validateCollection(collection)
		if err != nil {
			return err
		}

		// Resolve partial action matching
		resolvedAction, err := resolveAction(action)
		if err != nil {
			return err
		}

		// Route to appropriate action handler
		return routeToAction(ctx, collection, resolvedAction, actionArgs)
	},
}

var (
	configManager *config.Manager
	cmdResolver   *resolver.CommandResolver
)

func init() {
	// Register all possible flags for collections commands
	// List-specific flags
	CollectionsCmd.Flags().IntVar(&offsetFlag, "offset", 0, "Number of records to skip (for pagination)")
	CollectionsCmd.Flags().IntVar(&limitFlag, "limit", 30, "Maximum number of records to return")
	CollectionsCmd.Flags().StringVar(&filterFlag, "filter", "", "PocketBase filter expression (e.g., 'active=true && name~\"test\"')")
	CollectionsCmd.Flags().StringVar(&sortFlag, "sort", "", "Sort expression (e.g., 'name', '-created', 'name,-updated')")
	CollectionsCmd.Flags().StringSliceVar(&fieldsFlag, "fields", nil, "Specific fields to return (comma-separated)")
	CollectionsCmd.Flags().StringSliceVar(&expandFlag, "expand", nil, "Relations to expand (comma-separated)")
	
	// Create/Update flags
	CollectionsCmd.Flags().StringVar(&fileFlag, "file", "", "Path to JSON file containing record data")
	
	// Delete flags
	CollectionsCmd.Flags().BoolVarP(&forceFlag, "force", "f", false, "Skip confirmation prompt")
	CollectionsCmd.Flags().BoolVarP(&quietFlag, "quiet", "q", false, "Suppress success messages")
	
	// Common flags
	CollectionsCmd.Flags().StringVarP(&outputFlag, "output", "o", "", "Output format (json|yaml|table)")
}

// SetConfigManager sets the configuration manager for the collections commands
func SetConfigManager(cm *config.Manager) {
	configManager = cm
}

// SetCommandResolver sets the command resolver for partial matching
func SetCommandResolver(cr *resolver.CommandResolver) {
	cmdResolver = cr
}

// validateConfigManager ensures the config manager is available
func validateConfigManager() error {
	if configManager == nil {
		return fmt.Errorf("configuration manager not initialized")
	}
	return nil
}

// validateCommandResolver ensures the command resolver is available
func validateCommandResolver() error {
	if cmdResolver == nil {
		return fmt.Errorf("command resolver not initialized")
	}
	return nil
}

// validateActiveContext ensures there's an active context with authentication
func validateActiveContext() (*config.Context, error) {
	if err := validateConfigManager(); err != nil {
		return nil, err
	}

	ctx, err := configManager.GetActiveContext()
	if err != nil {
		return nil, fmt.Errorf("no active context set. Use 'flint context select <name>' to set one")
	}

	// Check authentication
	if ctx.PocketBase.AuthToken == "" {
		return nil, fmt.Errorf("authentication required. Run 'flint auth pb' to authenticate")
	}

	// Check if authentication is still valid
	if !pocketbase.IsAuthValid(ctx) {
		return nil, fmt.Errorf("authentication has expired. Run 'flint auth pb' to re-authenticate")
	}

	return ctx, nil
}

// validateCollection validates that the collection is available in the current context
func validateCollection(collection string) (*config.Context, error) {
	ctx, err := validateActiveContext()
	if err != nil {
		return nil, err
	}

	if err := validateCommandResolver(); err != nil {
		return nil, err
	}

	// Validate collection against context's available collections
	if err := cmdResolver.ValidateCollection(collection, ctx.PocketBase.AvailableCollections); err != nil {
		return nil, err
	}

	return ctx, nil
}

// resolveAction resolves a partial action command to its full form
func resolveAction(partialAction string) (string, error) {
	if err := validateCommandResolver(); err != nil {
		return "", err
	}

	return cmdResolver.ResolveCommand("collections", partialAction)
}

// createPocketBaseClient creates an authenticated PocketBase client from context
func createPocketBaseClient(ctx *config.Context) *pocketbase.Client {
	return pocketbase.NewClientFromContext(ctx)
}

// routeToAction routes the command to the appropriate action handler
func routeToAction(ctx *config.Context, collection, action string, args []string) error {
	switch action {
	case "list":
		return handleListAction(ctx, collection, args)
	case "get":
		return handleGetAction(ctx, collection, args)
	case "create":
		return handleCreateAction(ctx, collection, args)
	case "update":
		return handleUpdateAction(ctx, collection, args)
	case "delete":
		return handleDeleteAction(ctx, collection, args)
	default:
		return fmt.Errorf("unknown action '%s'. Available actions: list, get, create, update, delete", action)
	}
}

// parseJSONInput parses JSON input from string or file
func parseJSONInput(jsonStr, filePath string) (map[string]interface{}, error) {
	if filePath != "" && jsonStr != "" {
		return nil, fmt.Errorf("cannot specify both JSON string and file path")
	}

	if filePath == "" && jsonStr == "" {
		return nil, fmt.Errorf("either JSON data or file path is required")
	}

	var jsonData string
	var err error

	if filePath != "" {
		// Read from file
		jsonData, err = readJSONFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to read JSON file: %w", err)
		}
	} else {
		jsonData = jsonStr
	}

	// Validate and parse JSON
	return validateAndParseJSON(jsonData)
}

// validateAndParseJSON validates JSON format and parses to map
func validateAndParseJSON(jsonStr string) (map[string]interface{}, error) {
	if jsonStr == "" {
		return nil, fmt.Errorf("JSON data cannot be empty")
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return nil, fmt.Errorf("invalid JSON format: %w", err)
	}

	return data, nil
}

// readJSONFile reads and returns JSON content from a file
func readJSONFile(filePath string) (string, error) {
	if filePath == "" {
		return "", fmt.Errorf("file path cannot be empty")
	}

	// Basic path validation
	if strings.Contains(filePath, "..") {
		return "", fmt.Errorf("file path cannot contain '..' for security reasons")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", filePath, err)
	}

	return string(data), nil
}
