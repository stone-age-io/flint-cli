package nats

import (
	"fmt"
	"time"

	"github.com/nats-io/nats.go"
	"flint-cli/internal/utils"
)

// Publish publishes a message to a NATS subject
func (c *Client) Publish(subject string, data []byte, headers map[string]string) error {
	return c.executeWithConnection(func() error {
		return c.publishMessage(subject, data, headers, "")
	})
}

// PublishWithReply publishes a message with a reply subject
func (c *Client) PublishWithReply(subject, reply string, data []byte, headers map[string]string) error {
	return c.executeWithConnection(func() error {
		return c.publishMessage(subject, data, headers, reply)
	})
}

// Request performs a request-reply operation
func (c *Client) Request(subject string, data []byte, timeout time.Duration, headers map[string]string) (*Message, error) {
	var response *Message
	
	err := c.executeWithConnection(func() error {
		// Create the request message
		msg := nats.NewMsg(subject)
		msg.Data = data
		
		// Add headers if provided
		if len(headers) > 0 {
			if msg.Header == nil {
				msg.Header = make(nats.Header)
			}
			for key, value := range headers {
				msg.Header.Set(key, value)
			}
		}
		
		utils.PrintDebug(fmt.Sprintf("Sending NATS request to subject: %s (timeout: %v)", subject, timeout))
		
		// Send the request
		resp, err := c.conn.RequestMsg(msg, timeout)
		if err != nil {
			return WrapNATSError("request", subject, err)
		}
		
		// Convert NATS message to our Message format
		response = convertNATSMessage(resp)
		
		utils.PrintDebug(fmt.Sprintf("Received NATS response from subject: %s (size: %d bytes)", 
			resp.Subject, len(resp.Data)))
		
		return nil
	})
	
	return response, err
}

// PublishAsync publishes a message asynchronously
func (c *Client) PublishAsync(subject string, data []byte, headers map[string]string) error {
	return c.executeWithConnection(func() error {
		// Validate inputs
		if err := utils.ValidateNATSSubject(subject); err != nil {
			return WrapNATSError("publish", subject, err)
		}
		
		if err := utils.ValidateNATSMessage(data, headers); err != nil {
			return WrapNATSError("publish", subject, err)
		}
		
		// Create the message
		msg := nats.NewMsg(subject)
		msg.Data = data
		
		// Add headers if provided
		if len(headers) > 0 {
			if msg.Header == nil {
				msg.Header = make(nats.Header)
			}
			for key, value := range headers {
				msg.Header.Set(key, value)
			}
		}
		
		utils.PrintDebug(fmt.Sprintf("Publishing async message to subject: %s (size: %d bytes)", 
			subject, len(data)))
		
		// Publish asynchronously
		if err := c.conn.PublishMsg(msg); err != nil {
			return WrapNATSError("publish", subject, err)
		}
		
		return nil
	})
}

// PublishJSON publishes a JSON message (convenience method)
func (c *Client) PublishJSON(subject string, data interface{}, headers map[string]string) error {
	return c.executeWithConnection(func() error {
		// Convert data to JSON
		jsonData, err := utils.ToJSON(data)
		if err != nil {
			return WrapNATSError("publish", subject, fmt.Errorf("failed to marshal JSON: %w", err))
		}
		
		// Set content type header
		if headers == nil {
			headers = make(map[string]string)
		}
		headers["Content-Type"] = "application/json"
		
		return c.publishMessage(subject, jsonData, headers, "")
	})
}

// publishMessage is the internal method that handles the actual publishing
func (c *Client) publishMessage(subject string, data []byte, headers map[string]string, reply string) error {
	// Validate inputs
	if err := utils.ValidateNATSSubject(subject); err != nil {
		return WrapNATSError("publish", subject, err)
	}
	
	if err := utils.ValidateNATSMessage(data, headers); err != nil {
		return WrapNATSError("publish", subject, err)
	}
	
	// Create the message
	msg := nats.NewMsg(subject)
	msg.Data = data
	
	if reply != "" {
		msg.Reply = reply
	}
	
	// Add headers if provided
	if len(headers) > 0 {
		if msg.Header == nil {
			msg.Header = make(nats.Header)
		}
		for key, value := range headers {
			msg.Header.Set(key, value)
		}
	}
	
	utils.PrintDebug(fmt.Sprintf("Publishing message to subject: %s (size: %d bytes, headers: %d)", 
		subject, len(data), len(headers)))
	
	// Publish the message
	if err := c.conn.PublishMsg(msg); err != nil {
		return WrapNATSError("publish", subject, err)
	}
	
	// Flush to ensure message is sent immediately for CLI operations
	if err := c.conn.Flush(); err != nil {
		utils.PrintWarning(fmt.Sprintf("Failed to flush NATS connection: %v", err))
		// Don't fail the operation for flush errors
	}
	
	utils.PrintDebug(fmt.Sprintf("Successfully published message to subject: %s", subject))
	return nil
}

// GetPublishStats returns publishing statistics
func (c *Client) GetPublishStats() *MessageStats {
	if c.conn == nil {
		return &MessageStats{}
	}
	
	stats := c.conn.Stats()
	return &MessageStats{
		Published: stats.OutMsgs,
		Received:  stats.InMsgs,
		// Note: NATS doesn't track errors in stats, we'd need to implement our own counter
		LastMessage: time.Now(), // This would need to be tracked separately
	}
}

// ValidatePublishOptions validates publish operation parameters
func ValidatePublishOptions(opts *PublishOptions) error {
	if opts == nil {
		return fmt.Errorf("publish options cannot be nil")
	}
	
	if err := utils.ValidateNATSSubject(opts.Subject); err != nil {
		return fmt.Errorf("invalid subject: %w", err)
	}
	
	if err := utils.ValidateNATSMessage(opts.Data, opts.Headers); err != nil {
		return fmt.Errorf("invalid message: %w", err)
	}
	
	if opts.Reply != "" {
		if err := utils.ValidateNATSSubject(opts.Reply); err != nil {
			return fmt.Errorf("invalid reply subject: %w", err)
		}
	}
	
	return nil
}

// FormatPublishSummary formats a summary of the publish operation for display
func FormatPublishSummary(subject string, dataSize int, headers map[string]string) string {
	summary := fmt.Sprintf("Published to %s (%d bytes)", subject, dataSize)
	
	if len(headers) > 0 {
		summary += fmt.Sprintf(" with %d header(s)", len(headers))
	}
	
	return summary
}

// convertNATSMessage converts a NATS message to our Message format
func convertNATSMessage(msg *nats.Msg) *Message {
	message := &Message{
		Subject:   msg.Subject,
		Reply:     msg.Reply,
		Data:      msg.Data,
		Timestamp: time.Now(),
		Size:      len(msg.Data),
	}
	
	// Convert headers
	if msg.Header != nil {
		message.Headers = make(map[string]string)
		for key, values := range msg.Header {
			if len(values) > 0 {
				message.Headers[key] = values[0] // Take first value if multiple
			}
		}
	}
	
	return message
}
