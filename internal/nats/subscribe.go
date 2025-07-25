package nats

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nats-io/nats.go"
	"flint-cli/internal/utils"
)

// SubscriptionHandler defines a function type for handling received messages
type SubscriptionHandler func(*Message) error

// Subscribe subscribes to a NATS subject and processes messages
func (c *Client) Subscribe(subject, queue string, handler SubscriptionHandler, timeout time.Duration) error {
	return c.executeWithConnection(func() error {
		return c.subscribeToSubject(subject, queue, handler, timeout)
	})
}

// SubscribeSync subscribes to a subject and returns messages synchronously
func (c *Client) SubscribeSync(subject, queue string, timeout time.Duration) (*nats.Subscription, error) {
	var subscription *nats.Subscription
	
	err := c.executeWithConnection(func() error {
		// Validate inputs
		if err := utils.ValidateNATSSubject(subject); err != nil {
			return WrapNATSError("subscribe", subject, err)
		}
		
		if queue != "" {
			if err := utils.ValidateNATSQueue(queue); err != nil {
				return WrapNATSError("subscribe", subject, err)
			}
		}
		
		utils.PrintDebug(fmt.Sprintf("Creating synchronous subscription to subject: %s (queue: %s)", subject, queue))
		
		var err error
		if queue != "" {
			subscription, err = c.conn.QueueSubscribeSync(subject, queue)
		} else {
			subscription, err = c.conn.SubscribeSync(subject)
		}
		
		if err != nil {
			return WrapNATSError("subscribe", subject, err)
		}
		
		utils.PrintInfo(fmt.Sprintf("Subscribed to subject: %s", subject))
		return nil
	})
	
	return subscription, err
}

// SubscribeWithContext subscribes with context for cancellation support
func (c *Client) SubscribeWithContext(ctx context.Context, subject, queue string, handler SubscriptionHandler) error {
	return c.executeWithConnection(func() error {
		return c.subscribeWithContext(ctx, subject, queue, handler)
	})
}

// StreamMessages continuously streams messages from a subject until interrupted
func (c *Client) StreamMessages(subject, queue string, timeout time.Duration) error {
	return c.Subscribe(subject, queue, func(msg *Message) error {
		// Display the message to stdout
		DisplayMessage(msg)
		return nil
	}, timeout)
}

// subscribeToSubject is the internal method that handles subscription logic
func (c *Client) subscribeToSubject(subject, queue string, handler SubscriptionHandler, timeout time.Duration) error {
	// Validate inputs
	if err := utils.ValidateNATSSubject(subject); err != nil {
		return WrapNATSError("subscribe", subject, err)
	}
	
	if queue != "" {
		if err := utils.ValidateNATSQueue(queue); err != nil {
			return WrapNATSError("subscribe", subject, err)
		}
	}
	
	if handler == nil {
		return WrapNATSError("subscribe", subject, fmt.Errorf("message handler cannot be nil"))
	}
	
	utils.PrintDebug(fmt.Sprintf("Subscribing to subject: %s (queue: %s, timeout: %v)", subject, queue, timeout))
	
	// Create the subscription
	var subscription *nats.Subscription
	var err error
	
	// Wrap the handler to convert NATS messages to our Message format
	wrappedHandler := func(msg *nats.Msg) {
		message := convertNATSMessage(msg)
		if handlerErr := handler(message); handlerErr != nil {
			utils.PrintWarning(fmt.Sprintf("Message handler error: %v", handlerErr))
		}
	}
	
	if queue != "" {
		subscription, err = c.conn.QueueSubscribe(subject, queue, wrappedHandler)
	} else {
		subscription, err = c.conn.Subscribe(subject, wrappedHandler)
	}
	
	if err != nil {
		return WrapNATSError("subscribe", subject, err)
	}
	
	defer func() {
		if err := subscription.Unsubscribe(); err != nil {
			utils.PrintWarning(fmt.Sprintf("Error unsubscribing: %v", err))
		}
	}()
	
	// Print subscription info
	utils.PrintInfo(fmt.Sprintf("Subscribed to subject: %s", subject))
	if queue != "" {
		utils.PrintInfo(fmt.Sprintf("Queue group: %s", queue))
	}
	utils.PrintInfo("Press Ctrl+C to stop...")
	
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	// Set up timeout if specified
	var timeoutChan <-chan time.Time
	if timeout > 0 {
		timeoutTimer := time.NewTimer(timeout)
		defer timeoutTimer.Stop()
		timeoutChan = timeoutTimer.C
		utils.PrintDebug(fmt.Sprintf("Subscription will timeout after %v", timeout))
	}
	
	// Wait for interrupt signal or timeout
	select {
	case sig := <-sigChan:
		utils.PrintInfo(fmt.Sprintf("Received signal %v, stopping subscription...", sig))
	case <-timeoutChan:
		utils.PrintInfo("Subscription timeout reached, stopping...")
	}
	
	return nil
}

