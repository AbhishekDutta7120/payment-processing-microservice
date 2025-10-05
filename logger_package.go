package logger

import (
	"go.uber.org/zap"
)

type Logger interface {
	Info(msg string)
	Error(msg string)
	Warn(msg string)
	Debug(msg string)
}

type zapLogger struct {
	logger *zap.Logger
}

func NewLogger() Logger {
	logger, _ := zap.NewProduction()
	return &zapLogger{logger: logger}
}

func (l *zapLogger) Info(msg string) {
	l.logger.Info(msg)
}

func (l *zapLogger) Error(msg string) {
	l.logger.Error(msg)
}

func (l *zapLogger) Warn(msg string) {
	l.logger.Warn(msg)
}

func (l *zapLogger) Debug(msg string) {
	l.logger.Debug(msg)
}