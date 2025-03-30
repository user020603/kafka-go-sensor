package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// KafkaConfig holds the configuration for Kafka connection
type KafkaConfig struct {
	Brokers           []string
	SensorDataTopic   string
	SystemLogsTopic   string
	ConsumerGroup     string
	ProducerBatchSize int
	ProducerTimeout   time.Duration
}

// LoadConfig loads configuration from environment or .env file
func LoadConfig() (*KafkaConfig, error) {
	// Try to load .env file if exists
	_ = godotenv.Load()

	brokers := strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092,localhost:9093,localhost:9094"), ",")
	
	batchSize, err := strconv.Atoi(getEnv("PRODUCER_BATCH_SIZE", "100"))
	if err != nil {
		batchSize = 100
	}
	
	timeoutStr := getEnv("PRODUCER_TIMEOUT", "1s")
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		timeout = 1 * time.Second
	}

	return &KafkaConfig{
		Brokers:           brokers,
		SensorDataTopic:   getEnv("SENSOR_DATA_TOPIC", "sensor-data"),
		SystemLogsTopic:   getEnv("SYSTEM_LOGS_TOPIC", "system-logs"),
		ConsumerGroup:     getEnv("CONSUMER_GROUP", "data-processor"),
		ProducerBatchSize: batchSize,
		ProducerTimeout:   timeout,
	}, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}