package nats

import (
	"crypto/tls"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
	"flint-cli/internal/config"
	"flint-cli/internal/utils"
)

// NewClient creates a new NATS client with the provided configuration
func NewClient(natsConfig *config.NATSConfig) *Client {
	return &Client{
		servers:          natsConfig.Servers,
		authMethod:       natsConfig.AuthMethod,
		username:         natsConfig.Username,
		password:         natsConfig.Password,
		token:            natsConfig.Token,
		credsFile:        natsConfig.CredsFile,
		tlsEnabled:       natsConfig.TLSEnabled,
		tlsVerify:        natsConfig.TLSVerify,
		maxReconnects:    DefaultMaxReconnects,
		reconnectWait:    DefaultReconnectWait,
		pingInterval:     DefaultPingInterval,
		maxPingsOut:      DefaultMaxPingsOut,
		connectionName:   "flint-cli",
	}
}

// NewClientFromContext creates a NATS client from a context configuration
func NewClientFromContext(ctx *config.Context) *Client {
	client := NewClient(&ctx.NATS)
	
	// Convert relative creds file path to absolute if needed
	if client.credsFile != "" && strings.HasPrefix(client.credsFile, "./") {
		// Note: This will be properly resolved when the config manager is available
		// For now, we'll leave it as relative and let the calling code resolve it
		utils.PrintDebug(fmt.Sprintf("Using relative credentials file path: %s", client.credsFile))
	}
	
	return client
}

// Connect establishes a connection to the NATS servers
func (c *Client) Connect() error {
	if c.conn != nil && c.conn.IsConnected() {
		utils.PrintDebug("NATS client already connected")
		return nil
	}

	utils.PrintDebug(fmt.Sprintf("Connecting to NATS servers: %v", c.servers))

	// Build connection options
	opts, err := c.buildConnectionOptions()
	if err != nil {
		return fmt.Errorf("failed to build connection options: %w", err)
	}

	// Connect to NATS
	conn, err := nats.Connect(strings.Join(c.servers, ","), opts...)
	if err != nil {
		return fmt.Errorf("failed to connect to NATS servers: %w", err)
	}

	c.conn = conn
	
	utils.PrintDebug(fmt.Sprintf("Connected to NATS server: %s", conn.ConnectedUrl()))
	utils.PrintInfo(fmt.Sprintf("NATS connection established to %s", conn.ConnectedServerName()))

	return nil
}

// Disconnect closes the NATS connection gracefully
func (c *Client) Disconnect() error {
	if c.conn == nil {
		return nil
	}

	utils.PrintDebug("Disconnecting from NATS server")

	// Drain connection to allow pending messages to be processed
	if err := c.conn.Drain(); err != nil {
		utils.PrintWarning(fmt.Sprintf("Error draining NATS connection: %v", err))
	}

	c.conn.Close()
	c.conn = nil

	utils.PrintDebug("NATS connection closed")
	return nil
}

// IsConnected returns true if the client is connected to NATS
func (c *Client) IsConnected() bool {
	return c.conn != nil && c.conn.IsConnected()
}

// GetConnectionStatus returns detailed connection status information
func (c *Client) GetConnectionStatus() *ConnectionStatus {
	if c.conn == nil {
		return &ConnectionStatus{Connected: false}
	}

	stats := c.conn.Stats()
	
	status := &ConnectionStatus{
		Connected:  c.conn.IsConnected(),
		URL:        c.conn.ConnectedUrl(),
		ServerID:   c.conn.ConnectedServerId(),
		ServerName: c.conn.ConnectedServerName(),
		InMsgs:     stats.InMsgs,
		OutMsgs:    stats.OutMsgs,
		InBytes:    stats.InBytes,
		OutBytes:   stats.OutBytes,
		Reconnects: stats.Reconnects,
		Stats:      &stats,
	}

	// Handle ClientID with error return
	if clientID, err := c.conn.GetClientID(); err == nil {
		status.ClientID = clientID
	} else {
		utils.PrintDebug(fmt.Sprintf("Failed to get client ID: %v", err))
	}

	// Handle RTT with error return
	if rtt, err := c.conn.RTT(); err == nil {
		status.RTT = rtt
	} else {
		utils.PrintDebug(fmt.Sprintf("Failed to get RTT: %v", err))
	}

	// Get last error if available
	if lastErr := c.conn.LastError(); lastErr != nil {
		status.LastError = lastErr.Error()
	}

	return status
}

// executeWithConnection ensures connection and executes operation
func (c *Client) executeWithConnection(operation func() error) error {
	// Ensure we're connected
	if err := c.Connect(); err != nil {
		return fmt.Errorf("failed to establish NATS connection: %w", err)
	}

	// Verify connection is healthy
	if !c.IsConnected() {
		return fmt.Errorf("NATS connection is not healthy")
	}

	// Execute the operation
	return operation()
}

