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

// ValidateOrganizationID validates an organization ID format
func ValidateOrganizationID(orgID string) error {
	if orgID == "" {
		return fmt.Errorf("organization ID cannot be empty")
	}

	// Stone-Age.io organization IDs should be 15 characters, alphanumeric
	if len(orgID) != 15 {
		return fmt.Errorf("organization ID must be exactly 15 characters")
	}

	validID := regexp.MustCompile(`^[a-z0-9]+$`)
	if !validID.MatchString(orgID) {
		return fmt.Errorf("organization ID can only contain lowercase letters and numbers")
	}

	return nil
}

// ValidateEmail validates an email address format
func ValidateEmail(email string) error {
	if email == "" {
		return fmt.Errorf("email cannot be empty")
	}

	// Simple email validation
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return fmt.Errorf("invalid email format")
	}

	return nil
}

// ValidateNATSSubject validates a NATS subject pattern
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
