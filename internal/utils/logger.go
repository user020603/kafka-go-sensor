package utils

import (
    "os"

    "github.com/sirupsen/logrus"
)

var Logger *logrus.Entry

// InitLogger initializes the logger with a consumer name
func InitLogger(consumerName string) {
    baseLogger := logrus.New()
    baseLogger.SetOutput(os.Stdout)
    baseLogger.SetFormatter(&logrus.TextFormatter{
        FullTimestamp: true,
    })

    // Set log level từ biến môi trường hoặc mặc định là info
    levelStr := os.Getenv("LOG_LEVEL")
    if levelStr == "" {
        levelStr = "info"
    }

    level, err := logrus.ParseLevel(levelStr)
    if err != nil {
        level = logrus.InfoLevel
    }

    baseLogger.SetLevel(level)

    // Add consumer_name as a default field
    Logger = baseLogger.WithField("consumer_name", consumerName)
}