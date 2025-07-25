package nats

import (
	"time"

	"github.com/nats-io/nats.go"
)

// Client represents a NATS client for Stone-Age.io messaging
type Client struct {
	conn       *nats.Conn
	servers    []string
	authMethod string
	
	// Authentication credentials
	username   string
	password   string
	token      string
	credsFile  string
	
	// TLS configuration
	tlsEnabled bool
	tlsVerify  bool
	
	// Connection options
	maxReconnects    int
	reconnectWait    time.Duration
	pingInterval     time.Duration
	maxPingsOut      int
	connectionName   string
}

// ConnectionStatus represents the current NATS connection status
type ConnectionStatus struct {
	Connected      bool              `json:"connected"`
	URL            string            `json:"url"`
	ServerID       string            `json:"server_id"`
	ServerName     string            `json:"server_name"`
	ClientID       uint64            `json:"client_id"`
	RTT            time.Duration     `json:"rtt"`
	InMsgs         uint64            `json:"in_msgs"`
	OutMsgs        uint64            `json:"out_msgs"`
	InBytes        uint64            `json:"in_bytes"`
	OutBytes       uint64            `json:"out_bytes"`
	Reconnects     uint64            `json:"reconnects"`
	LastError      string            `json:"last_error,omitempty"`
	Stats          *nats.Statistics  `json:"stats,omitempty"`
}

// PublishOptions contains options for publishing messages
type PublishOptions struct {
	Subject string                 `json:"subject"`
	Data    []byte                 `json:"data"`
	Headers map[string]string      `json:"headers,omitempty"`
	Reply   string                 `json:"reply,omitempty"`
}

// SubscribeOptions contains options for subscribing to messages
type SubscribeOptions struct {
	Subject        string        `json:"subject"`
	Queue          string        `json:"queue,omitempty"`
	Timeout        time.Duration `json:"timeout,omitempty"`
	MaxMsgs        int           `json:"max_msgs,omitempty"`
	AutoUnsubscribe bool          `json:"auto_unsubscribe,omitempty"`
}

// RequestOptions contains options for request-reply messaging
type RequestOptions struct {
	Subject string            `json:"subject"`
	Data    []byte            `json:"data"`
	Timeout time.Duration     `json:"timeout"`
	Headers map[string]string `json:"headers,omitempty"`
}

// Message represents a received NATS message with metadata
type Message struct {
	Subject   string            `json:"subject"`
	Reply     string            `json:"reply,omitempty"`
	Data      []byte            `json:"data"`
	Headers   map[string]string `json:"headers,omitempty"`
	Timestamp time.Time         `json:"timestamp"`
	Size      int               `json:"size"`
}

// AuthMethod constants for NATS authentication
const (
	AuthMethodUserPass = "user_pass"
	AuthMethodToken    = "token"
	AuthMethodCreds    = "creds"
)

// Default connection configuration values
const (
	DefaultMaxReconnects   = 5
	DefaultReconnectWait   = 2 * time.Second
	DefaultPingInterval    = 30 * time.Second
	DefaultMaxPingsOut     = 3
	DefaultRequestTimeout  = 5 * time.Second
	DefaultConnectTimeout  = 10 * time.Second
	DefaultDrainTimeout    = 30 * time.Second
)

// ConnectionEvent represents connection state changes
type ConnectionEvent struct {
	Type      string    `json:"type"`       // connected, disconnected, reconnected, error
	Timestamp time.Time `json:"timestamp"`
	URL       string    `json:"url,omitempty"`
	Error     string    `json:"error,omitempty"`
	Attempt   int       `json:"attempt,omitempty"`
}

// MessageStats represents message processing statistics
type MessageStats struct {
	Published   uint64 `json:"published"`
	Received    uint64 `json:"received"`
	Errors      uint64 `json:"errors"`
	LastMessage time.Time `json:"last_message"`
}

// SubscriptionInfo represents active subscription information
type SubscriptionInfo struct {
	Subject      string    `json:"subject"`
	Queue        string    `json:"queue,omitempty"`
	Delivered    uint64    `json:"delivered"`
	Dropped      uint64    `json:"dropped"`
	Pending      int       `json:"pending"`
	MaxPending   int       `json:"max_pending"`
	Created      time.Time `json:"created"`
	LastActivity time.Time `json:"last_activity"`
}
