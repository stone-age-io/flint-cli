package nats

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	"flint-cli/internal/utils"
)

var (
	publishHeaders []string
	publishReply   string
	publishFile    string
	publishJSON    bool
)

var publishCmd = &cobra.Command{
	Use:   "publish <subject> [message]",
	Short: "Publish a message to a NATS subject",
	Long: `Publish a message to a Stone-Age.io NATS subject.

The message can be provided as a command line argument, read from a file,
or read from stdin. Headers can be added to provide additional metadata.

Subject Naming:
NATS uses hierarchical subject naming for organized message routing:
  - telemetry.edge.<edge_id>     Edge device telemetry
  - telemetry.thing.<thing_id>   Individual device telemetry  
  - command.edge.<edge_id>       Commands to edge devices
  - command.thing.<thing_id>     Commands to individual devices
  - system.alerts.<type>         System alerts and notifications
  - events.organization.<org_id> Organization-level events

Note: Subject naming is flexible - use patterns that fit your application.

Examples:
  # Publish simple telemetry data
  flint nats publish "telemetry.edge.edge_123" '{"temperature": 22.5}'
  
  # Publish with headers for additional metadata
  flint nats publish "system.alerts.critical" '{"message": "High CPU usage"}' \\
    --header source=monitoring --header priority=high --header timestamp=1642248600
  
  # Publish from file
  flint nats publish "command.thing.door_001" --file ./door_command.json
  
  # Publish JSON data with automatic content-type header
  flint nats publish "events.user.login" '{"user_id": "user_123", "timestamp": "2025-01-15T10:30:00Z"}' --json
  
  # Publish with reply subject for request-response pattern
  flint nats publish "command.edge.edge_456" '{"action": "restart"}' --reply "responses.edge.edge_456"

The command will validate the message format before publishing.
All operations use your active context's NATS configuration for authentication.`,
	Args: cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		subject := args[0]
		
		// Get message data
		var messageData []byte
		var err error
		
		if publishFile != "" {
			// Read from file
			messageData, err = readMessageFile(publishFile)
			if err != nil {
				return fmt.Errorf("failed to read message file: %w", err)
			}
		} else if len(args) > 1 {
			// Use command line argument
			messageData = []byte(args[1])
		} else {
			// Read from stdin
			messageData, err = readFromStdin()
			if err != nil {
				return fmt.Errorf("failed to read message from stdin: %w", err)
			}
		}
		
		// Parse headers
		headers, err := parseHeaders(publishHeaders)
		if err != nil {
			return fmt.Errorf("invalid headers: %w", err)
		}
		
		// Add JSON content type if requested
		if publishJSON {
			if headers == nil {
				headers = make(map[string]string)
			}
			headers["Content-Type"] = "application/json"
			
			// Validate JSON format
			if !json.Valid(messageData) {
				return fmt.Errorf("invalid JSON format in message data")
			}
		}
		
		// Validate NATS subject format (basic validation only)
		if err := utils.ValidateNATSSubject(subject); err != nil {
			return fmt.Errorf("invalid subject: %w", err)
		}
		
		// Create NATS client
		client, err := createNATSClient()
		if err != nil {
			return fmt.Errorf("failed to create NATS client: %w", err)
		}
		defer func() {
			if disconnectErr := client.Disconnect(); disconnectErr != nil {
				utils.PrintWarning(fmt.Sprintf("Error disconnecting from NATS: %v", disconnectErr))
			}
		}()
		
		// Show what we're about to publish (if verbose)
		if verbose {
			displayPublishInfo(subject, messageData, headers, publishReply)
		}
		
		// Parse timeout
		timeoutDuration, err := time.ParseDuration(timeout)
		if err != nil {
			return fmt.Errorf("invalid timeout format: %w", err)
		}
		
		// Publish the message
		utils.PrintInfo(fmt.Sprintf("Publishing to subject: %s", subject))
		
		if publishReply != "" {
			err = client.PublishWithReply(subject, publishReply, messageData, headers)
		} else {
			err = client.Publish(subject, messageData, headers)
		}
		
		if err != nil {
			return fmt.Errorf("failed to publish message: %w", err)
		}
		
		// Ensure message is sent to server within timeout
		utils.PrintDebug(fmt.Sprintf("Flushing connection with timeout: %v", timeoutDuration))
		if flushErr := client.FlushTimeout(timeoutDuration); flushErr != nil {
			utils.PrintWarning(fmt.Sprintf("Warning: Failed to confirm message delivery within %v: %v", timeoutDuration, flushErr))
			// Don't fail the operation - message was published, just confirmation timed out
		}
		
		// Success message
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("%s Message published successfully!\n", green("✓"))
		
		// Show publication details
		fmt.Printf("  Subject: %s\n", subject)
		fmt.Printf("  Size: %d bytes\n", len(messageData))
		
		if len(headers) > 0 {
			fmt.Printf("  Headers: %d\n", len(headers))
		}
		
		if publishReply != "" {
			fmt.Printf("  Reply Subject: %s\n", publishReply)
		}
		
		// Show message content based on output format
		if outputFormat != "text" {
			fmt.Printf("\nPublished Message:\n")
			return displayMessageContent(messageData, headers, outputFormat)
		}
		
		return nil
	},
}

