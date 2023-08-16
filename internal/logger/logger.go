package logger

import (
	"log"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
	flog   *os.File
}

var Log Logger

func NewLogger(level string, logFile string) {
	zerolog.TimeFieldFormat = time.RFC3339

	switch level {
	case "debug":
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	case "info":
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	case "warn":
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	case "error":
		zerolog.SetGlobalLevel(zerolog.ErrorLevel)
	case "fatal":
		zerolog.SetGlobalLevel(zerolog.FatalLevel)
	case "panic":
		zerolog.SetGlobalLevel(zerolog.PanicLevel)
	default:
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var err error
	if logFile != "" {
		Log.flog, err = os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatal(err)
		}

		Log.logger = zerolog.New(Log.flog).With().Timestamp().Logger()
		return
	}

	Log.logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
}

func Close() error {
	if Log.flog != nil {
		err := Log.flog.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

// Debug logs a debug message.
func Debug(msg string) {
	Log.logger.Debug().Msg(msg)
}

// Info logs an info message.
func Info(msg string) {
	Log.logger.Info().Msg(msg)
}

// Info logs an info message.
func Infof(format string, v ...interface{}) {
	Log.logger.Info().Msgf(format, v...)
}

// Warn logs a warning message.
func Warn(msg string) {
	Log.logger.Warn().Msg(msg)
}

// Error logs an error message.
func Error(msg string, err error) {
	Log.logger.Err(err).Msg(msg)
}

// Fatal logs a fatal message and exits the program.
func Fatal(msg string, err error) {
	Log.logger.Fatal().Err(err).Msg(msg)
}

// Panic logs a panic message and panics.
func Panic(msg string, err error) {
	Log.logger.Panic().Err(err).Msg(msg)
}
