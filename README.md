# Flint CLI

A command-line interface for managing the Stone-Age.io IoT platform. Flint provides comprehensive tools for context management, PocketBase authentication, collection operations, and NATS messaging.

## Features

- **Multi-environment context management** with organization support and directory-based configuration
- **PocketBase authentication** supporting multiple collection types (users, clients, edges, things, service_users)  
- **Complete CRUD operations** for all Stone-Age.io collections with filtering, pagination, and expansion
- **NATS messaging** with multiple authentication methods (user_pass, token, creds) and full pub/sub functionality
- **Partial command matching** with Cisco-style abbreviated commands
- **Comprehensive error handling** with user-friendly messages and suggestions
- **XDG-compliant configuration** with self-contained context directories

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

### 2. Select Active Context and Authenticate

```bash
# Select the production context
flint context select production

# Authenticate with PocketBase
flint auth pb --email admin@company.com

# Set your organization
flint context organization org_abc123def456789

# Configure NATS authentication
flint auth nats --method creds --creds-file /path/to/nats.creds
```

### 3. Start Using Collections and Messaging

```bash
# List all edges in your organization
flint collections edges list

# Create a new edge device
flint collections edges create '{"name":"Building-A-Edge","code":"bldg-a","type":"controller","region":"west"}'

# Publish a message via NATS
flint nats publish "telemetry.edge.edge_123" '{"temperature": 22.5, "timestamp": "2025-01-15T10:30:00Z"}'

# Subscribe to messages
flint nats subscribe "telemetry.>"
```

## Configuration

Flint uses XDG-compliant configuration directories with a context-based structure where each context has its own directory:

```
~/.config/flint/
├── config.yaml           # Global configuration
├── production/           # Production context directory
│   ├── context.yaml     # Context configuration
│   └── nats.creds       # NATS credentials (if using creds auth)
├── development/          # Development context directory
│   ├── context.yaml
│   └── nats.creds
└── staging/              # Staging context directory
    ├── context.yaml
    └── nats.creds
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

### Context Configuration (`~/.config/flint/production/context.yaml`)

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
  creds_file: ./nats.creds # Relative path to context directory
  username: ""
  password: ""
  token: ""
  tls_enabled: true
  tls_verify: true
```

### NATS Authentication Methods

**Credentials File (Recommended)**
```bash
# Configure with credentials file
flint auth nats --method creds --creds-file /path/to/client.creds

# Or place file in context directory for automatic relative path
cp client.creds ~/.config/flint/production/nats.creds
flint auth nats --method creds --creds-file ./nats.creds
```

**Username/Password**
```bash
# Configure username/password authentication
flint auth nats --method user_pass --username client001 --password secret
```

**Token Authentication**
```bash
# Configure token authentication
flint auth nats --method token --token eyJhbGciOiJSUzI1NiIs...
```

## Commands

### Context Management

```bash
# Create new context
flint context create <name> [flags]
  --pb-url string              PocketBase URL (required)
  --pb-auth-collection string  Auth collection (default: "users")
  --nats-servers strings       NATS server URLs (comma-separated, required)
  --nats-auth-method string    NATS auth method (user_pass|token|creds)

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

# NATS authentication configuration
flint auth nats [flags]
  --method string        Authentication method (user_pass|token|creds)
  --username string      Username for user_pass authentication
  --password string      Password for user_pass authentication
  --token string         JWT token for token authentication
  --creds-file string    Path to JWT credentials file for creds authentication
  --test                 Test NATS connection after configuration
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

### NATS Operations

```bash
# Publish a message to a NATS subject
flint nats publish <subject> [message] [flags]
  --header strings       Message headers in key=value format (multiple allowed)
  --reply string         Reply subject for request-response pattern
  --file string          Read message data from file
  --json                 Treat message as JSON and add Content-Type header

# Subscribe to messages from a NATS subject
flint nats subscribe <subject> [flags]
  --queue string         Queue group name for load-balanced subscription
  --timeout duration     Subscription timeout (default: unlimited)
  --count int            Stop after receiving this many messages (0 = unlimited)
  --raw                  Display raw message data without formatting
  --headers              Display message headers
  --timestamp            Display message timestamps
```

### Global Flags

```bash
--output, -o     Output format (json|yaml|table) [default: json]
--colors         Enable colored output [default: true]
--debug          Enable debug output [default: false]
```

## Stone-Age.io Collections

Flint supports all Stone-Age.io collections with full CRUD operations:

**Core Collections:**
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

**Authentication Collections:**
- `users` (default) - Human administrators
- `clients` - NATS client entities  
- `edges` - Edge device authentication
- `things` - Individual IoT device authentication
- `service_users` - System service accounts

## Partial Command Matching

Flint supports Cisco-style partial command matching for collection actions only:

```bash
# Collection actions support partial matching:
flint collections edges list
flint collections edges l        # "l" resolves to "list" 
flint collections edges cr       # "cr" resolves to "create"
flint collections edges up       # "up" resolves to "update"

# Main commands and collection names must be exact:
flint collections edges list    # ✓ Correct
flint col edges list            # ✗ Invalid - "col" not recognized
flint collections edg list      # ✗ Invalid - collection name must be exact

