package nats

import (
	"fmt"
	"strings"

	"github.com/nats-io/nats.go"
)

// NATSError represents a NATS-specific error with user-friendly messaging
type NATSError struct {
	Operation string
	Subject   string
	Err       error
}

// Error implements the error interface
func (e *NATSError) Error() string {
	return e.GetFriendlyMessage()
}

// NewNATSError creates a new NATSError
func NewNATSError(operation, subject string, err error) *NATSError {
	return &NATSError{
		Operation: operation,
		Subject:   subject,
		Err:       err,
	}
}

// GetFriendlyMessage returns a user-friendly error message for Stone-Age.io NATS operations
func (e *NATSError) GetFriendlyMessage() string {
	if e.Err == nil {
		return fmt.Sprintf("unknown error during %s operation", e.Operation)
	}

	errMsg := strings.ToLower(e.Err.Error())

	// Connection-related errors
	if strings.Contains(errMsg, "no servers available") {
		return e.formatConnectionError("no NATS servers are currently available", 
			"check your network connection and verify the NATS server URLs in your context")
	}

	if strings.Contains(errMsg, "connection refused") {
		return e.formatConnectionError("connection to NATS server was refused", 
			"verify the NATS server is running and the URL/port are correct")
	}

	if strings.Contains(errMsg, "timeout") {
		if e.Operation == "connect" {
			return e.formatConnectionError("connection to NATS server timed out", 
				"check your network connection and server availability")
		}
		return e.formatOperationError("operation timed out", 
			"the NATS server may be overloaded or your network connection is slow")
	}

	if strings.Contains(errMsg, "connection closed") {
		return e.formatConnectionError("NATS connection was closed unexpectedly", 
			"check your network connection and NATS server status")
	}

	// Authentication-related errors
	if strings.Contains(errMsg, "authorization violation") || 
	   strings.Contains(errMsg, "authentication failed") {
		return e.formatAuthError("NATS authentication failed", 
			"verify your credentials are correct and check your context configuration")
	}

	if strings.Contains(errMsg, "user credentials") {
		return e.formatAuthError("invalid NATS user credentials", 
			"check your credentials file path and ensure the file is readable")
	}

	if strings.Contains(errMsg, "nkey") || strings.Contains(errMsg, "jwt") {
		return e.formatAuthError("NATS JWT/NKey authentication failed", 
			"verify your credentials file is valid and not expired")
	}

	if strings.Contains(errMsg, "permission") {
		return e.formatPermissionError("insufficient permissions for NATS operation", 
			"contact your Stone-Age.io administrator to verify your topic permissions")
	}

	// Subject/topic related errors
	if strings.Contains(errMsg, "invalid subject") {
		return e.formatSubjectError("invalid NATS subject format", 
			"ensure the subject follows NATS naming conventions (e.g., 'telemetry.edge.123')")
	}

	if strings.Contains(errMsg, "subscription not found") {
		return e.formatOperationError("subscription not found", 
			"the subscription may have been closed or never existed")
	}

	// Message-related errors
	if strings.Contains(errMsg, "maximum payload") || strings.Contains(errMsg, "message too large") {
		return e.formatOperationError("message exceeds maximum size limit", 
			"reduce your message size or contact your administrator about server limits")
	}

	if strings.Contains(errMsg, "slow consumer") {
		return e.formatOperationError("message consumer is too slow", 
			"the subscription cannot keep up with incoming messages")
	}

	// TLS-related errors
	if strings.Contains(errMsg, "tls") || strings.Contains(errMsg, "certificate") {
		return e.formatConnectionError("TLS connection failed", 
			"check your TLS configuration and certificate validity")
	}

	// Server-specific errors
	if strings.Contains(errMsg, "server shutdown") {
		return e.formatConnectionError("NATS server is shutting down", 
			"wait for the server to restart or connect to a different server")
	}

	if strings.Contains(errMsg, "cluster") {
		return e.formatConnectionError("NATS cluster communication issue", 
			"this may be a temporary issue with the NATS cluster")
	}

	// Generic error with context
	return e.formatGenericError()
}

// formatConnectionError formats connection-related error messages
func (e *NATSError) formatConnectionError(problem, suggestion string) string {
	return fmt.Sprintf("NATS connection error: %s.\nSuggestion: %s", problem, suggestion)
}

// formatAuthError formats authentication-related error messages
func (e *NATSError) formatAuthError(problem, suggestion string) string {
	return fmt.Sprintf("NATS authentication error: %s.\nSuggestion: %s", problem, suggestion)
}

// formatPermissionError formats permission-related error messages
func (e *NATSError) formatPermissionError(problem, suggestion string) string {
	subjectInfo := ""
	if e.Subject != "" {
		subjectInfo = fmt.Sprintf(" for subject '%s'", e.Subject)
	}
	return fmt.Sprintf("NATS permission error: %s%s.\nSuggestion: %s", problem, subjectInfo, suggestion)
}

// formatSubjectError formats subject-related error messages
func (e *NATSError) formatSubjectError(problem, suggestion string) string {
	subjectInfo := ""
	if e.Subject != "" {
		subjectInfo = fmt.Sprintf(" (subject: '%s')", e.Subject)
	}
	return fmt.Sprintf("NATS subject error: %s%s.\nSuggestion: %s", problem, subjectInfo, suggestion)
}

