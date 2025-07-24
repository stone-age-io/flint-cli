package pocketbase

import (
	"fmt"
	"time"
)

// Collection represents a PocketBase collection definition
type Collection struct {
	ID         string                 `json:"id"`
	Name       string                 `json:"name"`
	Type       string                 `json:"type"` // "base" or "auth"
	System     bool                   `json:"system"`
	Schema     []Field                `json:"schema"`
	ListRule   *string                `json:"listRule"`
	ViewRule   *string                `json:"viewRule"`
	CreateRule *string                `json:"createRule"`
	UpdateRule *string                `json:"updateRule"`
	DeleteRule *string                `json:"deleteRule"`
	Options    map[string]interface{} `json:"options"`
	Created    time.Time              `json:"created"`
	Updated    time.Time              `json:"updated"`
}

// Field represents a collection field definition
type Field struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Type         string                 `json:"type"`
	System       bool                   `json:"system"`
	Required     bool                   `json:"required"`
	Presentable  bool                   `json:"presentable"`
	Unique       bool                   `json:"unique,omitempty"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// RecordsList represents a paginated list of records
type RecordsList struct {
	Page       int                      `json:"page"`
	PerPage    int                      `json:"perPage"`
	TotalItems int                      `json:"totalItems"`
	TotalPages int                      `json:"totalPages"`
	Items      []map[string]interface{} `json:"items"`
}

// ListOptions represents options for listing records
type ListOptions struct {
	Page    int      `json:"page,omitempty"`
	PerPage int      `json:"perPage,omitempty"`
	Sort    string   `json:"sort,omitempty"`
	Filter  string   `json:"filter,omitempty"`
	Fields  []string `json:"fields,omitempty"`
	Expand  []string `json:"expand,omitempty"`
}

// Record represents a generic PocketBase record
type Record map[string]interface{}

// GetID returns the record ID
func (r Record) GetID() string {
	if id, ok := r["id"].(string); ok {
		return id
	}
	return ""
}

// GetString returns a string field value
func (r Record) GetString(field string) string {
	if value, ok := r[field].(string); ok {
		return value
	}
	return ""
}

// GetBool returns a boolean field value
func (r Record) GetBool(field string) bool {
	if value, ok := r[field].(bool); ok {
		return value
	}
	return false
}

// GetInt returns an integer field value
func (r Record) GetInt(field string) int {
	switch value := r[field].(type) {
	case int:
		return value
	case float64:
		return int(value)
	}
	return 0
}

// GetTime returns a time field value
func (r Record) GetTime(field string) *time.Time {
	if timeStr, ok := r[field].(string); ok {
		if t, err := time.Parse(time.RFC3339, timeStr); err == nil {
			return &t
		}
	}
	return nil
}

// GetCreated returns the record creation time
func (r Record) GetCreated() *time.Time {
	return r.GetTime("created")
}

// GetUpdated returns the record update time
func (r Record) GetUpdated() *time.Time {
	return r.GetTime("updated")
}

// OrganizationRecord represents an organization record with type safety
type OrganizationRecord struct {
	Record
}

// GetName returns the organization name
func (o OrganizationRecord) GetName() string {
	return o.GetString("name")
}

// GetAccountName returns the organization account name
func (o OrganizationRecord) GetAccountName() string {
	return o.GetString("account_name")
}

// GetCode returns the organization code
func (o OrganizationRecord) GetCode() string {
	return o.GetString("code")
}

// GetDescription returns the organization description
func (o OrganizationRecord) GetDescription() string {
	return o.GetString("description")
}

// IsActive returns whether the organization is active
func (o OrganizationRecord) IsActive() bool {
	return o.GetBool("active")
}

// EdgeRecord represents an edge device record with type safety
type EdgeRecord struct {
	Record
}

// GetName returns the edge name
func (e EdgeRecord) GetName() string {
	return e.GetString("name")
}

// GetDescription returns the edge description
func (e EdgeRecord) GetDescription() string {
	return e.GetString("description")
}

// GetType returns the edge type
func (e EdgeRecord) GetType() string {
	return e.GetString("type")
}

// GetCode returns the edge code
func (e EdgeRecord) GetCode() string {
	return e.GetString("code")
}

// GetRegion returns the edge region
func (e EdgeRecord) GetRegion() string {
	return e.GetString("region")
}

// GetOrganizationID returns the organization ID
func (e EdgeRecord) GetOrganizationID() string {
	return e.GetString("organization_id")
}

// IsActive returns whether the edge is active
func (e EdgeRecord) IsActive() bool {
	return e.GetBool("active")
}

// ThingRecord represents a thing/device record with type safety
type ThingRecord struct {
	Record
}

// GetName returns the thing name
func (t ThingRecord) GetName() string {
	return t.GetString("name")
}

// GetDescription returns the thing description
func (t ThingRecord) GetDescription() string {
	return t.GetString("description")
}

// GetType returns the thing type
func (t ThingRecord) GetType() string {
	return t.GetString("type")
}

// GetCode returns the thing code
func (t ThingRecord) GetCode() string {
	return t.GetString("code")
}

// GetOrganizationID returns the organization ID
func (t ThingRecord) GetOrganizationID() string {
	return t.GetString("organization_id")
}

// GetEdgeID returns the edge ID this thing is connected to
func (t ThingRecord) GetEdgeID() string {
	return t.GetString("edge_id")
}

// GetLocationID returns the location ID where this thing resides
func (t ThingRecord) GetLocationID() string {
	return t.GetString("location_id")
}

// UserRecord represents a user record with type safety
type UserRecord struct {
	Record
}

// GetEmail returns the user email
func (u UserRecord) GetEmail() string {
	return u.GetString("email")
}

// GetFirstName returns the user first name
func (u UserRecord) GetFirstName() string {
	return u.GetString("first_name")
}

// GetLastName returns the user last name
func (u UserRecord) GetLastName() string {
	return u.GetString("last_name")
}

// GetUsername returns the user username
func (u UserRecord) GetUsername() string {
	return u.GetString("username")
}

// GetCurrentOrganizationID returns the user's current organization ID
func (u UserRecord) GetCurrentOrganizationID() string {
	return u.GetString("current_organization_id")
}

// IsOrgAdmin returns whether the user is an organization admin
func (u UserRecord) IsOrgAdmin() bool {
	return u.GetBool("is_org_admin")
}

// IsActive returns whether the user account is active
func (u UserRecord) IsActive() bool {
	return u.GetBool("active")
}

// GetFullName returns the user's full name
func (u UserRecord) GetFullName() string {
	firstName := u.GetFirstName()
	lastName := u.GetLastName()
	
	if firstName != "" && lastName != "" {
		return firstName + " " + lastName
	} else if firstName != "" {
		return firstName
	} else if lastName != "" {
		return lastName
	} else if username := u.GetUsername(); username != "" {
		return username
	}
	
	return u.GetEmail()
}

// ClientRecord represents a NATS client record with type safety
type ClientRecord struct {
	Record
}

// GetEmail returns the client email
func (c ClientRecord) GetEmail() string {
	return c.GetString("email")
}

// GetNATSUsername returns the NATS username
func (c ClientRecord) GetNATSUsername() string {
	return c.GetString("nats_username")
}

// GetDescription returns the client description
func (c ClientRecord) GetDescription() string {
	return c.GetString("description")
}

// GetOrganizationID returns the organization ID
func (c ClientRecord) GetOrganizationID() string {
	return c.GetString("organization_id")
}

// GetJWT returns the client JWT token
func (c ClientRecord) GetJWT() string {
	return c.GetString("jwt")
}

// GetCredsFile returns the client credentials file content
func (c ClientRecord) GetCredsFile() string {
	return c.GetString("creds_file")
}

// IsBearerToken returns whether this client uses bearer token auth
func (c ClientRecord) IsBearerToken() bool {
	return c.GetBool("bearer_token")
}

// IsActive returns whether the client is active
func (c ClientRecord) IsActive() bool {
	return c.GetBool("active")
}

// LocationRecord represents a location record with type safety
type LocationRecord struct {
	Record
}

// GetName returns the location name
func (l LocationRecord) GetName() string {
	return l.GetString("name")
}

// GetType returns the location type
func (l LocationRecord) GetType() string {
	return l.GetString("type")
}

// GetCode returns the location code
func (l LocationRecord) GetCode() string {
	return l.GetString("code")
}

// GetPath returns the location path
func (l LocationRecord) GetPath() string {
	return l.GetString("path")
}

// GetOrganizationID returns the organization ID
func (l LocationRecord) GetOrganizationID() string {
	return l.GetString("organization_id")
}

// GetEdgeID returns the edge ID
func (l LocationRecord) GetEdgeID() string {
	return l.GetString("edge_id")
}

// GetParentID returns the parent location ID
func (l LocationRecord) GetParentID() string {
	return l.GetString("parent_id")
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error implements the error interface
func (v ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", v.Field, v.Message)
}

// RequestOptions represents common request options
type RequestOptions struct {
	Headers map[string]string `json:"headers,omitempty"`
	Timeout time.Duration     `json:"timeout,omitempty"`
}

// HealthStatus represents the PocketBase health status
type HealthStatus struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		CanBackup bool `json:"canBackup"`
	} `json:"data"`
}
