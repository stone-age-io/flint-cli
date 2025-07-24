package collections

import (
	"fmt"
	"strings"

	"flint-cli/internal/config"
	"flint-cli/internal/pocketbase"
	"flint-cli/internal/utils"
)

// displayListTable displays the results in a user-friendly table format
func displayListTable(result *pocketbase.RecordsList, collection string) error {
	if result == nil || len(result.Items) == 0 {
		fmt.Printf("No %s found.\n", collection)
		return nil
	}

	// Show pagination info
	fmt.Printf("%s (%d-%d of %d total)\n\n", 
		utils.TitleCase(collection),
		((result.Page-1)*result.PerPage)+1,
		min(result.Page*result.PerPage, result.TotalItems),
		result.TotalItems)

	// Display table
	if err := utils.OutputData(result.Items, config.OutputFormatTable); err != nil {
		return fmt.Errorf("failed to display table: %w", err)
	}

	// Show pagination navigation hints
	if result.TotalPages > 1 {
		fmt.Printf("\nPagination:\n")
		if result.Page > 1 {
			prevOffset := (result.Page-2) * result.PerPage
			fmt.Printf("  Previous: --offset %d\n", prevOffset)
		}
		if result.Page < result.TotalPages {
			nextOffset := result.Page * result.PerPage
			fmt.Printf("  Next: --offset %d\n", nextOffset)
		}
		fmt.Printf("  Page %d of %d (use --offset to navigate)\n", 
			result.Page, result.TotalPages)
	}

	return nil
}

// displayGetTable displays a single record in table format
func displayGetTable(record map[string]interface{}, collection, recordID string) error {
	if record == nil {
		return fmt.Errorf("no record data received")
	}

	// Show header
	fmt.Printf("%s Record: %s\n", utils.TitleCase(collection), recordID)
	fmt.Println(strings.Repeat("=", 50))

	// Create a organized display based on record type
	if err := displayRecordDetails(record, collection); err != nil {
		// Fallback to generic table if specific display fails
		return utils.OutputData(record, config.OutputFormatTable)
	}

	return nil
}

// displayRecordDetails displays record details with collection-specific formatting
func displayRecordDetails(record map[string]interface{}, collection string) error {
	// Common fields that should be displayed first
	commonFields := []string{"id", "name", "email", "code", "description", "type"}
	
	// Time fields that should be displayed last
	timeFields := []string{"created", "updated"}

	// Display common fields first
	for _, field := range commonFields {
		if value, exists := record[field]; exists && value != nil {
			fmt.Printf("  %s: %v\n", utils.TitleCase(field), value)
		}
	}

	// Display collection-specific important fields
	switch collection {
	case "users":
		displayUserSpecificFields(record)
	case "edges":
		displayEdgeSpecificFields(record)
	case "things":
		displayThingSpecificFields(record)
	case "organizations":
		displayOrganizationSpecificFields(record)
	case "locations":
		displayLocationSpecificFields(record)
	}

	// Display other fields (excluding common and time fields)
	skipFields := make(map[string]bool)
	for _, field := range append(commonFields, timeFields...) {
		skipFields[field] = true
	}

	// Skip collection-specific fields we already handled
	collectionSpecific := getCollectionSpecificFields(collection)
	for _, field := range collectionSpecific {
		skipFields[field] = true
	}

	for key, value := range record {
		if !skipFields[key] && value != nil {
			fmt.Printf("  %s: %v\n", utils.TitleCase(key), value)
		}
	}

	// Display time fields last
	for _, field := range timeFields {
		if value, exists := record[field]; exists && value != nil {
			fmt.Printf("  %s: %v\n", utils.TitleCase(field), value)
		}
	}

	// Display expanded relations
	if expand, exists := record["expand"]; exists && expand != nil {
		fmt.Printf("\nExpanded Relations:\n")
		if err := utils.OutputData(expand, config.OutputFormatTable); err != nil {
			fmt.Printf("  %v\n", expand)
		}
	}

	return nil
}

// displayUserSpecificFields displays user-specific fields
func displayUserSpecificFields(record map[string]interface{}) {
	userFields := []string{"first_name", "last_name", "username", "current_organization_id", "is_org_admin", "active"}
	for _, field := range userFields {
		if value, exists := record[field]; exists && value != nil {
			displayName := field
			if field == "is_org_admin" {
				displayName = "Org Admin"
			} else if field == "current_organization_id" {
				displayName = "Current Organization"
			}
			fmt.Printf("  %s: %v\n", utils.TitleCase(displayName), value)
		}
	}
}

// displayEdgeSpecificFields displays edge-specific fields
func displayEdgeSpecificFields(record map[string]interface{}) {
	edgeFields := []string{"region", "organization_id", "location_id", "active", "public_key"}
	for _, field := range edgeFields {
		if value, exists := record[field]; exists && value != nil {
			displayName := field
			if field == "organization_id" {
				displayName = "Organization"
			} else if field == "location_id" {
				displayName = "Location"
			} else if field == "public_key" {
				displayName = "Public Key"
			}
			fmt.Printf("  %s: %v\n", utils.TitleCase(displayName), value)
		}
	}
}

// displayThingSpecificFields displays thing-specific fields
func displayThingSpecificFields(record map[string]interface{}) {
	thingFields := []string{"edge_id", "location_id", "organization_id", "active", "mac_address", "ip_address"}
	for _, field := range thingFields {
		if value, exists := record[field]; exists && value != nil {
			displayName := field
			if field == "edge_id" {
				displayName = "Edge"
			} else if field == "location_id" {
				displayName = "Location"
			} else if field == "organization_id" {
				displayName = "Organization"
			} else if field == "mac_address" {
				displayName = "MAC Address"
			} else if field == "ip_address" {
				displayName = "IP Address"
			}
			fmt.Printf("  %s: %v\n", utils.TitleCase(displayName), value)
		}
	}
}

// displayOrganizationSpecificFields displays organization-specific fields
func displayOrganizationSpecificFields(record map[string]interface{}) {
	orgFields := []string{"account_name", "active", "parent_id"}
	for _, field := range orgFields {
		if value, exists := record[field]; exists && value != nil {
			displayName := field
			if field == "account_name" {
				displayName = "Account Name"
			} else if field == "parent_id" {
				displayName = "Parent Organization"
			}
			fmt.Printf("  %s: %v\n", utils.TitleCase(displayName), value)
		}
	}
}

// displayLocationSpecificFields displays location-specific fields
func displayLocationSpecificFields(record map[string]interface{}) {
	locationFields := []string{"path", "organization_id", "edge_id", "parent_id"}
	for _, field := range locationFields {
		if value, exists := record[field]; exists && value != nil {
			displayName := field
			if field == "organization_id" {
				displayName = "Organization"
			} else if field == "edge_id" {
				displayName = "Edge"
			} else if field == "parent_id" {
				displayName = "Parent Location"
			}
			fmt.Printf("  %s: %v\n", utils.TitleCase(displayName), value)
		}
	}
}

// getCollectionSpecificFields returns fields that are handled by collection-specific display functions
func getCollectionSpecificFields(collection string) []string {
	switch collection {
	case "users":
		return []string{"first_name", "last_name", "username", "current_organization_id", "is_org_admin", "active"}
	case "edges":
		return []string{"region", "organization_id", "location_id", "active", "public_key"}
	case "things":
		return []string{"edge_id", "location_id", "organization_id", "active", "mac_address", "ip_address"}
	case "organizations":
		return []string{"account_name", "active", "parent_id"}
	case "locations":
		return []string{"path", "organization_id", "edge_id", "parent_id"}
	default:
		return []string{}
	}
}
