// cmd/producer/main.go
package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"kafka-go-app/internal/config"
	"kafka-go-app/internal/kafka" // Our own kafka package
	"kafka-go-app/internal/models"
	"kafka-go-app/internal/utils"
)

var (
	sensorTypes = []string{"temperature", "humidity", "pressure", "light", "motion"}
	locations   = []string{"building1", "building2", "building3"}
	services    = []string{"auth", "api", "database", "cache", "worker"}
	logLevels   = []string{"info", "warning", "error", "debug", "critical"}
)

func main() {
	// Initialize logger
	utils.InitLogger("producer")
	utils.Logger.Info("Starting producer...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Initialize random seed
	rand.Seed(time.Now().UnixNano())

	// Create producer
	producer := kafka.NewProducer(cfg)
	defer producer.Close()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigCh
		utils.Logger.WithField("signal", sig.String()).Info("Received shutdown signal")
		cancel()
	}()

	// Start producing messages
	wg := &sync.WaitGroup{}
	wg.Add(2)

	// Generate sensor data
	go func() {
		defer wg.Done()
		produceSensorData(ctx, producer, cfg.SensorDataTopic)
	}()

	// Generate system logs
	go func() {
		defer wg.Done()
		produceSystemLogs(ctx, producer, cfg.SystemLogsTopic)
	}()

	wg.Wait()
	utils.Logger.Info("Producer gracefully shut down")
}

func produceSensorData(ctx context.Context, producer *kafka.Producer, topic string) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	deviceCount := 10
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Generate a sensor data message
			deviceID := fmt.Sprintf("device-%d", rand.Intn(deviceCount))
			sensorType := sensorTypes[rand.Intn(len(sensorTypes))]
			locationID := locations[rand.Intn(len(locations))]

			// Use device ID as key to ensure same device data goes to same partition
			key := deviceID

			data := models.SensorData{
				ID:         fmt.Sprintf("sensor-%s-%d", sensorType, time.Now().UnixNano()),
				SensorType: sensorType,
				Value:      float64(rand.Intn(100)) + rand.Float64(),
				Unit:       getUnitForSensorType(sensorType),
				Timestamp:  time.Now(),
				DeviceID:   deviceID,
				LocationID: locationID,
			}

			if err := producer.SendMessage(ctx, topic, key, data); err != nil {
				utils.Logger.WithError(err).Error("Failed to send sensor data")
			}
		}
	}
}

func produceSystemLogs(ctx context.Context, producer *kafka.Producer, topic string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Generate a system log message
			service := services[rand.Intn(len(services))]
			level := logLevels[rand.Intn(len(logLevels))]

			// Use service name as key for consistent routing
			key := service

			log := models.SystemLog{
				ID:        fmt.Sprintf("log-%d", time.Now().UnixNano()),
				Level:     level,
				Message:   generateLogMessage(service, level),
				Service:   service,
				Timestamp: time.Now(),
			}

			if err := producer.SendMessage(ctx, topic, key, log); err != nil {
				utils.Logger.WithError(err).Error("Failed to send system log")
			}
		}
	}
}

func getUnitForSensorType(sensorType string) string {
	switch sensorType {
	case "temperature":
		return "°C"
	case "humidity":
		return "%"
	case "pressure":
		return "hPa"
	case "light":
		return "lux"
	case "motion":
		return "boolean"
	default:
		return "unknown"
	}
}

func generateLogMessage(_, level string) string {
	messages := map[string][]string{
		"info": {
			"Service started successfully",
			"Request processed in %dms",
			"User %s logged in",
			"Cache hit ratio at %d%%",
		},
		"warning": {
			"High resource usage detected",
			"Request took longer than expected: %dms",
			"Rate limit approaching for client %s",
			"Cache miss rate increasing",
		},
		"error": {
			"Failed to connect to database",
			"Request failed with status %d",
			"User %s authentication failed",
			"Service dependency unreachable",
		},
		"debug": {
			"Processing request: %s",
			"Query execution time: %dms",
			"Cache state updated",
			"Worker pool size adjusted to %d",
		},
		"critical": {
			"Service is unresponsive",
			"Database connection pool exhausted",
			"Memory usage exceeded threshold",
			"Critical security alert: %s",
		},
	}

	templates := messages[level]
	template := templates[rand.Intn(len(templates))]

	switch {
	case strings.Contains(template, "%d"):
		return fmt.Sprintf(template, rand.Intn(1000))
	case strings.Contains(template, "%s"):
		users := []string{"user123", "admin", "system", "guest", "api-client"}
		return fmt.Sprintf(template, users[rand.Intn(len(users))])
	default:
		return template
	}
}
