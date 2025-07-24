# Flint CLI

A command-line interface for managing the Stone-Age.io IoT platform. Flint provides comprehensive tools for context management, PocketBase authentication, collection operations, and NATS messaging.

## Phase 1 Implementation Status ✅

**Completed Features:**
- ✅ Go module initialization with dependencies
- ✅ Cobra CLI framework with root command
- ✅ XDG-compliant configuration directory management
- ✅ Context management system with organization support
- ✅ YAML configuration file handling
- ✅ Partial command matching resolver (Cisco-style)
- ✅ Complete context subcommands (create, list, select, show, delete, organization)
- ✅ Utility functions for output formatting and validation
- ✅ Stone-Age.io specific collection definitions

## Phase 2 Implementation Status ✅

**Completed Features:**
- ✅ PocketBase HTTP client with authentication
- ✅ Support for multiple auth collections (users, clients, edges, things, service_users)
- ✅ Persistent session token management
- ✅ Organization-aware operations with validation
- ✅ Comprehensive error handling with friendly messages
- ✅ Stone-Age.io specific error translations
- ✅ Interactive and non-interactive authentication modes
- ✅ Automatic organization selection and validation
- ✅ Session expiration handling and recovery

## Phase 3 Implementation Status ✅

**Completed Features:**
- ✅ Complete CRUD operations for all Stone-Age.io collections
- ✅ Collection validation against context's available collections
- ✅ Action partial matching (list, get, create, update, delete)
- ✅ JSON input support with file loading capability
- ✅ Organization-scoped operations (automatic via PocketBase rules)
- ✅ Comprehensive validation for create/update operations
- ✅ Collection-specific field validation and warnings
- ✅ Multiple output formats with collection-aware table display
- ✅ Pagination support with transparent offset/limit handling
- ✅ Rich error handling with actionable suggestions

## Recent Polish Improvements ✅

**Completed Improvements:**
- ✅ Fixed collections command help display (now shows usage instead of full help)
- ✅ Standardized error message formatting across the entire codebase
- ✅ Consistent error message style (lowercase start, no periods, proper wrapping)
- ✅ Enhanced user experience with cleaner error messages
- ✅ Improved CLI behavior to match standard conventions

## Installation

```bash
# Clone the repository
git clone https://github.com/yourusername/flint-cli.git
cd flint-cli

# Install dependencies
go mod tidy

# Build the binary
go build -o flint main.go

# Optional: Install globally
sudo mv flint /usr/local/bin/
```

## Quick Start

### 1. Create Your First Context

```bash
# Create a production context
flint context create production \
  --pb-url https://api.stone-age.io \
  --pb-auth-collection users \
  --nats-servers nats://nats1.stone-age.io:4222,nats://nats2.stone-age.io:4222 \
  --nats-auth-method creds

# Create a development context
flint context create development \
  --pb-url http://localhost:8090 \
  --nats-servers nats://localhost:4222 \
  --nats-auth-method user_pass
```

### 2. Select Active Context

```bash
# List all contexts
flint context list

# Select the production context
flint context select production

# View current context details
flint context show
```

### 3. Authenticate with PocketBase

```bash
# Interactive authentication (prompts for credentials)
flint auth pb

# Authenticate with specific credentials
flint auth pb --email admin@company.com --password secret

# Authenticate as different entity types
flint auth pb --collection edges --email edge001@company.com
flint auth pb --collection clients --email client001@company.com
```

### 4. Set Organization

```bash
# Set your organization ID (required for API operations)
flint context organization org_abc123def456789

# Or set during authentication
flint auth pb --email admin@company.com --organization org_abc123def456789
```

### 5. Manage Collections

```bash
# List all edges in your organization
flint collections edges list

# Get a specific user by ID
flint collections users get user_abc123def456

# Create a new edge device
flint collections edges create '{"name":"Building-A-Edge","code":"bldg-a","type":"controller","region":"west"}'

# Update an existing thing
flint collections things update thing_123456789 '{"name":"Updated Thing Name","active":true}'

# Delete a location (with confirmation)
flint collections locations delete loc_abc123def456
```

## Configuration

Flint uses XDG-compliant configuration directories:

```
~/.config/flint/
├── config.yaml           # Global configuration
└── contexts/             # Context configurations
    ├── production.yaml
    ├── development.yaml
    └── staging.yaml
```

### Global Configuration (`~/.config/flint/config.yaml`)

```yaml
active_context: production
output_format: json        # json|yaml|table
colors_enabled: true
pagination_size: 30
organization_display: true
debug: false
```

### Context Configuration (`~/.config/flint/contexts/production.yaml`)

```yaml
name: production
pocketbase:
  url: https://api.stone-age.io
  auth_collection: users
  organization_id: org_abc123def456789
  available_collections:
    - organizations
    - users
    - edges
    - things
    - locations
    - clients
    # ... more collections
  auth_token: ""           # Managed by CLI
  auth_expires: null       # Managed by CLI
  auth_record: {}          # Managed by CLI
nats:
  servers:
    - nats://nats1.stone-age.io:4222
    - nats://nats2.stone-age.io:4222
  auth_method: creds       # user_pass|token|creds
  creds_file: ~/.config/flint/nats/production.creds
  username: ""
  password: ""
  token: ""
  tls_enabled: true
  tls_verify: true
```

