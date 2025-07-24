package pocketbase

import (
	"encoding/json"
	"fmt"
	"time"

	"flint-cli/internal/config"
	"flint-cli/internal/utils"
)

// AuthResponse represents a PocketBase authentication response
type AuthResponse struct {
	Token  string                 `json:"token"`
	Record map[string]interface{} `json:"record"`
	Meta   map[string]interface{} `json:"meta,omitempty"`
}

// AuthRequest represents authentication request data
type AuthRequest struct {
	Identity string `json:"identity"`
	Password string `json:"password"`
}

// Authenticate performs authentication against a specific collection
func (c *Client) Authenticate(collection, identity, password string) (*AuthResponse, error) {
	// Validate collection
	if err := config.ValidateAuthCollection(collection); err != nil {
		return nil, fmt.Errorf("invalid auth collection: %w", err)
	}

	// Basic validation
	if identity == "" {
		return nil, fmt.Errorf("identity (email/username) is required")
	}
	if password == "" {
		return nil, fmt.Errorf("password is required")
	}

	// Prepare authentication request
	authData := AuthRequest{
		Identity: identity,
		Password: password,
	}

	// Make authentication request
	endpoint := fmt.Sprintf("collections/%s/auth-with-password", collection)
	
	utils.PrintDebug(fmt.Sprintf("Authenticating with collection: %s", collection))
	
	resp, err := c.makeRequest("POST", endpoint, authData)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Parse response
	var authResp AuthResponse
	if err := json.Unmarshal(resp.Body(), &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse authentication response: %w", err)
	}

	// Set authentication token
	c.SetAuthToken(authResp.Token)
	c.authRecord = authResp.Record

	utils.PrintDebug("Authentication successful")
	
	return &authResp, nil
}

// RefreshAuth refreshes the current authentication token
func (c *Client) RefreshAuth(collection string) (*AuthResponse, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("collections/%s/auth-refresh", collection)
	
	utils.PrintDebug("Refreshing authentication token")
	
	resp, err := c.makeRequest("POST", endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to refresh authentication: %w", err)
	}

	var authResp AuthResponse
	if err := json.Unmarshal(resp.Body(), &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	// Update authentication
	c.SetAuthToken(authResp.Token)
	c.authRecord = authResp.Record

	utils.PrintDebug("Authentication refreshed successfully")
	
	return &authResp, nil
}

// ValidateAuth checks if the current authentication is valid
func (c *Client) ValidateAuth(collection string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("not authenticated")
	}

	endpoint := fmt.Sprintf("collections/%s/auth-refresh", collection)
	
	// Try to refresh - if it fails, auth is invalid
	_, err := c.makeRequest("POST", endpoint, nil)
	if err != nil {
		return fmt.Errorf("authentication is invalid or expired: %w", err)
	}

	return nil
}

// GetAuthenticatedUser returns the currently authenticated user record
func (c *Client) GetAuthenticatedUser() (map[string]interface{}, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("not authenticated")
	}

	if c.authRecord == nil {
		return nil, fmt.Errorf("no authentication record available")
	}

	return c.authRecord, nil
}

// ValidateOrganizationAccess validates that the authenticated user belongs to the specified organization
func (c *Client) ValidateOrganizationAccess(organizationID string) error {
	if !c.IsAuthenticated() {
		return fmt.Errorf("authentication required")
	}

	authRecord := c.GetAuthRecord()
	if authRecord == nil {
		return fmt.Errorf("no authentication record available")
	}

	utils.PrintDebug(fmt.Sprintf("Validating access to organization: %s", organizationID))

	// Get user organizations to check membership
	userOrgs, err := c.GetUserOrganizations()
	if err != nil {
		return fmt.Errorf("failed to get user organizations: %w", err)
	}

	// Check if the user belongs to the specified organization
	for _, org := range userOrgs {
		if orgID, ok := org["id"].(string); ok && orgID == organizationID {
			utils.PrintDebug(fmt.Sprintf("User has access to organization: %s", organizationID))
			return nil // User has access
		}
	}

	return fmt.Errorf("user does not belong to organization '%s'", organizationID)
}

// GetUserOrganizations returns the organizations the authenticated user belongs to
func (c *Client) GetUserOrganizations() ([]map[string]interface{}, error) {
	if !c.IsAuthenticated() {
		return nil, fmt.Errorf("authentication required")
	}

	authRecord := c.GetAuthRecord()
	if authRecord == nil {
		return nil, fmt.Errorf("no authentication record available")
	}

	utils.PrintDebug("Retrieving user organizations")

	// Check if organizations are already expanded in the auth record
	if orgsInterface, exists := authRecord["expand"]; exists {
		if expandMap, ok := orgsInterface.(map[string]interface{}); ok {
			if orgData, exists := expandMap["organizations"]; exists {
				utils.PrintDebug("Found expanded organizations in auth record")
				return c.parseOrganizationData(orgData)
			}
		}
	}

	// Check if organizations field exists in auth record (could be IDs)
	if orgsInterface, exists := authRecord["organizations"]; exists {
		utils.PrintDebug("Found organizations field in auth record")
		// If it's already expanded objects, use them
		if orgs, err := c.parseOrganizationData(orgsInterface); err == nil && len(orgs) > 0 {
			// Check if first item has full organization data (not just ID)
			if firstOrg := orgs[0]; len(firstOrg) > 1 || (len(firstOrg) == 1 && firstOrg["id"] == nil) {
				utils.PrintDebug("Organizations are already expanded")
				return orgs, nil
			}
		}
	}

	// Need to fetch the user record with organization expansion
	userID, ok := authRecord["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid authentication record: missing user ID")
	}

	utils.PrintDebug(fmt.Sprintf("Fetching user record with expanded organizations: %s", userID))

	// Get user record with expanded organizations
	userRecord, err := c.GetRecord("users", userID, []string{"organizations"})
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	// Extract organizations from the expanded record
	if expandInterface, exists := userRecord["expand"]; exists {
		if expandMap, ok := expandInterface.(map[string]interface{}); ok {
			if orgsInterface, exists := expandMap["organizations"]; exists {
				utils.PrintDebug("Found expanded organizations in user record")
				return c.parseOrganizationData(orgsInterface)
			}
		}
	}

	utils.PrintDebug("No organizations found for user")
	return []map[string]interface{}{}, nil
}

