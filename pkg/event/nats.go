package event

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

type Client struct {
	nc *nats.Conn
	js jetstream.JetStream
}

func NewClient(url string) (*Client, error) {
	nc, err := nats.Connect(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to nats: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream context: %w", err)
	}

	return &Client{
		nc: nc,
		js: js,
	}, nil
}

func (c *Client) Close() {
	if c.nc != nil {
		c.nc.Close()
	}
}

func (c *Client) Publish(ctx context.Context, subject string, data interface{}) error {
	payload, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	_, err = c.js.Publish(ctx, subject, payload)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	return nil
}

func (c *Client) Subscribe(ctx context.Context, streamName, subject, consumerName string, handler func(ctx context.Context, data []byte) error) error {
	// Ensure stream exists
	_, err := c.js.CreateOrUpdateStream(ctx, jetstream.StreamConfig{
		Name:     streamName,
		Subjects: []string{subject},
	})
	if err != nil {
		return fmt.Errorf("failed to create stream: %w", err)
	}

	// Create consumer
	cons, err := c.js.CreateOrUpdateConsumer(ctx, streamName, jetstream.ConsumerConfig{
		Durable:   consumerName,
		AckPolicy: jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	// Consume messages
	go func() {
		iter, err := cons.Messages()
		if err != nil {
			// In a real app, you might want to log this or handle it better
			return
		}
		defer iter.Stop()

		for {
			msg, err := iter.Next()
			if err != nil {
				// Handle error or exit loop
				return
			}

			if err := handler(ctx, msg.Data()); err != nil {
				// Handle processing error (maybe nack?)
				msg.Nak()
				continue
			}
			msg.Ack()
		}
	}()

	return nil
}
