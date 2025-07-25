package nats

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	natsClient "flint-cli/internal/nats"
	"flint-cli/internal/utils"
)

var (
	subscribeQueue     string
	subscribeCount     int
	subscribeRaw       bool
	subscribeHeaders   bool
	subscribeTimestamp bool
)

var subscribeCmd = &cobra.Command{
	Use:   "subscribe <subject>",
	Short: "Subscribe to messages from a NATS subject",
	Long: `Subscribe to messages from a Stone-Age.io NATS subject and display them in real-time.

The subscription will continue until interrupted with Ctrl+C or until the specified
timeout is reached. Messages are displayed with formatting based on the output format.

Subject Patterns:
Use NATS wildcards to subscribe to multiple subjects:
  - "*" matches a single token: "telemetry.*.temperature"
  - ">" matches multiple tokens: "telemetry.>" (all telemetry)

NATS Subject Examples:
  telemetry.>              All telemetry messages
  telemetry.edge.*         All edge telemetry
  telemetry.thing.thing_123 Specific thing telemetry
  command.edge.*           Commands to any edge
  system.alerts.>          All system alerts
  events.organization.*    Organization events

Note: Subject patterns are flexible - use wildcards that match your needs.

Queue Groups:
Use queue groups to distribute messages among multiple subscribers:
  - All subscribers in the same queue group form a load-balanced group
  - Only one subscriber in the group receives each message
  - Useful for scalable message processing

Examples:
  # Subscribe to all telemetry messages
  flint nats subscribe "telemetry.>"
  
  # Subscribe to edge commands with queue group for load balancing
  flint nats subscribe "command.edge.*" --queue edge_processors
  
  # Subscribe with timeout and message count limit
  flint nats subscribe "system.alerts.critical" --timeout 30s --count 10
  
  # Subscribe with raw output (no formatting)
  flint nats subscribe "events.user.*" --raw --output json
  
  # Subscribe and show message headers and timestamps
  flint nats subscribe "telemetry.edge.edge_123" --headers --timestamp

Press Ctrl+C to stop the subscription at any time.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		subject := args[0]
		
		// Validate NATS subject format (basic validation only)
		if err := utils.ValidateNATSSubject(subject); err != nil {
			return fmt.Errorf("invalid subject: %w", err)
		}
		
		// Validate queue name if provided
		if subscribeQueue != "" {
			if err := utils.ValidateNATSQueue(subscribeQueue); err != nil {
				return fmt.Errorf("invalid queue name: %w", err)
			}
		}
		
		// Parse timeout
		var timeoutDuration time.Duration
		var err error
		if timeout != "" && timeout != "0" {
			timeoutDuration, err = time.ParseDuration(timeout)
			if err != nil {
				return fmt.Errorf("invalid timeout format: %w", err)
			}
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
		
		// Display subscription info
		displaySubscriptionInfo(subject, subscribeQueue, timeoutDuration, subscribeCount)
		
		// Create message handler based on output preferences
		var messageCount int
		handler := func(msg *natsClient.Message) error {
			messageCount++
			
			// Display the message
			if subscribeRaw {
				displayRawMessage(msg)
			} else {
				displayFormattedMessage(msg, outputFormat, subscribeHeaders, subscribeTimestamp)
			}
			
			// Check if we've reached the message count limit
			if subscribeCount > 0 && messageCount >= subscribeCount {
				utils.PrintInfo(fmt.Sprintf("Reached message limit (%d), stopping subscription...", subscribeCount))
				return fmt.Errorf("message_limit_reached") // Special error to trigger stop
			}
			
			return nil
		}
		
		// Set up signal handling for graceful shutdown
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		
		// Start the subscription in a goroutine
		errChan := make(chan error, 1)
		go func() {
			err := client.Subscribe(subject, subscribeQueue, handler, timeoutDuration)
			errChan <- err
		}()
		
		// Wait for completion or interruption
		select {
		case err := <-errChan:
			if err != nil && err.Error() != "message_limit_reached" {
				return fmt.Errorf("subscription error: %w", err)
			}
		case sig := <-sigChan:
			utils.PrintInfo(fmt.Sprintf("Received signal %v, stopping subscription...", sig))
		}
		
		// Show final statistics
		green := color.New(color.FgGreen).SprintFunc()
		fmt.Printf("\n%s Subscription completed\n", green("✓"))
		fmt.Printf("  Subject: %s\n", subject)
		if subscribeQueue != "" {
			fmt.Printf("  Queue: %s\n", subscribeQueue)
		}
		fmt.Printf("  Messages received: %d\n", messageCount)
		
		return nil
	},
}

func init() {
	subscribeCmd.Flags().StringVar(&subscribeQueue, "queue", "", 
		"Queue group name for load-balanced subscription")
	subscribeCmd.Flags().IntVar(&subscribeCount, "count", 0, 
		"Stop after receiving this many messages (0 = unlimited)")
	subscribeCmd.Flags().BoolVar(&subscribeRaw, "raw", false, 
		"Display raw message data without formatting")
	subscribeCmd.Flags().BoolVar(&subscribeHeaders, "headers", false, 
		"Display message headers")
	subscribeCmd.Flags().BoolVar(&subscribeTimestamp, "timestamp", false, 
		"Display message timestamps")
}

// displaySubscriptionInfo shows information about the subscription being created
func displaySubscriptionInfo(subject, queue string, timeout time.Duration, count int) {
	cyan := color.New(color.FgCyan).SprintFunc()
	
	fmt.Printf("Subscribing to NATS:\n")
	fmt.Printf("  Subject: %s\n", cyan(subject))
	
	if queue != "" {
		fmt.Printf("  Queue Group: %s\n", cyan(queue))
	}
	
	if timeout > 0 {
		fmt.Printf("  Timeout: %v\n", timeout)
	}
	
	if count > 0 {
		fmt.Printf("  Message Limit: %d\n", count)
	}
	
	fmt.Printf("\nWaiting for messages... (Press Ctrl+C to stop)\n")
	fmt.Println(strings.Repeat("═", 80))
}

// displayRawMessage displays a message in raw format
func displayRawMessage(msg *natsClient.Message) {
	fmt.Printf("[%s] %s: %s\n", 
		msg.Timestamp.Format("15:04:05.000"), 
		msg.Subject, 
		string(msg.Data))
}

// displayFormattedMessage displays a message with full formatting
func displayFormattedMessage(msg *natsClient.Message, format string, showHeaders, showTimestamp bool) {
	switch format {
	case "json":
		displayJSONMessage(msg, showHeaders, showTimestamp)
	case "yaml":
		displayYAMLMessage(msg, showHeaders, showTimestamp)
	default:
		displayTextMessage(msg, showHeaders, showTimestamp)
	}
}

// displayTextMessage displays a message in human-readable text format
func displayTextMessage(msg *natsClient.Message, showHeaders, showTimestamp bool) {
	// Header with subject
	fmt.Printf("┌─ Subject: %s", msg.Subject)
	if msg.Reply != "" {
		fmt.Printf(" (Reply: %s)", msg.Reply)
	}
	fmt.Printf(" [%d bytes]\n", msg.Size)
	
	// Timestamp if requested
	if showTimestamp {
		fmt.Printf("├─ Time: %s\n", msg.Timestamp.Format("2006-01-02 15:04:05.000"))
	}
	
	// Headers if present and requested
	if showHeaders && len(msg.Headers) > 0 {
		fmt.Printf("├─ Headers:\n")
		for key, value := range msg.Headers {
			fmt.Printf("│  %s: %s\n", key, value)
		}
	}
	
	// Message data
	fmt.Printf("└─ Data:\n")
	
	// Try to format JSON if it looks like JSON
	if len(msg.Data) > 0 && (msg.Data[0] == '{' || msg.Data[0] == '[') {
		if formatted, err := utils.FormatJSON(msg.Data); err == nil {
			// Indent the JSON
			lines := strings.Split(formatted, "\n")
			for _, line := range lines {
				fmt.Printf("   %s\n", line)
			}
		} else {
			fmt.Printf("   %s\n", string(msg.Data))
		}
	} else {
		fmt.Printf("   %s\n", string(msg.Data))
	}
	
	fmt.Println()
}

// displayJSONMessage displays a message as JSON
func displayJSONMessage(msg *natsClient.Message, showHeaders, showTimestamp bool) {
	output := map[string]interface{}{
		"subject": msg.Subject,
		"data":    string(msg.Data),
		"size":    msg.Size,
	}
	
	if msg.Reply != "" {
		output["reply"] = msg.Reply
	}
	
	if showTimestamp {
		output["timestamp"] = msg.Timestamp.Format(time.RFC3339Nano)
	}
	
	if showHeaders && len(msg.Headers) > 0 {
		output["headers"] = msg.Headers
	}
	
	if err := utils.OutputData(output, "json"); err != nil {
		utils.PrintError(fmt.Errorf("failed to format message as JSON: %w", err))
		fmt.Printf("Raw: %s\n", string(msg.Data))
	}
}

// displayYAMLMessage displays a message as YAML
func displayYAMLMessage(msg *natsClient.Message, showHeaders, showTimestamp bool) {
	output := map[string]interface{}{
		"subject": msg.Subject,
		"data":    string(msg.Data),
		"size":    msg.Size,
	}
	
	if msg.Reply != "" {
		output["reply"] = msg.Reply
	}
	
	if showTimestamp {
		output["timestamp"] = msg.Timestamp.Format(time.RFC3339Nano)
	}
	
	if showHeaders && len(msg.Headers) > 0 {
		output["headers"] = msg.Headers
	}
	
	if err := utils.OutputData(output, "yaml"); err != nil {
		utils.PrintError(fmt.Errorf("failed to format message as YAML: %w", err))
		fmt.Printf("Raw: %s\n", string(msg.Data))
	}
}
