package collections

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/fatih/color"
	"flint-cli/internal/config"
	"flint-cli/internal/pocketbase"
	"flint-cli/internal/utils"
)

// handleListAction handles the list action for a collection
func handleListAction(ctx *config.Context, collection string, args []string) error {
	// Use global flags instead of manual parsing
	// Note: args are any remaining positional arguments (none expected for list)
	if len(args) > 0 {
		return fmt.Errorf("list action does not accept positional arguments, use flags instead")
	}

	// Create PocketBase client
	client := createPocketBaseClient(ctx)

	// Build list options from global flags
	options := &pocketbase.ListOptions{
		Page:    calculatePage(offsetFlag, limitFlag),
		PerPage: limitFlag,
		Filter:  filterFlag,
		Sort:    sortFlag,
		Fields:  fieldsFlag,
		Expand:  expandFlag,
	}

	// Validate pagination parameters
	if err := validatePaginationOptions(options); err != nil {
		return fmt.Errorf("invalid pagination options: %w", err)
	}

	utils.PrintDebug(fmt.Sprintf("Listing records from collection '%s' with options: page=%d, perPage=%d, filter='%s', sort='%s', fields=%v, expand=%v", 
		collection, options.Page, options.PerPage, options.Filter, options.Sort, options.Fields, options.Expand))

	// List records from PocketBase
	result, err := client.ListRecords(collection, options)
	if err != nil {
		if pbErr, ok := err.(*pocketbase.PocketBaseError); ok {
			utils.PrintError(fmt.Errorf("%s", pbErr.GetFriendlyMessage()))
			if suggestion := pbErr.GetSuggestion(); suggestion != "" {
				fmt.Printf("\nSuggestion: %s\n", suggestion)
			}
			return fmt.Errorf("failed to list records")
		}
		return fmt.Errorf("failed to list records: %w", err)
	}

	// Display results based on output format
	outputFormat := outputFlag
	if outputFormat == "" {
		outputFormat = config.Global.OutputFormat
	}

	switch outputFormat {
	case config.OutputFormatJSON:
		return utils.OutputData(result, config.OutputFormatJSON)
	case config.OutputFormatYAML:
		return utils.OutputData(result, config.OutputFormatYAML) 
	case config.OutputFormatTable:
		return displayListTable(result, collection)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// handleGetAction handles the get action for a collection
func handleGetAction(ctx *config.Context, collection string, args []string) error {
	// Expect exactly one positional argument (record ID)
	if len(args) != 1 {
		return fmt.Errorf("get requires exactly one record ID argument")
	}
	
	recordID := args[0]

	// Basic validation - just check that ID is not empty
	// Let PocketBase handle format validation since IDs are configurable
	if strings.TrimSpace(recordID) == "" {
		return fmt.Errorf("record ID cannot be empty")
	}

	// Create PocketBase client
	client := createPocketBaseClient(ctx)

	utils.PrintDebug(fmt.Sprintf("Getting record '%s' from collection '%s' with expand=%v", recordID, collection, expandFlag))

	// Get record from PocketBase
	record, err := client.GetRecord(collection, recordID, expandFlag)
	if err != nil {
		if pbErr, ok := err.(*pocketbase.PocketBaseError); ok {
			utils.PrintError(fmt.Errorf("%s", pbErr.GetFriendlyMessage()))
			if suggestion := pbErr.GetSuggestion(); suggestion != "" {
				fmt.Printf("\nSuggestion: %s\n", suggestion)
			}
			return fmt.Errorf("failed to get record")
		}
		return fmt.Errorf("failed to get record: %w", err)
	}

	// Display result based on output format
	outputFormat := outputFlag
	if outputFormat == "" {
		outputFormat = config.Global.OutputFormat
	}

	switch outputFormat {
	case config.OutputFormatJSON:
		return utils.OutputData(record, config.OutputFormatJSON)  
	case config.OutputFormatYAML:
		return utils.OutputData(record, config.OutputFormatYAML)
	case config.OutputFormatTable:
		return displayGetTable(record, collection, recordID)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// handleCreateAction handles the create action for a collection
func handleCreateAction(ctx *config.Context, collection string, args []string) error {
	// Get JSON data from positional argument or file flag
	var jsonData string
	if len(args) > 0 {
		jsonData = args[0]
	}

	// Parse JSON input from string or file
	data, err := parseJSONInput(jsonData, fileFlag)
	if err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}

	// Validate that we don't have restricted fields
	if err := validateCreateData(data, collection); err != nil {
		return fmt.Errorf("invalid create data: %w", err)
	}

	// Create PocketBase client
	client := createPocketBaseClient(ctx)

	utils.PrintDebug(fmt.Sprintf("Creating record in collection '%s' with data: %+v", collection, data))

	// Create record in PocketBase
	record, err := client.CreateRecord(collection, data)
	if err != nil {
		if pbErr, ok := err.(*pocketbase.PocketBaseError); ok {
			utils.PrintError(fmt.Errorf("%s", pbErr.GetFriendlyMessage()))
			if suggestion := pbErr.GetSuggestion(); suggestion != "" {
				fmt.Printf("\nSuggestion: %s\n", suggestion)
			}
			return fmt.Errorf("failed to create record")
		}
		return fmt.Errorf("failed to create record: %w", err)
	}

	// Display success message
	recordID := ""
	if id, ok := record["id"].(string); ok {
		recordID = id
	}

	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Record created successfully!\n", green("✓"))
	
	if recordID != "" {
		fmt.Printf("  Record ID: %s\n", recordID)
		fmt.Printf("  Collection: %s\n", collection)
		
		// Show record name if available
		if name, ok := record["name"].(string); ok && name != "" {
			fmt.Printf("  Name: %s\n", name)
		}
	}

	// Display result based on output format
	outputFormat := outputFlag
	if outputFormat == "" {
		outputFormat = config.Global.OutputFormat
	}

	fmt.Printf("\nCreated Record:\n")
	switch outputFormat {
	case config.OutputFormatJSON:
		return utils.OutputData(record, config.OutputFormatJSON)
	case config.OutputFormatYAML:
		return utils.OutputData(record, config.OutputFormatYAML)
	case config.OutputFormatTable:
		return utils.OutputData(record, config.OutputFormatTable)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// handleUpdateAction handles the update action for a collection
func handleUpdateAction(ctx *config.Context, collection string, args []string) error {
	// Expect at least record ID, optionally JSON data
	if len(args) < 1 {
		return fmt.Errorf("update requires a record ID argument")
	}
	
	recordID := args[0]
	var jsonData string
	if len(args) > 1 {
		jsonData = args[1]
	}

	// Basic validation - just check that ID is not empty
	// Let PocketBase handle format validation since IDs are configurable
	if strings.TrimSpace(recordID) == "" {
		return fmt.Errorf("record ID cannot be empty")
	}

	// Parse JSON input from string or file
	data, err := parseJSONInput(jsonData, fileFlag)
	if err != nil {
		return fmt.Errorf("invalid JSON input: %w", err)
	}

	// Validate that we don't have restricted fields
	if err := validateUpdateData(data, collection); err != nil {
		return fmt.Errorf("invalid update data: %w", err)
	}

	// Create PocketBase client
	client := createPocketBaseClient(ctx)

	utils.PrintDebug(fmt.Sprintf("Updating record '%s' in collection '%s' with data: %+v", recordID, collection, data))

	// Update record in PocketBase
	record, err := client.UpdateRecord(collection, recordID, data)
	if err != nil {
		if pbErr, ok := err.(*pocketbase.PocketBaseError); ok {
			utils.PrintError(fmt.Errorf("%s", pbErr.GetFriendlyMessage()))
			if suggestion := pbErr.GetSuggestion(); suggestion != "" {
				fmt.Printf("\nSuggestion: %s\n", suggestion)
			}
			return fmt.Errorf("failed to update record")
		}
		return fmt.Errorf("failed to update record: %w", err)
	}

	// Display success message
	green := color.New(color.FgGreen).SprintFunc()
	fmt.Printf("%s Record updated successfully!\n", green("✓"))
	
	fmt.Printf("  Record ID: %s\n", recordID)
	fmt.Printf("  Collection: %s\n", collection)
	
	// Show record name if available
	if name, ok := record["name"].(string); ok && name != "" {
		fmt.Printf("  Name: %s\n", name)
	}

	// Show which fields were updated
	fieldCount := len(data)
	if fieldCount > 0 {
		fmt.Printf("  Updated %d field(s)\n", fieldCount)
	}

	// Display result based on output format
	outputFormat := outputFlag
	if outputFormat == "" {
		outputFormat = config.Global.OutputFormat
	}

	fmt.Printf("\nUpdated Record:\n")
	switch outputFormat {
	case config.OutputFormatJSON:
		return utils.OutputData(record, config.OutputFormatJSON)
	case config.OutputFormatYAML:
		return utils.OutputData(record, config.OutputFormatYAML)
	case config.OutputFormatTable:
		return utils.OutputData(record, config.OutputFormatTable)
	default:
		return fmt.Errorf("unsupported output format: %s", outputFormat)
	}
}

// handleDeleteAction handles the delete action for a collection
func handleDeleteAction(ctx *config.Context, collection string, args []string) error {
	// Expect exactly one positional argument (record ID)
	if len(args) != 1 {
		return fmt.Errorf("delete requires exactly one record ID argument")
	}
	
	recordID := args[0]

	// Basic validation - just check that ID is not empty
	// Let PocketBase handle format validation since IDs are configurable
	if strings.TrimSpace(recordID) == "" {
		return fmt.Errorf("record ID cannot be empty")
	}

	// Create PocketBase client
	client := createPocketBaseClient(ctx)

	// Get record details for confirmation (unless forced)
	var record map[string]interface{}
	if !forceFlag {
		utils.PrintDebug(fmt.Sprintf("Fetching record details for confirmation: %s", recordID))
		
		var err error
		record, err = client.GetRecord(collection, recordID, nil)
		if err != nil {
			if pbErr, ok := err.(*pocketbase.PocketBaseError); ok {
				utils.PrintError(fmt.Errorf("%s", pbErr.GetFriendlyMessage()))
				if suggestion := pbErr.GetSuggestion(); suggestion != "" {
					fmt.Printf("\nSuggestion: %s\n", suggestion)
				}
				return fmt.Errorf("failed to retrieve record for confirmation")
			}
			return fmt.Errorf("failed to retrieve record: %w", err)
		}

		// Show confirmation prompt with record details
		if err := confirmDeletion(collection, recordID, record); err != nil {
			return err
		}
	}

	utils.PrintDebug(fmt.Sprintf("Deleting record '%s' from collection '%s'", recordID, collection))

	// Delete record from PocketBase
	err := client.DeleteRecord(collection, recordID)
	if err != nil {
		if pbErr, ok := err.(*pocketbase.PocketBaseError); ok {
			utils.PrintError(fmt.Errorf("%s", pbErr.GetFriendlyMessage()))
			if suggestion := pbErr.GetSuggestion(); suggestion != "" {
				fmt.Printf("\nSuggestion: %s\n", suggestion)
			}
			return fmt.Errorf("failed to delete record")
		}
		return fmt.Errorf("failed to delete record: %w", err)
	}

	// Display success message (unless quiet mode)
	if !quietFlag {
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Record deleted successfully!\n", green("✓"))
		fmt.Printf("  Record ID: %s\n", recordID)
		fmt.Printf("  Collection: %s\n", collection)
		
		// Show record name if we have it
		if record != nil {
			if name, ok := record["name"].(string); ok && name != "" {
				fmt.Printf("  Name: %s\n", name)
			}
		}
	}

	return nil
}

// Helper functions

// calculatePage calculates the page number from offset and limit
func calculatePage(offset, limit int) int {
	if offset <= 0 || limit <= 0 {
		return 1
	}
	return (offset / limit) + 1
}

// validatePaginationOptions validates pagination parameters
func validatePaginationOptions(options *pocketbase.ListOptions) error {
	if options.PerPage < 1 {
		return fmt.Errorf("limit must be at least 1")
	}
	if options.PerPage > 500 {
		return fmt.Errorf("limit cannot exceed 500 records")
	}
	if options.Page < 1 {
		return fmt.Errorf("page must be at least 1")
	}
	return nil
}

// confirmDeletion prompts the user to confirm deletion and shows record details
func confirmDeletion(collection, recordID string, record map[string]interface{}) error {
	red := color.New(color.FgRed).SprintFunc()
	yellow := color.New(color.FgYellow).SprintFunc()
	bold := color.New(color.Bold).SprintFunc()

	fmt.Printf("%s Record to be deleted:\n", red("⚠"))
	fmt.Printf("  Collection: %s\n", bold(collection))
	fmt.Printf("  Record ID: %s\n", recordID)

	// Show key record details
	if record != nil {
		if name, ok := record["name"].(string); ok && name != "" {
			fmt.Printf("  Name: %s\n", name)
		}
		if email, ok := record["email"].(string); ok && email != "" {
			fmt.Printf("  Email: %s\n", email)
		}
		if code, ok := record["code"].(string); ok && code != "" {
			fmt.Printf("  Code: %s\n", code)
		}
		if description, ok := record["description"].(string); ok && description != "" {
			descDisplay := description
			if len(descDisplay) > 50 {
				descDisplay = descDisplay[:47] + "..."
			}
			fmt.Printf("  Description: %s\n", descDisplay)
		}
	}

	// Show collection-specific warnings
	showDeletionWarnings(collection)

	fmt.Printf("\n%s This action cannot be undone.\n", yellow("Warning:"))
	fmt.Print("Are you sure you want to delete this record? (y/N): ")

	reader := bufio.NewReader(os.Stdin)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("failed to read confirmation: %w", err)
	}

	response = strings.TrimSpace(strings.ToLower(response))
	if response != "y" && response != "yes" {
		fmt.Println("Deletion cancelled.")
		return fmt.Errorf("deletion cancelled by user")
	}

	return nil
}

// showDeletionWarnings displays collection-specific warnings about deletion impact
func showDeletionWarnings(collection string) {
	yellow := color.New(color.FgYellow).SprintFunc()

	switch collection {
	case "organizations":
		fmt.Printf("  %s Deleting an organization may affect all associated resources\n", yellow("⚠"))
		fmt.Printf("  %s including users, edges, things, and locations.\n", yellow("⚠"))
		
	case "edges":
		fmt.Printf("  %s Deleting an edge may affect all connected things and locations.\n", yellow("⚠"))
		fmt.Printf("  %s Consider moving things to another edge first.\n", yellow("⚠"))
		
	case "users":
		fmt.Printf("  %s Deleting a user will remove their access and may affect\n", yellow("⚠"))
		fmt.Printf("  %s any resources they created or manage.\n", yellow("⚠"))
		
	case "locations":
		fmt.Printf("  %s Deleting a location may affect child locations and\n", yellow("⚠"))
		fmt.Printf("  %s any things assigned to this location.\n", yellow("⚠"))
		
	case "things":
		fmt.Printf("  %s Deleting a thing will permanently remove the device\n", yellow("⚠"))
		fmt.Printf("  %s from your Stone-Age.io system.\n", yellow("⚠"))
		
	case "clients":
		fmt.Printf("  %s Deleting a client will invalidate NATS authentication\n", yellow("⚠"))
		fmt.Printf("  %s and may break messaging functionality.\n", yellow("⚠"))
		
	case "audit_logs":
		fmt.Printf("  %s Audit logs provide important security and compliance\n", yellow("⚠"))
		fmt.Printf("  %s information. Deletion may not be permitted.\n", yellow("⚠"))
	}
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