// subscribeWithContext handles subscription with context cancellation
func (c *Client) subscribeWithContext(ctx context.Context, subject, queue string, handler SubscriptionHandler) error {
	// Validate inputs
	if err := utils.ValidateNATSSubject(subject); err != nil {
		return WrapNATSError("subscribe", subject, err)
	}
	
	if queue != "" {
		if err := utils.ValidateNATSQueue(queue); err != nil {
			return WrapNATSError("subscribe", subject, err)
		}
	}
	
	if handler == nil {
		return WrapNATSError("subscribe", subject, fmt.Errorf("message handler cannot be nil"))
	}
	
	utils.PrintDebug(fmt.Sprintf("Subscribing with context to subject: %s (queue: %s)", subject, queue))
	
	// Create the subscription
	var subscription *nats.Subscription
	var err error
	
	// Wrap the handler to convert NATS messages to our Message format
	wrappedHandler := func(msg *nats.Msg) {
		message := convertNATSMessage(msg)
		if handlerErr := handler(message); handlerErr != nil {
			utils.PrintWarning(fmt.Sprintf("Message handler error: %v", handlerErr))
		}
	}
	
	if queue != "" {
		subscription, err = c.conn.QueueSubscribe(subject, queue, wrappedHandler)
	} else {
		subscription, err = c.conn.Subscribe(subject, wrappedHandler)
	}
	
	if err != nil {
		return WrapNATSError("subscribe", subject, err)
	}
	
	defer func() {
		if err := subscription.Unsubscribe(); err != nil {
			utils.PrintWarning(fmt.Sprintf("Error unsubscribing: %v", err))
		}
	}()
	
	utils.PrintInfo(fmt.Sprintf("Subscribed to subject: %s", subject))
	if queue != "" {
		utils.PrintInfo(fmt.Sprintf("Queue group: %s", queue))
	}
	
	// Wait for context cancellation
	<-ctx.Done()
	utils.PrintInfo("Context cancelled, stopping subscription...")
	
	return ctx.Err()
}

// GetSubscriptionInfo returns information about active subscriptions
func (c *Client) GetSubscriptionInfo() ([]*SubscriptionInfo, error) {
	if c.conn == nil {
		return nil, fmt.Errorf("not connected to NATS")
	}
	
	// Note: The NATS Go client doesn't provide a direct way to list all subscriptions
	// This would need to be tracked manually if detailed subscription info is needed
	return []*SubscriptionInfo{}, nil
}

// ValidateSubscribeOptions validates subscription parameters
func ValidateSubscribeOptions(opts *SubscribeOptions) error {
	if opts == nil {
		return fmt.Errorf("subscribe options cannot be nil")
	}
	
	if err := utils.ValidateNATSSubject(opts.Subject); err != nil {
		return fmt.Errorf("invalid subject: %w", err)
	}
	
	if opts.Queue != "" {
		if err := utils.ValidateNATSQueue(opts.Queue); err != nil {
			return fmt.Errorf("invalid queue: %w", err)
		}
	}
	
	if opts.Timeout < 0 {
		return fmt.Errorf("timeout cannot be negative")
	}
	
	if opts.MaxMsgs < 0 {
		return fmt.Errorf("max messages cannot be negative")
	}
	
	return nil
}

// DisplayMessage formats and displays a received NATS message
func DisplayMessage(msg *Message) {
	// Use a structured format for message display
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("Subject: %s\n", msg.Subject)
	
	if msg.Reply != "" {
		fmt.Printf("Reply: %s\n", msg.Reply)
	}
	
	fmt.Printf("Timestamp: %s\n", msg.Timestamp.Format("2006-01-02 15:04:05.000"))
	fmt.Printf("Size: %d bytes\n", msg.Size)
	
	// Display headers if present
	if len(msg.Headers) > 0 {
		fmt.Printf("Headers:\n")
		for key, value := range msg.Headers {
			fmt.Printf("  %s: %s\n", key, value)
		}
	}
	
	fmt.Printf("Data:\n")
	
	// Try to pretty-print JSON data if it looks like JSON
	if len(msg.Data) > 0 {
		if msg.Data[0] == '{' || msg.Data[0] == '[' {
			// Attempt to format as JSON
			if formatted, err := utils.FormatJSON(msg.Data); err == nil {
				fmt.Printf("%s\n", formatted)
			} else {
				// Fall back to raw data if JSON formatting fails
				fmt.Printf("%s\n", string(msg.Data))
			}
		} else {
			// Display as raw string
			fmt.Printf("%s\n", string(msg.Data))
		}
	} else {
		fmt.Printf("(empty)\n")
	}
	
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")
}

// CreateMessageHandler creates a simple message handler that prints messages
func CreateMessageHandler() SubscriptionHandler {
	return func(msg *Message) error {
		DisplayMessage(msg)
		return nil
	}
}

// CreateJSONMessageHandler creates a handler that expects JSON messages
func CreateJSONMessageHandler() SubscriptionHandler {
	return func(msg *Message) error {
		// Try to validate and pretty-print JSON
		if formatted, err := utils.FormatJSON(msg.Data); err == nil {
			fmt.Printf("Subject: %s | JSON Message: %s\n", msg.Subject, formatted)
		} else {
			fmt.Printf("Subject: %s | Raw Message: %s\n", msg.Subject, string(msg.Data))
		}
		return nil
	}
}