// buildConnectionOptions constructs NATS connection options based on configuration
func (c *Client) buildConnectionOptions() ([]nats.Option, error) {
	var opts []nats.Option

	// Basic connection options
	opts = append(opts,
		nats.Name(c.connectionName),
		nats.MaxReconnects(c.maxReconnects),
		nats.ReconnectWait(c.reconnectWait),
		nats.PingInterval(c.pingInterval),
		nats.MaxPingsOutstanding(c.maxPingsOut),
		nats.Timeout(DefaultConnectTimeout),
		nats.DrainTimeout(DefaultDrainTimeout),
	)

	// Add authentication options
	authOpts, err := c.buildAuthOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build authentication options: %w", err)
	}
	opts = append(opts, authOpts...)

	// Add TLS options
	tlsOpts, err := c.buildTLSOptions()
	if err != nil {
		return nil, fmt.Errorf("failed to build TLS options: %w", err)
	}
	opts = append(opts, tlsOpts...)

	// Add connection event handlers for enhanced debugging
	if config.Global.Debug {
		opts = append(opts, c.buildEventHandlers()...)
	}

	return opts, nil
}

// buildAuthOptions constructs authentication options based on the configured method
func (c *Client) buildAuthOptions() ([]nats.Option, error) {
	var opts []nats.Option

	switch c.authMethod {
	case AuthMethodUserPass:
		if c.username == "" || c.password == "" {
			return nil, fmt.Errorf("username and password are required for user_pass authentication")
		}
		utils.PrintDebug(fmt.Sprintf("Using username/password authentication for user: %s", c.username))
		opts = append(opts, nats.UserInfo(c.username, c.password))

	case AuthMethodToken:
		if c.token == "" {
			return nil, fmt.Errorf("token is required for token authentication")
		}
		utils.PrintDebug("Using token-based authentication")
		opts = append(opts, nats.Token(c.token))

	case AuthMethodCreds:
		if c.credsFile == "" {
			return nil, fmt.Errorf("credentials file path is required for creds authentication")
		}
		utils.PrintDebug(fmt.Sprintf("Using JWT credentials file: %s", c.credsFile))
		opts = append(opts, nats.UserCredentials(c.credsFile))

	default:
		return nil, fmt.Errorf("unsupported authentication method: %s", c.authMethod)
	}

	return opts, nil
}

// buildTLSOptions constructs TLS options based on configuration
func (c *Client) buildTLSOptions() ([]nats.Option, error) {
	var opts []nats.Option

	if c.tlsEnabled {
		utils.PrintDebug("TLS enabled for NATS connection")
		if c.tlsVerify {
			utils.PrintDebug("TLS certificate verification enabled")
			opts = append(opts, nats.Secure())
		} else {
			utils.PrintDebug("TLS certificate verification disabled")
			opts = append(opts, nats.Secure(&tls.Config{
				InsecureSkipVerify: true,
			}))
		}
	}

	return opts, nil
}

// buildEventHandlers creates connection event handlers for debugging
func (c *Client) buildEventHandlers() []nats.Option {
	var opts []nats.Option

	// Connection established handler
	opts = append(opts, nats.ConnectHandler(func(conn *nats.Conn) {
		utils.PrintDebug(fmt.Sprintf("NATS connected to %s (server: %s)", 
			conn.ConnectedUrl(), conn.ConnectedServerName()))
	}))

	// Disconnection handler
	opts = append(opts, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		if err != nil {
			utils.PrintDebug(fmt.Sprintf("NATS disconnected with error: %v", err))
		} else {
			utils.PrintDebug("NATS disconnected")
		}
	}))

	// Reconnection handler
	opts = append(opts, nats.ReconnectHandler(func(conn *nats.Conn) {
		utils.PrintDebug(fmt.Sprintf("NATS reconnected to %s (attempt: %d)", 
			conn.ConnectedUrl(), conn.Stats().Reconnects))
	}))

	// Closed connection handler
	opts = append(opts, nats.ClosedHandler(func(conn *nats.Conn) {
		utils.PrintDebug("NATS connection closed")
	}))

	// Error handler
	opts = append(opts, nats.ErrorHandler(func(conn *nats.Conn, subscription *nats.Subscription, err error) {
		if subscription != nil {
			utils.PrintDebug(fmt.Sprintf("NATS error on subject %s: %v", subscription.Subject, err))
		} else {
			utils.PrintDebug(fmt.Sprintf("NATS connection error: %v", err))
		}
	}))

	return opts
}

// Flush ensures all pending messages are sent to the server
func (c *Client) Flush() error {
	if c.conn == nil {
		return fmt.Errorf("not connected to NATS")
	}

	return c.conn.Flush()
}

// FlushTimeout ensures all pending messages are sent within the timeout
func (c *Client) FlushTimeout(timeout time.Duration) error {
	if c.conn == nil {
		return fmt.Errorf("not connected to NATS")
	}

	return c.conn.FlushTimeout(timeout)
}