// parseOrganizationData helper function to parse organization data from various formats
func (c *Client) parseOrganizationData(orgsInterface interface{}) ([]map[string]interface{}, error) {
	switch orgs := orgsInterface.(type) {
	case []interface{}:
		var organizations []map[string]interface{}
		for _, org := range orgs {
			if orgMap, ok := org.(map[string]interface{}); ok {
				organizations = append(organizations, orgMap)
			}
		}
		utils.PrintDebug(fmt.Sprintf("Parsed %d organizations from interface slice", len(organizations)))
		return organizations, nil
	case []map[string]interface{}:
		utils.PrintDebug(fmt.Sprintf("Parsed %d organizations from map slice", len(orgs)))
		return orgs, nil
	case map[string]interface{}:
		// Single organization object
		utils.PrintDebug("Parsed single organization")
		return []map[string]interface{}{orgs}, nil
	default:
		return nil, fmt.Errorf("invalid organization data format")
	}
}

// GetCurrentOrganizationID returns the user's current organization ID from their record
func (c *Client) GetCurrentOrganizationID() string {
	if !c.IsAuthenticated() {
		return ""
	}

	authRecord := c.GetAuthRecord()
	if authRecord == nil {
		return ""
	}

	// Check for current_organization_id field
	if currentOrgID, ok := authRecord["current_organization_id"].(string); ok && currentOrgID != "" {
		utils.PrintDebug(fmt.Sprintf("Found current organization ID in auth record: %s", currentOrgID))
		return currentOrgID
	}

	utils.PrintDebug("No current organization ID found in auth record")
	return ""
}

// UpdateAuthContextFromResponse updates a context with authentication data
func UpdateAuthContextFromResponse(ctx *config.Context, authResp *AuthResponse, organizationID string) error {
	if authResp == nil {
		return fmt.Errorf("authentication response is nil")
	}

	// Update context with auth data
	ctx.PocketBase.AuthToken = authResp.Token
	ctx.PocketBase.AuthRecord = authResp.Record
	
	// Set expiration (PocketBase tokens typically last 7 days)
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	ctx.PocketBase.AuthExpires = &expiresAt

	// Set organization if provided
	if organizationID != "" {
		ctx.PocketBase.OrganizationID = organizationID
	}

	return nil
}

// IsAuthValid checks if the authentication in a context is still valid
func IsAuthValid(ctx *config.Context) bool {
	if ctx.PocketBase.AuthToken == "" {
		return false
	}

	if ctx.PocketBase.AuthExpires == nil {
		// No expiration set, assume valid for backward compatibility
		return true
	}

	// Check if token has expired (with 5-minute buffer)
	return time.Now().Before(ctx.PocketBase.AuthExpires.Add(-5 * time.Minute))
}

// GetOrganizationInfo extracts organization information from auth record
func GetOrganizationInfo(authRecord map[string]interface{}) []map[string]interface{} {
	if authRecord == nil {
		return nil
	}

	// Try to get organizations from the auth record
	orgsInterface, exists := authRecord["organizations"]
	if !exists {
		return nil
	}

	switch orgs := orgsInterface.(type) {
	case []interface{}:
		var organizations []map[string]interface{}
		for _, org := range orgs {
			if orgMap, ok := org.(map[string]interface{}); ok {
				organizations = append(organizations, orgMap)
			}
		}
		return organizations
	case []map[string]interface{}:
		return orgs
	}

	return nil
}

// GetCollectionDisplayName returns a human-readable name for auth collections
func GetCollectionDisplayName(collection string) string {
	switch collection {
	case config.AuthCollectionUsers:
		return "Users (Human Administrators)"
	case config.AuthCollectionClients:
		return "Clients (NATS Client Entities)"
	case config.AuthCollectionEdges:
		return "Edges (Edge Device Authentication)"
	case config.AuthCollectionThings:
		return "Things (Individual Device Authentication)"
	case config.AuthCollectionServiceUsers:
		return "Service Users (System Service Accounts)"
	default:
		return collection
	}
}

// ValidateAuthCollection ensures the collection supports authentication
func ValidateAuthCollection(collection string) error {
	validCollections := []string{
		config.AuthCollectionUsers,
		config.AuthCollectionClients,
		config.AuthCollectionEdges,
		config.AuthCollectionThings,
		config.AuthCollectionServiceUsers,
	}

	for _, valid := range validCollections {
		if collection == valid {
			return nil
		}
	}

	return fmt.Errorf("collection '%s' does not support authentication", collection)
}
