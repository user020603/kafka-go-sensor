package models

import (
	"time"
)

// SensorData represents the structure of messages sent to the sensor-data topic
type SensorData struct {
	ID          string    `json:"id"`
	SensorType  string    `json:"sensor_type"`
	Value       float64   `json:"value"`
	Unit        string    `json:"unit"`
	Timestamp   time.Time `json:"timestamp"`
	DeviceID    string    `json:"device_id"`
	LocationID  string    `json:"location_id"`
}

// SystemLog represents the structure of messages sent to the system-logs topic
type SystemLog struct {
	ID        string    `json:"id"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}