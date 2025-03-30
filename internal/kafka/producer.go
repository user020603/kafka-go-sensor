package kafka

import (
	"context"
	"encoding/json"
	"time"

	kafkago "github.com/segmentio/kafka-go" 
	"kafka-go-app/internal/config"
	"kafka-go-app/internal/utils"
)

type Producer struct {
	writers map[string]*kafkago.Writer
	config  *config.KafkaConfig
}

// NewProducer creates a new Kafka producer
func NewProducer(config *config.KafkaConfig) *Producer {
	writers := make(map[string]*kafkago.Writer)
	
	// Create writers for each topic
	writers[config.SensorDataTopic] = createWriter(config.Brokers, config.SensorDataTopic, config.ProducerBatchSize, config.ProducerTimeout)
	writers[config.SystemLogsTopic] = createWriter(config.Brokers, config.SystemLogsTopic, config.ProducerBatchSize, config.ProducerTimeout)
	
	return &Producer{
		writers: writers,
		config:  config,
	}
}

// createWriter creates a new Kafka writer for a given topic
func createWriter(brokers []string, topic string, batchSize int, timeout time.Duration) *kafkago.Writer {
	return &kafkago.Writer{
		Addr:         kafkago.TCP(brokers...),
		Topic:        topic,
		Balancer:     &kafkago.Hash{},
		BatchSize:    batchSize,
		BatchTimeout: timeout,
		Async:        true,
	}
}

// SendMessage sends a message to the specified topic with the given key
func (p *Producer) SendMessage(ctx context.Context, topic, key string, message interface{}) error {
	messageBytes, err := json.Marshal(message) 
	if err != nil {
		utils.Logger.WithError(err).Error("Failed to marshal message")
		return err
	}
	
	writer, ok := p.writers[topic]
	if !ok {
		utils.Logger.WithField("topic", topic).Error("Writer not found for topic")
		return err
	}
	
	err = writer.WriteMessages(ctx, kafkago.Message{
		Key:   []byte(key),
		Value: messageBytes,
		Time:  time.Now(),
	})
	
	if err != nil {
		utils.Logger.WithError(err).WithFields(map[string]interface{}{
			"topic": topic,
			"key":   key,
		}).Error("Failed to send message")
		return err
	}
	
	utils.Logger.WithFields(map[string]interface{}{
		"topic": topic,
		"key":   key,
	}).Debug("Message sent successfully")
	
	return nil
}

// Close closes all writers
func (p *Producer) Close() {
	for topic, writer := range p.writers {
		utils.Logger.WithField("topic", topic).Info("Closing writer")
		writer.Close()
	}
}