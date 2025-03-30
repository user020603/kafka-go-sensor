// internal/kafka/consumer.go
package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	kafkago "github.com/segmentio/kafka-go"  
	"kafka-go-app/internal/config"
	"kafka-go-app/internal/utils"
)

// MessageHandler is a function type for handling received messages
type MessageHandler func(topic string, key, value []byte) error

// Consumer encapsulates Kafka consumer functionality
type Consumer struct {
	readers  []*kafkago.Reader
	config   *config.KafkaConfig
	handlers map[string]MessageHandler
}

// NewConsumer creates a new Kafka consumer
func NewConsumer(config *config.KafkaConfig, topics []string) *Consumer {
	var readers []*kafkago.Reader
	
	for _, topic := range topics {
		reader := createReader(config.Brokers, topic, config.ConsumerGroup)
		readers = append(readers, reader)
	}
	
	return &Consumer{
		readers:  readers,
		config:   config,
		handlers: make(map[string]MessageHandler),
	}
}

// createReader creates a new Kafka reader for a given topic
func createReader(brokers []string, topic, groupID string) *kafkago.Reader {
	return kafkago.NewReader(kafkago.ReaderConfig{
		Brokers:     brokers,
		Topic:       topic,
		GroupID:     groupID,
		StartOffset: kafkago.FirstOffset,
		MinBytes:    10e3, // 10KB
		MaxBytes:    10e6, // 10MB
	})
}

// RegisterHandler registers a message handler for a specific topic
func (c *Consumer) RegisterHandler(topic string, handler MessageHandler) {
	c.handlers[topic] = handler
}

// Start starts consuming messages from all readers
func (c *Consumer) Start(ctx context.Context) error {
	wg := &sync.WaitGroup{}
	
	for _, reader := range c.readers {
		wg.Add(1)
		go c.consume(ctx, reader, wg)
	}
	
	wg.Wait()
	return nil
}

// consume processes messages from a specific reader
func (c *Consumer) consume(ctx context.Context, reader *kafkago.Reader, wg *sync.WaitGroup) {
	defer wg.Done()
	defer reader.Close()
	
	utils.Logger.WithField("topic", reader.Config().Topic).Info("Starting consumption")
	
	for {
		select {
		case <-ctx.Done():
			utils.Logger.WithField("topic", reader.Config().Topic).Info("Context cancelled, stopping consumer")
			return
		default:
			message, err := reader.ReadMessage(ctx)
			if err != nil {
				utils.Logger.WithError(err).WithField("topic", reader.Config().Topic).Error("Error reading message")
				continue
			}
			
			utils.Logger.WithFields(map[string]interface{}{
				"topic":     message.Topic,
				"partition": message.Partition,
				"offset":    message.Offset,
				"key":       string(message.Key),
			}).Debug("Received message")
			
			// Process message with registered handler
			handler, ok := c.handlers[message.Topic]
			if !ok {
				utils.Logger.WithField("topic", message.Topic).Warn("No handler registered for topic")
				continue
			}
			
			if err := handler(message.Topic, message.Key, message.Value); err != nil {
				utils.Logger.WithError(err).WithField("topic", message.Topic).Error("Error handling message")
			}
		}
	}
}

func PrintMessageAsJSON(topic string, key, value []byte) error {
	var data interface{}
	
	if err := json.Unmarshal(value, &data); err != nil {
		return fmt.Errorf("failed to unmarshal JSON: %w", err)
	}
	
	keyStr := string(key)
	utils.Logger.WithFields(map[string]interface{}{
		"topic": topic,
		"key":   keyStr,
		"data":  data,
	}).Info("Processed message")
	
	return nil
}