// formatOperationError formats operation-specific error messages
func (e *NATSError) formatOperationError(problem, suggestion string) string {
	return fmt.Sprintf("NATS %s error: %s.\nSuggestion: %s", e.Operation, problem, suggestion)
}

// formatGenericError formats generic errors with operation context
func (e *NATSError) formatGenericError() string {
	operationContext := ""
	if e.Operation != "" {
		operationContext = fmt.Sprintf(" during %s operation", e.Operation)
	}
	
	subjectContext := ""
	if e.Subject != "" {
		subjectContext = fmt.Sprintf(" (subject: %s)", e.Subject)
	}

	return fmt.Sprintf("NATS error%s%s: %v.\nSuggestion: check your connection and try again, or contact support if the issue persists", 
		operationContext, subjectContext, e.Err)
}

// IsConnectionError checks if the error is connection-related
func (e *NATSError) IsConnectionError() bool {
	if e.Err == nil {
		return false
	}
	
	errMsg := strings.ToLower(e.Err.Error())
	return strings.Contains(errMsg, "no servers available") ||
		   strings.Contains(errMsg, "connection refused") ||
		   strings.Contains(errMsg, "connection closed") ||
		   strings.Contains(errMsg, "connection timeout") ||
		   strings.Contains(errMsg, "server shutdown")
}

// IsAuthError checks if the error is authentication-related
func (e *NATSError) IsAuthError() bool {
	if e.Err == nil {
		return false
	}
	
	errMsg := strings.ToLower(e.Err.Error())
	return strings.Contains(errMsg, "authorization") ||
		   strings.Contains(errMsg, "authentication") ||
		   strings.Contains(errMsg, "credentials") ||
		   strings.Contains(errMsg, "nkey") ||
		   strings.Contains(errMsg, "jwt")
}

// IsPermissionError checks if the error is permission-related
func (e *NATSError) IsPermissionError() bool {
	if e.Err == nil {
		return false
	}
	
	errMsg := strings.ToLower(e.Err.Error())
	return strings.Contains(errMsg, "permission") ||
		   strings.Contains(errMsg, "not authorized")
}

// IsTemporaryError checks if the error might be temporary
func (e *NATSError) IsTemporaryError() bool {
	if e.Err == nil {
		return false
	}
	
	errMsg := strings.ToLower(e.Err.Error())
	return strings.Contains(errMsg, "timeout") ||
		   strings.Contains(errMsg, "slow consumer") ||
		   strings.Contains(errMsg, "server shutdown") ||
		   strings.Contains(errMsg, "cluster")
}

// GetRecoveryAction provides specific recovery actions based on error type
func (e *NATSError) GetRecoveryAction() string {
	if e.IsConnectionError() {
		return "try running 'flint context show' to verify your NATS server configuration"
	}
	
	if e.IsAuthError() {
		return "run 'flint auth nats' to reconfigure your NATS authentication"
	}
	
	if e.IsPermissionError() {
		return "contact your Stone-Age.io administrator to verify your topic permissions"
	}
	
	if e.IsTemporaryError() {
		return "wait a moment and try the operation again"
	}
	
	return "check your configuration with 'flint context show' and try again"
}

// WrapNATSError wraps a standard NATS error with enhanced messaging
func WrapNATSError(operation, subject string, err error) error {
	if err == nil {
		return nil
	}
	
	// If it's already a NATSError, return as-is
	if _, ok := err.(*NATSError); ok {
		return err
	}
	
	return NewNATSError(operation, subject, err)
}

// TranslateNATSError converts common NATS library errors to user-friendly messages
func TranslateNATSError(err error) string {
	if err == nil {
		return ""
	}

	// Handle specific NATS error types that exist in current version
	if err == nats.ErrConnectionClosed {
		return "NATS connection was closed. Please reconnect."
	}
	
	if err == nats.ErrConnectionDraining {
		return "NATS connection is draining. Please wait or reconnect."
	}
	
	if err == nats.ErrInvalidConnection {
		return "NATS connection is invalid. Please reconnect."
	}
	
	if err == nats.ErrInvalidMsg {
		return "invalid NATS message format"
	}
	
	if err == nats.ErrTimeout {
		return "NATS operation timed out"
	}
	
	if err == nats.ErrNoServers {
		return "no NATS servers available"
	}
	
	if err == nats.ErrMaxPayload {
		return "message exceeds maximum payload size"
	}
	
	// Check for specific error strings that might not have constants
	errStr := err.Error()
	if strings.Contains(strings.ToLower(errStr), "invalid subject") {
		return "invalid NATS subject format"
	}
	
	if strings.Contains(strings.ToLower(errStr), "auth expired") {
		return "NATS authentication has expired. Please re-authenticate."
	}
	
	if strings.Contains(strings.ToLower(errStr), "auth revoked") {
		return "NATS authentication has been revoked. Please re-authenticate."
	}
	
	// Return the original error message for unhandled cases
	return err.Error()
}
