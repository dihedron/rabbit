package main

import (
	"context"
	"log/slog"
	"os"

	"github.com/dihedron/rabbit"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {

	slog.SetDefault(
		slog.New(
			slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
				Level:     slog.LevelDebug,
				AddSource: true,
			}),
		),
	)

	// Create rabbit instance
	r, err := setup()
	if err != nil {
		slog.Error("unable to setup rabbit", "error", err)
		os.Exit(1)
	}

	errChan := make(chan *rabbit.ConsumeError, 1)

	slog.Debug("starting error listener...")

	// Launch an error listener
	go func() {
		for err := range errChan {
			slog.Debug("received rabbit error", "error", err)
		}
	}()

	slog.Debug("running consumer...")

	// Run a consumer
	r.Consume(context.Background(), errChan, func(d amqp.Delivery) error {
		slog.Debug("received message", "headers", d.Headers, "body", d.Body)

		// Acknowledge the message
		if err := d.Ack(false); err != nil {
			slog.Error("error acknowledging message", "error", err)
		}

		return nil
	})
}

func setup() (*rabbit.Rabbit, error) {
	return rabbit.New(&rabbit.Options{
		URLs:      []string{"amqp://guest:guest@localhost:5672/"},
		Mode:      rabbit.Both,
		QueueName: "test-queue",
		Bindings: []rabbit.Binding{
			{
				ExchangeName:       "test-exchange",
				BindingKeys:        []string{"test-key"},
				ExchangeDeclare:    true,
				ExchangeType:       "topic",
				ExchangeDurable:    true,
				ExchangeAutoDelete: true,
			},
		},
		RetryReconnectSec: 1,
		QueueDurable:      true,
		QueueExclusive:    false,
		QueueAutoDelete:   true,
		QueueDeclare:      true,
		AutoAck:           false,
		ConsumerTag:       "rabbit-example",
	})
}
