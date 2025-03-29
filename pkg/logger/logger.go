package logger

import (
	"os"

	"github.com/rs/zerolog"
)

//go:generate mockgen -destination=../mocks/mock_logger.go -package=mocks github.com/Notifuse/notifuse/pkg/logger Logger

type Logger interface {
	Debug(msg string)
	Info(msg string)
	Warn(msg string)
	Error(msg string)
	Fatal(msg string)
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
}

type zerologLogger struct {
	logger zerolog.Logger
}

func NewLogger() Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	return &zerologLogger{
		logger: logger,
	}
}

func (l *zerologLogger) Debug(msg string) {
	l.logger.Debug().Msg(msg)
}

func (l *zerologLogger) Info(msg string) {
	l.logger.Info().Msg(msg)
}

func (l *zerologLogger) Warn(msg string) {
	l.logger.Warn().Msg(msg)
}

func (l *zerologLogger) Error(msg string) {
	l.logger.Error().Msg(msg)
}

func (l *zerologLogger) Fatal(msg string) {
	l.logger.Fatal().Msg(msg)
}

func (l *zerologLogger) WithField(key string, value interface{}) Logger {
	return &zerologLogger{
		logger: l.logger.With().Interface(key, value).Logger(),
	}
}

func (l *zerologLogger) WithFields(fields map[string]interface{}) Logger {
	for key, value := range fields {
		l.logger = l.logger.With().Interface(key, value).Logger()
	}
	return l
}
