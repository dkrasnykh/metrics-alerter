package logger

import (
	"time"

	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

func New() (*Logger, error) {
	c := zap.NewProductionConfig()
	zapl, err := c.Build()
	if err != nil {
		return nil, err
	}
	return &Logger{zapl}, err
}

func (l *Logger) InfoRequest(method, uri string, duration time.Duration) {
	l.logger.Info("request",
		zap.String("method", method),
		zap.String("URI", uri),
		zap.Duration("duration", duration))
}

func (l *Logger) InfoResponse(statusCode, length int) {
	l.logger.Info("response",
		zap.Int("code", statusCode),
		zap.Int("length", length))
}