func init() {
	publishCmd.Flags().StringSliceVar(&publishHeaders, "header", nil, 
		"Message headers in key=value format (can be used multiple times)")
	publishCmd.Flags().StringVar(&publishReply, "reply", "", 
		"Reply subject for request-response pattern")
	publishCmd.Flags().StringVar(&publishFile, "file", "", 
		"Read message data from file instead of command line")
	publishCmd.Flags().BoolVar(&publishJSON, "json", false, 
		"Treat message as JSON and add Content-Type header")
}

// readMessageFile reads message data from a file
func readMessageFile(filename string) ([]byte, error) {
	if err := utils.ValidateFileExists(filename); err != nil {
		return nil, err
	}
	
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s': %w", filename, err)
	}
	
	return data, nil
}

// readFromStdin reads message data from stdin
func readFromStdin() ([]byte, error) {
	utils.PrintInfo("Reading message from stdin (press Ctrl+D when done)...")
	
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("failed to read from stdin: %w", err)
	}
	
	return data, nil
}

// parseHeaders parses header strings in key=value format
func parseHeaders(headerStrings []string) (map[string]string, error) {
	if len(headerStrings) == 0 {
		return nil, nil
	}
	
	headers := make(map[string]string)
	
	for _, headerStr := range headerStrings {
		parts := strings.SplitN(headerStr, "=", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid header format '%s'. Use key=value format", headerStr)
		}
		
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		
		if key == "" {
			return nil, fmt.Errorf("header key cannot be empty in '%s'", headerStr)
		}
		
		// Validate header field
		if err := utils.ValidateNATSHeaderField(key, value); err != nil {
			return nil, fmt.Errorf("invalid header '%s': %w", key, err)
		}
		
		headers[key] = value
	}
	
	return headers, nil
}

// displayPublishInfo shows information about what will be published
func displayPublishInfo(subject string, data []byte, headers map[string]string, reply string) {
	fmt.Printf("Publishing Message:\n")
	fmt.Printf("  Subject: %s\n", subject)
	fmt.Printf("  Size: %d bytes\n", len(data))
	
	if reply != "" {
		fmt.Printf("  Reply Subject: %s\n", reply)
	}
	
	if len(headers) > 0 {
		fmt.Printf("  Headers:\n")
		for key, value := range headers {
			fmt.Printf("    %s: %s\n", key, value)
		}
	}
	
	fmt.Printf("  Data Preview: %s\n", truncateForDisplay(string(data), 100))
	fmt.Println()
}

// displayMessageContent displays the message content in the specified format
func displayMessageContent(data []byte, headers map[string]string, format string) error {
	message := map[string]interface{}{
		"data": string(data),
		"size": len(data),
	}
	
	if len(headers) > 0 {
		message["headers"] = headers
	}
	
	return utils.OutputData(message, format)
}

// truncateForDisplay truncates text for display purposes
func truncateForDisplay(text string, maxLen int) string {
	if len(text) <= maxLen {
		return text
	}
	return text[:maxLen-3] + "..."
}
