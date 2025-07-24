package pocketbase

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
)

// PocketBaseError represents a structured PocketBase API error
type PocketBaseError struct {
	StatusCode int                    `json:"code"`
	Message    string                 `json:"message"`
	Data       map[string]interface{} `json:"data,omitempty"`
	RawBody    string                 `json:"-"`
}

// Error implements the error interface
func (e *PocketBaseError) Error() string {
	return e.GetFriendlyMessage()
}

// NewPocketBaseError creates a PocketBaseError from an HTTP response
func NewPocketBaseError(resp *resty.Response) *PocketBaseError {
	err := &PocketBaseError{
		StatusCode: resp.StatusCode(),
		RawBody:    string(resp.Body()),
	}

	// Try to parse error response JSON
	var errorResp struct {
		Code    int                    `json:"code"`
		Message string                 `json:"message"`
		Data    map[string]interface{} `json:"data"`
	}

	if jsonErr := json.Unmarshal(resp.Body(), &errorResp); jsonErr == nil {
		err.StatusCode = errorResp.Code
		err.Message = errorResp.Message
		err.Data = errorResp.Data
	} else {
		// Fallback to status code message
		err.Message = fmt.Sprintf("HTTP %d: %s", resp.StatusCode(), resp.Status())
	}

	return err
}

// GetFriendlyMessage returns a user-friendly error message for Stone-Age.io operations
func (e *PocketBaseError) GetFriendlyMessage() string {
	// Handle specific HTTP status codes first
	switch e.StatusCode {
	case 400:
		return e.handleBadRequestError()
	case 401:
		return e.handleUnauthorizedError()
	case 403:
		return e.handleForbiddenError()
	case 404:
		return e.handleNotFoundError()
	case 429:
		return "rate limit exceeded. Please wait a moment before trying again"
	case 500:
		return "Stone-Age.io server error. Please try again later or contact support"
	case 503:
		return "Stone-Age.io service is temporarily unavailable. Please try again later"
	}

	// Handle specific error messages
	msgLower := strings.ToLower(e.Message)
	
	// Authentication errors
	if strings.Contains(msgLower, "invalid credentials") ||
	   strings.Contains(msgLower, "wrong email or password") ||
	   strings.Contains(msgLower, "invalid login credentials") {
		return "invalid email or password. Please check your credentials and try again"
	}

	if strings.Contains(msgLower, "auth record not found") {
		return "account not found. Please verify your email address or contact your organization administrator"
	}

	// Organization-related errors
	if strings.Contains(msgLower, "organization") {
		if strings.Contains(msgLower, "not found") {
			return "organization not found. Please verify the organization ID or contact support"
		}
		if strings.Contains(msgLower, "access denied") || strings.Contains(msgLower, "permission") {
			return "you don't have permission to access this organization. Please contact your organization administrator"
		}
	}

	// Collection-related errors
	if strings.Contains(msgLower, "collection") {
		if strings.Contains(msgLower, "not found") {
			return "the requested resource collection was not found. This may indicate a configuration issue"
		}
	}

	// Network/connection errors
	if strings.Contains(msgLower, "connection") || strings.Contains(msgLower, "timeout") {
		return "connection error. Please check your network connection and the Stone-Age.io server URL"
	}

	// Token/session errors
	if strings.Contains(msgLower, "token") || strings.Contains(msgLower, "expired") {
		return "your session has expired. Please authenticate again using 'flint auth pb'"
	}

	// Validation errors
	if e.Data != nil && len(e.Data) > 0 {
		return e.handleValidationErrors()
	}

	// Default fallback
	if e.Message != "" {
		return fmt.Sprintf("Stone-Age.io error: %s", e.Message)
	}

	return fmt.Sprintf("unexpected error occurred (HTTP %d). Please try again or contact support", e.StatusCode)
}

// handleBadRequestError handles 400 Bad Request errors
func (e *PocketBaseError) handleBadRequestError() string {
	if e.Data != nil && len(e.Data) > 0 {
		return e.handleValidationErrors()
	}

	msgLower := strings.ToLower(e.Message)
	
	if strings.Contains(msgLower, "invalid json") {
		return "invalid request format. This appears to be a client error - please report this issue"
	}

	if strings.Contains(msgLower, "missing required") {
		return "required information is missing from your request. Please check your input and try again"
	}

	return fmt.Sprintf("invalid request: %s", e.Message)
}

// handleUnauthorizedError handles 401 Unauthorized errors
func (e *PocketBaseError) handleUnauthorizedError() string {
	msgLower := strings.ToLower(e.Message)

	if strings.Contains(msgLower, "missing authorization header") ||
	   strings.Contains(msgLower, "invalid auth token") ||
	   strings.Contains(msgLower, "expired") {
		return "authentication required. Please run 'flint auth pb' to authenticate"
	}

	if strings.Contains(msgLower, "invalid credentials") {
		return "invalid credentials. Please check your email and password"
	}

	return "authentication failed. Please run 'flint auth pb' to authenticate"
}

// handleForbiddenError handles 403 Forbidden errors
func (e *PocketBaseError) handleForbiddenError() string {
	msgLower := strings.ToLower(e.Message)

	// Organization-specific permission errors
	if strings.Contains(msgLower, "organization") {
		return "you don't have permission to access resources in this organization. Please verify your organization membership or contact your administrator"
	}

	// Collection-specific permission errors
	if strings.Contains(msgLower, "collection") {
		return "you don't have permission to access this collection. Please check your role permissions"
	}

	// General permission error
	return "access denied. You don't have permission to perform this action. Please contact your organization administrator"
}

