package logger

import (
	"os"

	log "github.com/sirupsen/logrus"
)

// Init initializes the logger with a professional configuration
func Init() {
	// Use text formatter with colors and full timestamp
	log.SetFormatter(&log.TextFormatter{
		ForceColors:   true,
		FullTimestamp: true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// Output to stdout
	log.SetOutput(os.Stdout)

	// Default level (can be changed dynamically)
	log.SetLevel(log.InfoLevel)
}

// GetLogger returns the logrus logger instance
func GetLogger() *log.Logger {
	return log.StandardLogger()
}

// Convenience wrappers for structured logging
func Info(msg string, fields log.Fields) {
	if fields != nil {
		log.WithFields(fields).Info(msg)
	} else {
		log.Info(msg)
	}
}

func Warn(msg string, fields log.Fields) {
	if fields != nil {
		log.WithFields(fields).Warn(msg)
	} else {
		log.Warn(msg)
	}
}

func Error(msg string, fields log.Fields) {
	if fields != nil {
		log.WithFields(fields).Error(msg)
	} else {
		log.Error(msg)
	}
}

func Debug(msg string, fields log.Fields) {
	if fields != nil {
		log.WithFields(fields).Debug(msg)
	} else {
		log.Debug(msg)
	}
}

