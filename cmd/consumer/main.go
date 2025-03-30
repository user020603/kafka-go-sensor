// cmd/consumer/main.go
package main

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"kafka-go-app/internal/config"
	"kafka-go-app/internal/kafka"
	"kafka-go-app/internal/models"
	"kafka-go-app/internal/utils"
)

func main() {

	consumerName := flag.String("name", "default-consumer", "Name of the consumer")
    flag.Parse()

	
   // Initialize logger với tên consumer
   utils.InitLogger(*consumerName)
   utils.Logger.Infof("Starting consumer with name: %s", *consumerName)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		utils.Logger.WithError(err).Fatal("Failed to load configuration")
	}

	// Define topics to consume
	topics := []string{cfg.SensorDataTopic, cfg.SystemLogsTopic}

	// Create consumer
	consumer := kafka.NewConsumer(cfg, topics)

	// Register handlers for each topic
	consumer.RegisterHandler(cfg.SensorDataTopic, handleSensorData)
	consumer.RegisterHandler(cfg.SystemLogsTopic, handleSystemLog)

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

	// Start consuming
	if err := consumer.Start(ctx); err != nil {
		utils.Logger.WithError(err).Fatal("Error starting consumer")
	}

	utils.Logger.Info("Consumer gracefully shut down")
}

func handleSensorData(topic string, key, value []byte) error {
	var data models.SensorData
	
	if err := json.Unmarshal(value, &data); err != nil {
		utils.Logger.WithError(err).Error("Failed to unmarshal sensor data")
		return err
	}
	
	utils.Logger.WithFields(map[string]interface{}{
		"topic":       topic,
		"device_id":   data.DeviceID,
		"sensor_type": data.SensorType,
		"value":       data.Value,
		"unit":        data.Unit,
		"location":    data.LocationID,
		"timestamp":   data.Timestamp,
	}).Info("Processed sensor data")
	
	// Here you would typically store the data in a database or perform other processing
	// For example:
	// db.InsertSensorData(data)
	
	return nil
}

func handleSystemLog(topic string, key, value []byte) error {
	var log models.SystemLog
	
	if err := json.Unmarshal(value, &log); err != nil {
		utils.Logger.WithError(err).Error("Failed to unmarshal system log")
		return err
	}
	
	// Use different log levels based on the severity of the system log
	logEntry := utils.Logger.WithFields(map[string]interface{}{
		"topic":     topic,
		"log_id":    log.ID,
		"service":   log.Service,
		"timestamp": log.Timestamp,
	})
	
	switch log.Level {
	case "info":
		logEntry.Info(log.Message)
	case "warning":
		logEntry.Warn(log.Message)
	case "error":
		logEntry.Error(log.Message)
	case "debug":
		logEntry.Debug(log.Message)
	case "critical":
		logEntry.WithField("ALERT", true).Error(log.Message)
	default:
		logEntry.Info(log.Message)
	}
	
	// Here you could write logs to a file, send alerts, etc.
	// For example:
	// if log.Level == "critical" {
	//     alertService.SendAlert(log)
	// }
	
	return nil
}