// handleNotFoundError handles 404 Not Found errors
func (e *PocketBaseError) handleNotFoundError() string {
	msgLower := strings.ToLower(e.Message)

	if strings.Contains(msgLower, "record not found") {
		return "the requested resource was not found. It may have been deleted or you may not have access to it"
	}

	if strings.Contains(msgLower, "collection not found") {
		return "the requested collection was not found. Please check the collection name"
	}

	if strings.Contains(msgLower, "organization not found") {
		return "organization not found. Please verify the organization ID"
	}

	return "the requested resource was not found"
}

// handleValidationErrors processes validation errors from PocketBase
func (e *PocketBaseError) handleValidationErrors() string {
	if e.Data == nil {
		return "validation error occurred"
	}

	var errorMessages []string

	// Process field-specific validation errors
	for field, errors := range e.Data {
		if errorList, ok := errors.(map[string]interface{}); ok {
			if code, exists := errorList["code"].(string); exists {
				if message, exists := errorList["message"].(string); exists {
					friendlyMsg := e.getFieldValidationMessage(field, code, message)
					errorMessages = append(errorMessages, friendlyMsg)
				}
			}
		}
	}

	if len(errorMessages) > 0 {
		return fmt.Sprintf("validation failed:\n  - %s", strings.Join(errorMessages, "\n  - "))
	}

	return "input validation failed. Please check your data and try again"
}

// getFieldValidationMessage returns user-friendly validation messages for specific fields
func (e *PocketBaseError) getFieldValidationMessage(field, code, message string) string {
	// Stone-Age.io specific field mappings
	fieldDisplayNames := map[string]string{
		"email":                    "email address",
		"password":                 "password",
		"organization_id":          "organization ID",
		"current_organization_id":  "current organization",
		"name":                     "name",
		"description":              "description",
		"type":                     "type",
		"code":                     "code",
		"nats_username":           "NATS username",
		"edge_id":                 "edge ID",
		"location_id":             "location ID",
		"public_key":              "public key",
		"private_key":             "private key",
	}

	fieldDisplay := fieldDisplayNames[field]
	if fieldDisplay == "" {
		fieldDisplay = strings.Title(strings.ReplaceAll(field, "_", " "))
	}

	// Handle specific validation codes
	switch code {
	case "validation_required":
		return fmt.Sprintf("%s is required", fieldDisplay)
	case "validation_min_length":
		return fmt.Sprintf("%s is too short", fieldDisplay)
	case "validation_max_length":
		return fmt.Sprintf("%s is too long", fieldDisplay)
	case "validation_invalid_email":
		return fmt.Sprintf("%s must be a valid email address", fieldDisplay)
	case "validation_unique":
		return fmt.Sprintf("%s must be unique (already exists)", fieldDisplay)
	case "validation_invalid_format":
		return fmt.Sprintf("%s has an invalid format", fieldDisplay)
	case "validation_relation_not_found":
		return fmt.Sprintf("referenced %s was not found", fieldDisplay)
	default:
		// Use the original message if we don't have a specific handler
		return fmt.Sprintf("%s: %s", fieldDisplay, message)
	}
}

// IsAuthenticationError checks if the error is related to authentication
func (e *PocketBaseError) IsAuthenticationError() bool {
	return e.StatusCode == 401 || 
		   strings.Contains(strings.ToLower(e.Message), "auth") ||
		   strings.Contains(strings.ToLower(e.Message), "unauthorized") ||
		   strings.Contains(strings.ToLower(e.Message), "credentials")
}

// IsPermissionError checks if the error is related to permissions
func (e *PocketBaseError) IsPermissionError() bool {
	return e.StatusCode == 403 ||
		   strings.Contains(strings.ToLower(e.Message), "forbidden") ||
		   strings.Contains(strings.ToLower(e.Message), "permission") ||
		   strings.Contains(strings.ToLower(e.Message), "access denied")
}

// IsNotFoundError checks if the error is a not found error
func (e *PocketBaseError) IsNotFoundError() bool {
	return e.StatusCode == 404 ||
		   strings.Contains(strings.ToLower(e.Message), "not found")
}

// IsValidationError checks if the error is a validation error
func (e *PocketBaseError) IsValidationError() bool {
	return e.StatusCode == 400 && e.Data != nil && len(e.Data) > 0
}

// IsOrganizationError checks if the error is organization-related
func (e *PocketBaseError) IsOrganizationError() bool {
	return strings.Contains(strings.ToLower(e.Message), "organization")
}

// GetSuggestion returns a helpful suggestion based on the error type
func (e *PocketBaseError) GetSuggestion() string {
	if e.IsAuthenticationError() {
		return "try running 'flint auth pb' to authenticate with PocketBase"
	}

	if e.IsPermissionError() {
		if e.IsOrganizationError() {
			return "verify your organization membership with 'flint context show' and contact your administrator if needed"
		}
		return "contact your organization administrator to verify your permissions"
	}

	if e.IsNotFoundError() {
		return "verify the resource exists and that you have access to it"
	}

	if e.StatusCode >= 500 {
		return "this appears to be a server issue. Please try again later or contact support"
	}

	return "please check your input and try again"
}
