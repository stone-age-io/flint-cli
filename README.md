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

### Partial Command Matching

Flint supports Cisco-style partial command matching:

```bash
# These are all equivalent:
flint context list
flint con list
flint cont l
flint context ls

# Ambiguous commands will show suggestions:
flint con c
# Error: ambiguous command 'c'. Possible matches: create

# Unknown commands show available options:
flint context xyz
# Error: unknown command 'xyz'. Available commands: create, list, select, show, delete, organization
```

### Global Flags

```bash
--output, -o     Output format (json|yaml|table) [default: json]
--colors         Enable colored output [default: true]
--debug          Enable debug output [default: false]
```

## Stone-Age.io Collections

Flint is aware of all Stone-Age.io collections and supports authentication with multiple entity types:

### Authentication Collections

**Supported Auth Collections:**
- `users` (default) - Human administrators with organization membership
- `clients` - NATS client authentication entities  
- `edges` - Edge device authentication
- `things` - Individual IoT device authentication
- `service_users` - System service accounts

### Collection Types

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

## Error Handling

Flint provides comprehensive error handling with user-friendly messages:

### Authentication Errors
```bash
# Invalid credentials
Error: Invalid email or password. Please check your credentials and try again.
Suggestion: Try running 'flint auth pb' to authenticate with PocketBase.

# Session expired
Error: Your session has expired. Please authenticate again using 'flint auth pb'.

# Organization access denied
Error: You don't have permission to access resources in this organization. 
Please verify your organization membership or contact your administrator.
```

### Connection Errors
```bash
# Server unreachable
Error: Connection error. Please check your network connection and the Stone-Age.io server URL.

# Server error
Error: Stone-Age.io server error. Please try again later or contact support.
```

### Validation Errors
```bash
# Invalid email format
Error: invalid email format: invalid email format

# Invalid collection
Error: invalid auth collection 'invalid'. Valid options: users, clients, edges, things, service_users
```

## Architecture

```
flint/
├── cmd/                    # Cobra command definitions
│   ├── root.go            # Root command setup
│   ├── context/           # Context management commands
│   └── auth/              # Authentication commands
├── internal/
│   ├── config/            # Context and configuration management
│   ├── pocketbase/        # PocketBase client and operations
│   ├── resolver/          # Partial command matching logic
│   └── utils/             # Shared utilities
└── main.go               # Application entry point
```

## Development

### Phase 1 Completion Checklist

- [x] Context management with organization support
- [x] Configuration files properly managed
- [x] Partial command matching working
- [x] Organization selection functionality
- [x] XDG-compliant directory structure
- [x] Comprehensive error handling
- [x] Stone-Age.io collection definitions
- [x] Utility functions for output and validation

### Phase 2 Completion Checklist

- [x] PocketBase HTTP client implementation
- [x] Authentication system for multiple collections
- [x] Session token management with expiration
- [x] Organization validation via PocketBase API
- [x] Error translation from PocketBase responses
- [x] Interactive and non-interactive authentication
- [x] Comprehensive error handling with suggestions
- [x] Integration with existing context system

### Next Steps (Phase 3)

- [ ] CRUD operations for all Stone-Age.io collections
- [ ] Organization-scoped operations working
- [ ] Filtering, pagination, and formatting
- [ ] Collection validation and helpers
- [ ] Batch operations support

## Contributing

1. Follow Go best practices and maintain DRY principles
2. Add comprehensive comments for maintainability
3. Break complex tasks into smaller, manageable functions
4. Use the established directory structure
5. Test all partial command matching scenarios

## License

[Add your license here]
