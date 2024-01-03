package logger

import (
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

func InitLogger() error {
	if logger != nil {
		return nil
	}
	c := zap.NewProductionConfig()
	var err error
	logger, err = c.Build()
	return err
}

func InfoRequest(method, uri string, duration time.Duration) {
	logger.Info("request",
		zap.String("method", method),
		zap.String("URI", uri),
		zap.Duration("duration", duration))
}

func InfoResponse(statusCode, length int) {
	logger.Info("response",
		zap.Int("code", statusCode),
		zap.Int("length", length))
}

func Info(msg string) {
	logger.Info(msg)
}

func Error(msg string) {
	logger.Error(msg)
}

func Fatal(msg string) {
	logger.Fatal(msg)
}