## Commands

### Context Management

```bash
# Create new context
flint context create <name> [flags]

# List all contexts
flint context list

# Select active context
flint context select <name>

# Show context details
flint context show [name]

# Delete context
flint context delete <name>

# Set organization for current context
flint context organization <org_id>
```

### Authentication

```bash
# PocketBase authentication
flint auth pb [flags]
  --email string         Email address for authentication
  --password string      Password for authentication
  --collection string    Authentication collection (users|clients|edges|things|service_users)
  --organization string  Organization ID to set after authentication
```

### Collections Operations

```bash
# List records from a collection
flint collections <collection> list [flags]
  --offset int           Number of records to skip (default: 0)
  --limit int            Maximum number of records to return (default: 30)
  --filter string        PocketBase filter expression
  --sort string          Sort expression (e.g., 'name', '-created')
  --fields strings       Specific fields to return (comma-separated)
  --expand strings       Relations to expand (comma-separated)
  --output string        Output format (json|yaml|table)

# Get a single record by ID
flint collections <collection> get <record_id> [flags]
  --expand strings       Relations to expand (comma-separated)
  --output string        Output format (json|yaml|table)

# Create a new record
flint collections <collection> create [json_data] [flags]
  --file string          Path to JSON file containing record data
  --output string        Output format (json|yaml|table)

# Update an existing record
flint collections <collection> update <record_id> [json_data] [flags]
  --file string          Path to JSON file containing update data
  --output string        Output format (json|yaml|table)

# Delete a record
flint collections <collection> delete <record_id> [flags]
  --force                Skip confirmation prompt
  --quiet                Suppress success messages
```

### Partial Command Matching

Flint supports Cisco-style partial command matching:

```bash
# These are all equivalent:
flint collections edges list
flint col edges list
flint collections edges ls
flint col edg l

# But collection names must be exact:
flint collections edges list    # ✓ Correct
flint collections edg list      # ✗ Invalid - collection name must be exact

# Ambiguous commands will show suggestions:
flint collections edges c
# Error: ambiguous command 'c'. Possible matches: create

# Unknown commands show available options:
flint collections edges xyz
# Error: unknown command 'xyz'. Available commands: create, delete, get, list, update
```

### Global Flags

```bash
--output, -o     Output format (json|yaml|table) [default: json]
--colors         Enable colored output [default: true]
--debug          Enable debug output [default: false]
```

## Stone-Age.io Collections

Flint supports all Stone-Age.io collections with full CRUD operations:

### Core Collections

**Available Collections:**
- `organizations` - Multi-tenant organization management
- `users` - Human administrators with organization membership
- `edges` - Edge computing nodes managing local IoT devices
- `things` - IoT devices (door controllers, sensors, etc.)
- `locations` - Physical locations with hierarchical structure
- `clients` - NATS client authentication entities

**Type Collections:**
- `edge_types` - Edge device type definitions
- `thing_types` - IoT device type definitions
- `location_types` - Location type definitions
- `edge_regions` - Geographic/logical edge groupings

**System Collections:**
- `audit_logs` - System audit trail (read-only)
- `topic_permissions` - NATS topic access control

### Authentication Collections

**Supported Auth Collections:**
- `users` (default) - Human administrators with organization membership
- `clients` - NATS client authentication entities  
- `edges` - Edge device authentication
- `things` - Individual IoT device authentication
- `service_users` - System service accounts

## Examples

### Collection Management Workflows

```bash
# Organization setup
flint collections organizations create '{"name":"ACME Corp","code":"acme","account_name":"acme-corp"}'
flint collections organizations list --filter 'active=true' --output table

# User management
flint collections users create '{"email":"john@acme.com","password":"secure123","first_name":"John","last_name":"Doe"}'
flint collections users list --fields email,first_name,last_name,is_org_admin
flint collections users update user_123456789 '{"is_org_admin":true}'

# Edge device management
flint collections edges create '{"name":"Building-A-Gateway","code":"bldg-a-gw","type":"gateway","region":"us-west"}'
flint collections edges list --filter 'region="us-west" && active=true' --sort name
flint collections edges get edge_abc123def456 --expand organization_id,edge_types

# IoT device management
flint collections things create '{"name":"Door-Controller-01","code":"door-01","type":"access_control","edge_id":"edge_abc123def456"}'
flint collections things list --filter 'edge_id="edge_abc123def456"' --expand edge_id
flint collections things update thing_123456789 '{"location_id":"loc_conference_room_a"}'

# Location hierarchy
flint collections locations create '{"name":"Building A","type":"building","code":"bldg-a"}'
flint collections locations create '{"name":"Floor 1","type":"floor","code":"bldg-a-f1","parent_id":"loc_building_a"}'
flint collections locations list --filter 'type="room"' --sort path

# Bulk operations with files
echo '{"name":"Test Edge","code":"test-edge","type":"controller","region":"test"}' > edge.json
flint collections edges create --file edge.json

# Advanced filtering and pagination
flint collections edges list --filter 'active=true && region~"us-"' --sort '-created' --limit 10
flint collections users list --offset 30 --limit 10 --fields email,first_name,last_name
```

