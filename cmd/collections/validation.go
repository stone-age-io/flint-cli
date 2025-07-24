package collections

import (
	"fmt"

	"flint-cli/internal/utils"
)

// validateCreateData validates the JSON data for creating a record
func validateCreateData(data map[string]interface{}, collection string) error {
	if data == nil || len(data) == 0 {
		return fmt.Errorf("record data cannot be empty")
	}

	// Check for fields that should not be manually set
	restrictedFields := []string{"id", "created", "updated"}
	
	for _, field := range restrictedFields {
		if _, exists := data[field]; exists {
			return fmt.Errorf("field '%s' is automatically managed and should not be included", field)
		}
	}

	// Collection-specific validation
	switch collection {
	case "users":
		return validateUserCreateData(data)
	case "organizations":
		return validateOrganizationCreateData(data)
	case "edges":
		return validateEdgeCreateData(data)
	case "things":
		return validateThingCreateData(data)
	case "locations":
		return validateLocationCreateData(data)
	}

	return nil
}

// validateUpdateData validates the JSON data for updating a record
func validateUpdateData(data map[string]interface{}, collection string) error {
	if data == nil || len(data) == 0 {
		return fmt.Errorf("update data cannot be empty")
	}

	// Check for fields that should not be manually updated
	restrictedFields := []string{"id", "created", "updated"}
	
	for _, field := range restrictedFields {
		if _, exists := data[field]; exists {
			return fmt.Errorf("field '%s' is automatically managed and cannot be updated", field)
		}
	}

	// Warn about organization_id changes
	if _, exists := data["organization_id"]; exists {
		utils.PrintWarning("Updating organization_id may cause access issues. This field is typically managed by your context.")
	}

	// Collection-specific validation
	switch collection {
	case "users":
		return validateUserUpdateData(data)
	case "organizations":
		return validateOrganizationUpdateData(data)
	case "edges":
		return validateEdgeUpdateData(data)
	case "things":
		return validateThingUpdateData(data)
	case "locations":
		return validateLocationUpdateData(data)
	}

	return nil
}

// validateUserCreateData validates user creation data
func validateUserCreateData(data map[string]interface{}) error {
	requiredFields := []string{"email", "password"}
	
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("field '%s' is required for user creation", field)
		}
	}

	// Validate email format if present
	if email, ok := data["email"].(string); ok {
		if err := utils.ValidateEmail(email); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}
	}

	// Warn about organization_id - it's usually set automatically
	if _, exists := data["organization_id"]; exists {
		utils.PrintWarning("organization_id is typically set automatically based on your context")
	}

	return nil
}

// validateOrganizationCreateData validates organization creation data
func validateOrganizationCreateData(data map[string]interface{}) error {
	requiredFields := []string{"name", "code", "account_name"}
	
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("field '%s' is required for organization creation", field)
		}
	}

	return nil
}

// validateEdgeCreateData validates edge creation data
func validateEdgeCreateData(data map[string]interface{}) error {
	requiredFields := []string{"name", "code", "type", "region"}
	
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("field '%s' is required for edge creation", field)
		}
	}

	// organization_id is set automatically based on context
	if _, exists := data["organization_id"]; exists {
		utils.PrintWarning("organization_id is automatically set based on your current context")
	}

	return nil
}

// validateThingCreateData validates thing creation data
func validateThingCreateData(data map[string]interface{}) error {
	requiredFields := []string{"name", "code", "type", "edge_id"}
	
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("field '%s' is required for thing creation", field)
		}
	}

	// organization_id is set automatically based on context
	if _, exists := data["organization_id"]; exists {
		utils.PrintWarning("organization_id is automatically set based on your current context")
	}

	return nil
}

// validateLocationCreateData validates location creation data
func validateLocationCreateData(data map[string]interface{}) error {
	requiredFields := []string{"name", "type", "code"}
	
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("field '%s' is required for location creation", field)
		}
	}

	// organization_id is set automatically based on context
	if _, exists := data["organization_id"]; exists {
		utils.PrintWarning("organization_id is automatically set based on your current context")
	}

	return nil
}

// validateUserUpdateData validates user update data
func validateUserUpdateData(data map[string]interface{}) error {
	// Validate email format if present
	if email, ok := data["email"].(string); ok {
		if err := utils.ValidateEmail(email); err != nil {
			return fmt.Errorf("invalid email: %w", err)
		}
	}

	// Warn about sensitive field updates
	sensitiveFields := []string{"password", "current_organization_id"}
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			utils.PrintWarning(fmt.Sprintf("Updating '%s' - ensure this is intentional", field))
		}
	}

	return nil
}

// validateOrganizationUpdateData validates organization update data
func validateOrganizationUpdateData(data map[string]interface{}) error {
	// Warn about critical field updates
	criticalFields := []string{"code", "account_name"}
	for _, field := range criticalFields {
		if _, exists := data[field]; exists {
			utils.PrintWarning(fmt.Sprintf("Updating '%s' may affect system integrations", field))
		}
	}

	return nil
}

// validateEdgeUpdateData validates edge update data
func validateEdgeUpdateData(data map[string]interface{}) error {
	// Warn about fields that may affect connectivity
	connectivityFields := []string{"region", "public_key", "private_key"}
	for _, field := range connectivityFields {
		if _, exists := data[field]; exists {
			utils.PrintWarning(fmt.Sprintf("Updating '%s' may affect edge connectivity", field))
		}
	}

	return nil
}

// validateThingUpdateData validates thing update data
func validateThingUpdateData(data map[string]interface{}) error {
	// Warn about fields that may affect device operation
	operationalFields := []string{"edge_id", "mac_address", "ip_address"}
	for _, field := range operationalFields {
		if _, exists := data[field]; exists {
			utils.PrintWarning(fmt.Sprintf("Updating '%s' may affect device operation", field))
		}
	}

	return nil
}

// validateLocationUpdateData validates location update data
func validateLocationUpdateData(data map[string]interface{}) error {
	// Warn about structural changes
	if _, exists := data["parent_id"]; exists {
		utils.PrintWarning("Updating parent_id will change the location hierarchy")
	}

	if _, exists := data["path"]; exists {
		utils.PrintWarning("Updating path should be done carefully to maintain location hierarchy")
	}

	return nil
}