# Ambiguous actions show suggestions:
flint collections edges c
# Error: ambiguous command 'c'. Possible matches: create

# Unknown actions show available options:
flint collections edges xyz
# Error: unknown command 'xyz'. Available commands: create, delete, get, list, update
```

## Examples

### Context Management Workflows

```bash
# Create contexts for different environments
flint context create production --pb-url https://api.stone-age.io --nats-servers nats://nats.stone-age.io:4222
flint context create development --pb-url http://localhost:8090 --nats-servers nats://localhost:4222

# Switch between contexts
flint context select production
flint context select development

# View context details and directory structure
flint context show production
flint context list
```

### Authentication Workflows

```bash
# PocketBase authentication with organization setup
flint auth pb --email admin@company.com --password secret
flint context organization org_abc123def456789

# NATS authentication configuration
flint auth nats --method creds --creds-file ./nats.creds --test
flint auth nats --method user_pass --username client001 --password secret
flint auth nats --method token --token eyJhbGciOiJSUzI1NiIs...
```

### Collection Management Workflows

```bash
# Organization and user management
flint collections organizations create '{"name":"ACME Corp","code":"acme","account_name":"acme-corp"}'
flint collections users create '{"email":"john@acme.com","password":"secure123","first_name":"John","last_name":"Doe"}'
flint collections users list --fields email,first_name,last_name,is_org_admin

# Edge and device management
flint collections edges create '{"name":"Building-A-Gateway","code":"bldg-a-gw","type":"gateway","region":"us-west"}'
flint collections things create '{"name":"Door-Controller-01","code":"door-01","type":"access_control","edge_id":"edge_abc123def456"}'

# Advanced filtering and pagination
flint collections edges list --filter 'active=true && region~"us-"' --sort '-created' --limit 10
flint collections things list --filter 'edge_id="edge_abc123def456"' --expand edge_id

# Using files for create/update operations
echo '{"name":"Test Edge","code":"test-edge","type":"controller","region":"test"}' > edge.json
flint collections edges create --file edge.json
```

### NATS Messaging Workflows

```bash
# Basic publish and subscribe
flint nats publish "telemetry.temperature" '{"sensor": "temp001", "value": 22.5}'
flint nats subscribe "telemetry.>"

# Publishing with headers and metadata
flint nats publish "system.alerts" '{"level": "warning", "message": "High CPU usage"}' \
  --header source=monitoring \
  --header priority=high \
  --header timestamp=$(date -u +%s)

# Subscribing with queue groups for load balancing
flint nats subscribe "commands.>" --queue command_processors

# Publishing from files
echo '{"action": "restart", "component": "edge_service"}' > command.json
flint nats publish "commands.edge.edge_123" --file command.json --json

# Subscribing with limits and formatting
flint nats subscribe "events.>" --count 10 --timeout 30s --headers --timestamp
flint nats subscribe "telemetry.sensors" --raw --output json
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

**NATS Message (telemetry.json):**
```json
{
  "timestamp": "2025-01-15T10:30:00Z",
  "sensor_id": "temp_001",
  "location": "warehouse_a",
  "readings": {
    "temperature": 22.5,
    "humidity": 45.2,
    "pressure": 1013.25
  },
  "status": "ok"
}
```

## Error Handling

Flint provides comprehensive error handling with user-friendly messages:

### Collection Errors
```bash
# Invalid collection
$ flint collections invalid_collection list
Error: collection 'invalid_collection' not available in current context. Available collections: organizations, users, edges, things, locations, clients, edge_types, thing_types, location_types, edge_regions, audit_logs, topic_permissions

# Authentication errors
$ flint collections edges list
Error: your session has expired. Please authenticate again using 'flint auth pb'.
```

### NATS Errors
```bash
# Connection errors
$ flint nats publish test.subject "hello"
Error: NATS connection error: no NATS servers are currently available.
Suggestion: check your network connection and verify the NATS server URLs in your context

# Authentication errors
$ flint nats subscribe "test.>"
Error: NATS authentication error: NATS authentication failed.
Suggestion: verify your credentials are correct and check your context configuration
```

## Architecture

```
flint/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command setup
│   ├── context/           # Context management commands
│   ├── auth/              # Authentication commands (PocketBase & NATS)
│   ├── collections/       # Collection CRUD operations
│   └── nats/              # NATS messaging commands
├── internal/
│   ├── config/            # Context and configuration management
│   ├── pocketbase/        # PocketBase client and operations
│   ├── nats/              # NATS client and operations
│   ├── resolver/          # Partial command matching logic  
│   └── utils/             # Shared utilities
└── main.go               # Application entry point
```

## Future Features

The following features are planned for future development:

### File Operations & Final Polish
- **Generic file operations** for uploading and downloading files from any collection field
- **Progress indicators** for large file transfers with resumable uploads
- **File validation** and type checking before upload operations
- **Comprehensive testing suite** with unit and integration tests
- **Complete documentation** with advanced usage examples and best practices
- **Automated builds and releases** with GitHub Actions and cross-platform binaries
- **Version management** with update notifications and automatic updates
- **Enhanced CLI experience** with shell completion and improved help system

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
