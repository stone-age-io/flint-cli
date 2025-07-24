package config

import (
	"fmt"
	"strings"
	"time"
)

// GlobalConfig represents the global CLI configuration
type GlobalConfig struct {
	ActiveContext       string `yaml:"active_context"`
	OutputFormat        string `yaml:"output_format"`        // json|yaml|table
	ColorsEnabled       bool   `yaml:"colors_enabled"`
	PaginationSize      int    `yaml:"pagination_size"`
	OrganizationDisplay bool   `yaml:"organization_display"`
	Debug               bool   `yaml:"debug"`
}

// Context represents a single environment context configuration
type Context struct {
	Name       string             `yaml:"name"`
	PocketBase PocketBaseConfig   `yaml:"pocketbase"`
	NATS       NATSConfig         `yaml:"nats"`
}

// PocketBaseConfig contains PocketBase-specific configuration
type PocketBaseConfig struct {
	URL                    string                 `yaml:"url"`
	AuthCollection         string                 `yaml:"auth_collection"`         // users|clients|edges|things|service_users
	OrganizationID         string                 `yaml:"organization_id"`         // Set after authentication
	AvailableCollections   []string               `yaml:"available_collections"`
	AuthToken              string                 `yaml:"auth_token"`              // Session token
	AuthExpires            *time.Time             `yaml:"auth_expires"`            // Token expiration
	AuthRecord             map[string]interface{} `yaml:"auth_record"`             // Cached auth record
}

// NATSConfig contains NATS-specific configuration
type NATSConfig struct {
	Servers     []string `yaml:"servers"`
	AuthMethod  string   `yaml:"auth_method"`  // user_pass|token|creds
	Username    string   `yaml:"username"`     // For user_pass method
	Password    string   `yaml:"password"`     // For user_pass method
	Token       string   `yaml:"token"`        // For token method
	CredsFile   string   `yaml:"creds_file"`   // For creds method
	TLSEnabled  bool     `yaml:"tls_enabled"`
	TLSVerify   bool     `yaml:"tls_verify"`
}

// StoneAgeCollections defines the available Stone-Age.io collections
type StoneAgeCollections struct {
	// Core entities
	Organizations string
	Users         string
	Edges         string
	Things        string
	Locations     string
	Clients       string

	// Type definitions
	EdgeTypes     string
	ThingTypes    string
	LocationTypes string
	EdgeRegions   string

	// System
	AuditLogs           string
	TopicPermissions    string
	NATSPublishQueue    string
	NATSSystemOperator  string
}

// GetStoneAgeCollections returns the collection definitions
func GetStoneAgeCollections() StoneAgeCollections {
	return StoneAgeCollections{
		// Core entities
		Organizations: "organizations",
		Users:         "users",
		Edges:         "edges",
		Things:        "things",
		Locations:     "locations",
		Clients:       "clients",

		// Type definitions
		EdgeTypes:     "edge_types",
		ThingTypes:    "thing_types",
		LocationTypes: "location_types",
		EdgeRegions:   "edge_regions",

		// System
		AuditLogs:           "audit_logs",
		TopicPermissions:    "topic_permissions",
		NATSPublishQueue:    "nats_publish_queue",
		NATSSystemOperator:  "nats_system_operator",
	}
}

// GetDefaultCollections returns the default list of available collections
func GetDefaultCollections() []string {
	collections := GetStoneAgeCollections()
	return []string{
		collections.Organizations,
		collections.Users,
		collections.Edges,
		collections.Things,
		collections.Locations,
		collections.Clients,
		collections.EdgeTypes,
		collections.ThingTypes,
		collections.LocationTypes,
		collections.EdgeRegions,
		collections.AuditLogs,
		collections.TopicPermissions,
	}
}

// NATSAuthMethod constants
const (
	NATSAuthUserPass = "user_pass"
	NATSAuthToken    = "token"
	NATSAuthCreds    = "creds"
)

// Output format constants
const (
	OutputFormatJSON  = "json"
	OutputFormatYAML  = "yaml"
	OutputFormatTable = "table"
)

// PocketBase auth collection constants
const (
	AuthCollectionUsers        = "users"
	AuthCollectionClients      = "clients"
	AuthCollectionEdges        = "edges"
	AuthCollectionThings       = "things"
	AuthCollectionServiceUsers = "service_users"
)

// ValidateAuthCollection validates a PocketBase auth collection name
func ValidateAuthCollection(collection string) error {
	validCollections := []string{
		AuthCollectionUsers,
		AuthCollectionClients,
		AuthCollectionEdges,
		AuthCollectionThings,
		AuthCollectionServiceUsers,
	}

	for _, valid := range validCollections {
		if collection == valid {
			return nil
		}
	}

	return fmt.Errorf("invalid auth collection '%s'. Valid options: %s", 
		collection, strings.Join(validCollections, ", "))
}

// Global configuration instance (will be populated by root command)
var Global = &GlobalConfig{
	OutputFormat:        OutputFormatJSON,
	ColorsEnabled:       true,
	PaginationSize:      30,
	OrganizationDisplay: true,
	Debug:               false,
}
