package utils

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"flint-cli/internal/config"
)

// ValidateURL validates that a string is a valid URL
func ValidateURL(urlStr string) error {
	if urlStr == "" {
		return fmt.Errorf("URL cannot be empty")
	}

	_, err := url.Parse(urlStr)
	if err != nil {
		return fmt.Errorf("invalid URL format: %w", err)
	}

	return nil
}

// ValidateContextName validates a context name
func ValidateContextName(name string) error {
	if name == "" {
		return fmt.Errorf("context name cannot be empty")
	}

	// Check for valid characters (alphanumeric, dash, underscore)
	validName := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !validName.MatchString(name) {
		return fmt.Errorf("context name can only contain letters, numbers, hyphens, and underscores")
	}

	// Check length
	if len(name) > 50 {
		return fmt.Errorf("context name must be 50 characters or less")
	}

	return nil
}

// ValidateAuthCollection validates a PocketBase auth collection name
func ValidateAuthCollection(collection string) error {
	return config.ValidateAuthCollection(collection)
}

// ValidateNATSAuthMethod validates a NATS authentication method
func ValidateNATSAuthMethod(method string) error {
	validMethods := []string{
		config.NATSAuthUserPass,
		config.NATSAuthToken,
		config.NATSAuthCreds,
	}

	for _, valid := range validMethods {
		if method == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid NATS auth method '%s'. Valid options: %s", 
		method, strings.Join(validMethods, ", "))
}

// ValidateOutputFormat validates an output format
func ValidateOutputFormat(format string) error {
	validFormats := []string{
		config.OutputFormatJSON,
		config.OutputFormatYAML,
		config.OutputFormatTable,
	}

	format = strings.ToLower(format)
	for _, valid := range validFormats {
		if format == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid output format '%s'. Valid options: %s", 
		format, strings.Join(validFormats, ", "))
}

// ValidateCollectionName validates a Stone-Age.io collection name
func ValidateCollectionName(collection string) error {
	validCollections := config.GetDefaultCollections()
	
	for _, valid := range validCollections {
		if collection == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid collection '%s'. Valid collections: %s", 
		collection, strings.Join(validCollections, ", "))
}

// ValidateOrganizationID validates an organization ID format (minimal - PocketBase handles detailed validation)
func ValidateOrganizationID(orgID string) error {
	if orgID == "" {
		return fmt.Errorf("organization ID cannot be empty")
	}
	return nil
}

// ValidateEmail validates an email address format (minimal - PocketBase handles detailed validation)
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Basic check for @ symbol
	if !strings.Contains(email, "@") {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateNATSSubject validates a NATS subject pattern (basic validation only)
func ValidateNATSSubject(subject string) error {
	if subject == "" {
		return fmt.Errorf("NATS subject cannot be empty")
	}

	// NATS subjects can contain letters, numbers, dots, and wildcards
	validSubject := regexp.MustCompile(`^[a-zA-Z0-9.*>_-]+$`)
	if !validSubject.MatchString(subject) {
		return fmt.Errorf("invalid NATS subject format. Use letters, numbers, dots, and wildcards (*, >)")
	}

	return nil
}

// ValidateNATSServers validates NATS server URLs
func ValidateNATSServers(servers []string) error {
	if len(servers) == 0 {
		return fmt.Errorf("at least one NATS server is required")
	}

	for i, server := range servers {
		if err := ValidateNATSServerURL(server); err != nil {
			return fmt.Errorf("invalid NATS server %d: %w", i+1, err)
		}
	}

	return nil
}

// ValidateNATSServerURL validates a single NATS server URL
func ValidateNATSServerURL(serverURL string) error {
	if serverURL == "" {
		return fmt.Errorf("NATS server URL cannot be empty")
	}

	// Parse the URL to validate format
	if err := ValidateURL(serverURL); err != nil {
		return fmt.Errorf("invalid NATS server URL format: %w", err)
	}

	// Check for supported NATS schemes
	if !strings.HasPrefix(serverURL, "nats://") && 
	   !strings.HasPrefix(serverURL, "tls://") &&
	   !strings.HasPrefix(serverURL, "ws://") &&
	   !strings.HasPrefix(serverURL, "wss://") {
		return fmt.Errorf("NATS server URL must use nats://, tls://, ws://, or wss:// scheme")
	}

	return nil
}

// ValidateNATSMessage validates message data and headers
func ValidateNATSMessage(data []byte, headers map[string]string) error {
	// Check message size (NATS has a default max of 1MB)
	maxMessageSize := 1024 * 1024 // 1MB
	if len(data) > maxMessageSize {
		return fmt.Errorf("message size %d bytes exceeds maximum allowed size of %d bytes", 
			len(data), maxMessageSize)
	}

	// Validate headers if present
	if headers != nil {
		for key, value := range headers {
			if err := ValidateNATSHeaderField(key, value); err != nil {
				return fmt.Errorf("invalid header '%s': %w", key, err)
			}
		}
	}

	return nil
}

// ValidateNATSHeaderField validates a NATS message header field
func ValidateNATSHeaderField(key, value string) error {
	if key == "" {
		return fmt.Errorf("header key cannot be empty")
	}

	// NATS header keys should be valid HTTP header names
	validHeaderKey := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9\-_]*$`)
	if !validHeaderKey.MatchString(key) {
		return fmt.Errorf("header key must contain only letters, numbers, hyphens, and underscores")
	}

	// Check for reasonable header value length
	if len(value) > 4096 { // 4KB limit per header value
		return fmt.Errorf("header value exceeds maximum length of 4096 characters")
	}

	return nil
}

// ValidateNATSQueue validates a NATS queue group name
func ValidateNATSQueue(queue string) error {
	if queue == "" {
		return nil // Queue is optional
	}

	// Queue names should be simple identifiers
	validQueue := regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_-]*$`)
	if !validQueue.MatchString(queue) {
		return fmt.Errorf("queue name must start with letter/number and contain only letters, numbers, hyphens, and underscores")
	}

	if len(queue) > 64 {
		return fmt.Errorf("queue name must be 64 characters or less")
	}

	return nil
}

// ValidateNATSCredentialsFile validates that a credentials file path is reasonable
func ValidateNATSCredentialsFile(credsPath string) error {
	if credsPath == "" {
		return fmt.Errorf("credentials file path cannot be empty")
	}

	// Basic path validation (file existence will be checked by NATS client)
	if err := ValidateFileExists(credsPath); err != nil {
		return fmt.Errorf("invalid credentials file path: %w", err)
	}

	// Check file extension
	if !strings.HasSuffix(credsPath, ".creds") {
		return fmt.Errorf("credentials file should have .creds extension")
	}

	return nil
}

// ValidateRequiredString validates that a string is not empty
func ValidateRequiredString(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// ValidateStringLength validates string length constraints
func ValidateStringLength(value, fieldName string, min, max int) error {
	length := len(value)
	if length < min {
		return fmt.Errorf("%s must be at least %d characters", fieldName, min)
	}
	if max > 0 && length > max {
		return fmt.Errorf("%s must be no more than %d characters", fieldName, max)
	}
	return nil
}

// ValidateFileExists checks if a file exists at the given path
func ValidateFileExists(path string) error {
	if path == "" {
		return fmt.Errorf("file path cannot be empty")
	}

	// We'll implement actual file checking in Phase 5 when we add file operations
	// For now, just validate the path format
	if strings.Contains(path, "..") {
		return fmt.Errorf("file path cannot contain '..' for security reasons")
	}

	return nil
}
