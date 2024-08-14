package utils

import (
	"go.uber.org/zap"
	"os"
)

var sugar *zap.SugaredLogger
var logLevel string

// initializeLogger initializes and returns a logger instance.
func InitializeLogger() *zap.Logger {
	cfg := zap.NewProductionConfig()
	cfg.Level.UnmarshalText([]byte(LOG_LEVEL))
	logger, _ := cfg.Build()
	return logger
}

// init initializes the logger and sets the log level based on the environment variable.
func init() {
	sugar = InitializeLogger().Sugar()
	if os.Getenv(LOG_LEVEL) != "" {
		logLevel = os.Getenv(LOG_LEVEL)
	} else {
		logLevel = "info"
	}
}
