package mq

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"go.uber.org/zap"
)

// MessageQueue defines the interface for message queue operations
type MessageQueue interface {
	Publish(ctx context.Context, topic string, message interface{}) error
	Subscribe(topic string, handler MessageHandler) error
	Close() error
}

// MessageHandler is a function that processes messages
type MessageHandler func(ctx context.Context, message []byte) error

// RabbitMQ implements MessageQueue interface using RabbitMQ
type RabbitMQ struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	logger       *zap.Logger
	url          string
	exchangeName string
	mu           sync.RWMutex
	closed       bool
	reconnecting bool
}

// Config holds RabbitMQ configuration
type Config struct {
	URL          string
	ExchangeName string
	Logger       *zap.Logger
}

// NewRabbitMQ creates a new RabbitMQ instance
func NewRabbitMQ(config *Config) (*RabbitMQ, error) {
	if config.Logger == nil {
		config.Logger = zap.NewNop()
	}

	if config.ExchangeName == "" {
		config.ExchangeName = "airy.events"
	}

	mq := &RabbitMQ{
		url:          config.URL,
		exchangeName: config.ExchangeName,
		logger:       config.Logger,
	}

	if err := mq.connect(); err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	return mq, nil
}

// connect establishes connection to RabbitMQ
func (mq *RabbitMQ) connect() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.closed {
		return fmt.Errorf("message queue is closed")
	}

	conn, err := amqp.Dial(mq.url)
	if err != nil {
		return fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	channel, err := conn.Channel()
	if err != nil {
		conn.Close()
		return fmt.Errorf("failed to open channel: %w", err)
	}

	// Declare exchange
	err = channel.ExchangeDeclare(
		mq.exchangeName, // name
		"topic",         // type
		true,            // durable
		false,           // auto-deleted
		false,           // internal
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		channel.Close()
		conn.Close()
		return fmt.Errorf("failed to declare exchange: %w", err)
	}

	mq.conn = conn
	mq.channel = channel

	mq.logger.Info("connected to RabbitMQ",
		zap.String("exchange", mq.exchangeName))

	// Set up connection close notification
	go mq.handleConnectionClose()

	return nil
}

// handleConnectionClose handles connection close events and attempts reconnection
func (mq *RabbitMQ) handleConnectionClose() {
	closeErr := <-mq.conn.NotifyClose(make(chan *amqp.Error))
	if closeErr != nil {
		mq.logger.Error("RabbitMQ connection closed", zap.Error(closeErr))
	}

	mq.mu.Lock()
	if mq.closed {
		mq.mu.Unlock()
		return
	}
	mq.reconnecting = true
	mq.mu.Unlock()

	// Attempt to reconnect
	for i := 0; i < 10; i++ {
		mq.logger.Info("attempting to reconnect to RabbitMQ",
			zap.Int("attempt", i+1))

		time.Sleep(time.Duration(i+1) * time.Second)

		if err := mq.connect(); err != nil {
			mq.logger.Error("failed to reconnect", zap.Error(err))
			continue
		}

		mq.mu.Lock()
		mq.reconnecting = false
		mq.mu.Unlock()

		mq.logger.Info("successfully reconnected to RabbitMQ")
		return
	}

	mq.logger.Error("failed to reconnect to RabbitMQ after multiple attempts")
}

// Publish publishes a message to the specified topic
func (mq *RabbitMQ) Publish(ctx context.Context, topic string, message interface{}) error {
	mq.mu.RLock()
	if mq.closed {
		mq.mu.RUnlock()
		return fmt.Errorf("message queue is closed")
	}
	if mq.reconnecting {
		mq.mu.RUnlock()
		return fmt.Errorf("message queue is reconnecting")
	}
	channel := mq.channel
	mq.mu.RUnlock()

	// Marshal message to JSON
	body, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish message
	err = channel.PublishWithContext(
		ctx,
		mq.exchangeName, // exchange
		topic,           // routing key
		false,           // mandatory
		false,           // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
			Timestamp:    time.Now(),
		},
	)
	if err != nil {
		return fmt.Errorf("failed to publish message: %w", err)
	}

	mq.logger.Debug("published message",
		zap.String("topic", topic),
		zap.Int("size", len(body)))

	return nil
}

// Subscribe subscribes to messages on the specified topic
func (mq *RabbitMQ) Subscribe(topic string, handler MessageHandler) error {
	mq.mu.RLock()
	if mq.closed {
		mq.mu.RUnlock()
		return fmt.Errorf("message queue is closed")
	}
	channel := mq.channel
	mq.mu.RUnlock()

	// Declare queue
	queue, err := channel.QueueDeclare(
		"",    // name (empty for auto-generated)
		false, // durable
		true,  // delete when unused
		true,  // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to declare queue: %w", err)
	}

	// Bind queue to exchange with routing key
	err = channel.QueueBind(
		queue.Name,      // queue name
		topic,           // routing key
		mq.exchangeName, // exchange
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		return fmt.Errorf("failed to bind queue: %w", err)
	}

	// Start consuming messages
	msgs, err := channel.Consume(
		queue.Name, // queue
		"",         // consumer
		false,      // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // args
	)
	if err != nil {
		return fmt.Errorf("failed to register consumer: %w", err)
	}

	mq.logger.Info("subscribed to topic",
		zap.String("topic", topic),
		zap.String("queue", queue.Name))

	// Process messages in a goroutine
	go func() {
		for msg := range msgs {
			ctx := context.Background()

			mq.logger.Debug("received message",
				zap.String("topic", topic),
				zap.Int("size", len(msg.Body)))

			// Call handler
			if err := handler(ctx, msg.Body); err != nil {
				mq.logger.Error("failed to handle message",
					zap.String("topic", topic),
					zap.Error(err))
				// Reject message and requeue
				msg.Nack(false, true)
			} else {
				// Acknowledge message
				msg.Ack(false)
			}
		}
	}()

	return nil
}

// Close closes the RabbitMQ connection
func (mq *RabbitMQ) Close() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.closed {
		return nil
	}

	mq.closed = true

	var errs []error

	if mq.channel != nil {
		if err := mq.channel.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close channel: %w", err))
		}
	}

	if mq.conn != nil {
		if err := mq.conn.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors closing RabbitMQ: %v", errs)
	}

	mq.logger.Info("closed RabbitMQ connection")
	return nil
}