### JSON File Examples

**Create Edge (edge.json):**
```json
{
  "name": "Production Gateway",
  "code": "prod-gw-01",
  "type": "gateway",
  "region": "us-east",
  "description": "Primary gateway for east coast operations"
}
```

**Update User (user-update.json):**
```json
{
  "first_name": "Jane",
  "last_name": "Smith",
  "is_org_admin": true,
  "active": true
}
```

**Create Location (location.json):**
```json
{
  "name": "Conference Room A",
  "type": "room",
  "code": "conf-room-a",
  "description": "Main conference room with AV equipment",
  "parent_id": "loc_floor_1"
}
```

## Error Handling

Flint provides comprehensive error handling with user-friendly messages:

### Collection Errors
```bash
# Invalid collection
$ flint collections invalid_collection list
Error: collection 'invalid_collection' not available in current context. Available collections: organizations, users, edges, things, locations, clients, edge_types, thing_types, location_types, edge_regions, audit_logs, topic_permissions

# Record not found
$ flint collections users get invalid_id
Error: the requested resource was not found. It may have been deleted or you may not have access to it.
Suggestion: verify the resource exists and that you have access to it.
```

### Authentication Errors
```bash
# Session expired
$ flint collections edges list
Error: your session has expired. Please authenticate again using 'flint auth pb'.

# Organization access denied
$ flint collections organizations list
Error: you don't have permission to access resources in this organization. Please verify your organization membership or contact your administrator.
```

### Validation Errors
```bash
# Invalid JSON
$ flint collections edges create 'invalid json'
Error: invalid JSON format: invalid character 'i' looking for beginning of value

# Missing required fields
$ flint collections users create '{}'
Error: validation failed:
  - email address is required
  - password is required
```

## Architecture

```
flint/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command setup
│   ├── context/           # Context management commands
│   ├── auth/              # Authentication commands
│   └── collections/       # Collection CRUD operations
├── internal/
│   ├── config/            # Context and configuration management
│   ├── pocketbase/        # PocketBase client and operations
│   ├── resolver/          # Partial command matching logic  
│   └── utils/             # Shared utilities
└── main.go               # Application entry point
```

## Development Status

### Phase 1 Completion Checklist ✅

- [x] Context management with organization support
- [x] Configuration files properly managed
- [x] Partial command matching working
- [x] Organization selection functionality
- [x] XDG-compliant directory structure
- [x] Comprehensive error handling
- [x] Stone-Age.io collection definitions
- [x] Utility functions for output and validation

### Phase 2 Completion Checklist ✅

- [x] PocketBase HTTP client implementation
- [x] Authentication system for multiple collections
- [x] Session token management with expiration
- [x] Organization validation via PocketBase API
- [x] Error translation from PocketBase responses
- [x] Interactive and non-interactive authentication
- [x] Comprehensive error handling with suggestions
- [x] Integration with existing context system

### Phase 3 Completion Checklist ✅

- [x] CRUD operations for all Stone-Age.io collections
- [x] Organization-scoped operations working
- [x] Collection validation against context configuration
- [x] Action partial matching (list, get, create, update, delete)
- [x] JSON input support with file loading
- [x] Basic JSON validation before PocketBase calls
- [x] Filtering, pagination, and multiple output formats
- [x] Collection-specific validation and field handling
- [x] Rich table output with collection-aware formatting
- [x] Comprehensive error handling with actionable suggestions
- [x] Delete confirmations with record details and warnings

### Polish Improvements Checklist ✅

- [x] Fixed collections command help display (shows usage instead of full help)
- [x] Standardized error message formatting across codebase
- [x] Consistent error message style (lowercase, no periods, proper format)
- [x] Enhanced user experience with cleaner error messages
- [x] Improved CLI behavior following standard conventions

### Next Steps (Phase 4) - NATS Integration 

- [ ] NATS client integration with multiple auth methods
- [ ] NATS publish/subscribe operations  
- [ ] Real-time message streaming to stdout
- [ ] Connection management and error handling
- [ ] Integration with Stone-Age.io topic structure

### Phase 5 (Future) - File Operations & Final Polish

- [ ] Generic file operations for any collection field
- [ ] File upload/download with progress indicators
- [ ] Comprehensive testing suite
- [ ] Complete documentation and usage examples
- [ ] Automated builds and GitHub Actions
- [ ] Version information and update checking

## Contributing

1. Follow Go best practices and maintain DRY principles
2. Add comprehensive comments for maintainability
3. Break complex tasks into smaller, manageable functions
4. Use the established directory structure
5. Test all partial command matching scenarios
6. Ensure proper error handling with user-friendly messages
7. Validate all JSON inputs and provide helpful error messages
8. Maintain consistent error message formatting

## License

[Add your license here]
