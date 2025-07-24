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

	// Validate credentials
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

	// Check if user has organizations field
	orgsInterface, exists := authRecord["organizations"]
	if !exists {
		return fmt.Errorf("user has no organization memberships")
	}

	// Handle organizations field (could be array of IDs or objects)
	var userOrgIDs []string
	
	switch orgs := orgsInterface.(type) {
	case []interface{}:
		// Array of organization IDs or objects
		for _, org := range orgs {
			switch orgData := org.(type) {
			case string:
				// Simple ID
				userOrgIDs = append(userOrgIDs, orgData)
			case map[string]interface{}:
				// Organization object with ID
				if id, ok := orgData["id"].(string); ok {
					userOrgIDs = append(userOrgIDs, id)
				}
			}
		}
	case []string:
		// Array of organization ID strings
		userOrgIDs = orgs
	case string:
		// Single organization ID
		userOrgIDs = []string{orgs}
	default:
		return fmt.Errorf("invalid organizations field format")
	}

	// Check if the user belongs to the specified organization
	for _, orgID := range userOrgIDs {
		if orgID == organizationID {
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

	// Check if organizations are already expanded in the auth record
	if orgsInterface, exists := authRecord["organizations"]; exists {
		if orgs, ok := orgsInterface.([]interface{}); ok {
			var organizations []map[string]interface{}
			for _, org := range orgs {
				if orgMap, ok := org.(map[string]interface{}); ok {
					organizations = append(organizations, orgMap)
				}
			}
			return organizations, nil
		}
	}

	// If not expanded, we need to fetch the user record with organization expansion
	userID, ok := authRecord["id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid authentication record: missing user ID")
	}

	// Get user record with expanded organizations
	userRecord, err := c.GetRecord("users", userID, []string{"organizations"})
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	// Extract organizations from the expanded record
	if orgsInterface, exists := userRecord["expand"].(map[string]interface{})["organizations"]; exists {
		if orgs, ok := orgsInterface.([]interface{}); ok {
			var organizations []map[string]interface{}
			for _, org := range orgs {
				if orgMap, ok := org.(map[string]interface{}); ok {
					organizations = append(organizations, orgMap)
				}
			}
			return organizations, nil
		}
	}

	return []map[string]interface{}{}, nil
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
