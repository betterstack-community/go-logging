package logger

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var zeroLogger zerolog.Logger

// Get initializes a zerolog.Logger instance if it has not been initialized
// already and returns the same instance for subsequent calls.
func Get() zerolog.Logger {
	once.Do(func() {
		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel := zerolog.InfoLevel
		levelEnv := os.Getenv("LOG_LEVEL")
		if levelEnv != "" {
			levelFromEnv, err := zerolog.ParseLevel(levelEnv)
			if err != nil {
				log.Println(fmt.Errorf("invalid level from env defaulting to Info: %w", err))
			}

			logLevel = levelFromEnv
		}

		// Configure console logging in a human-friendly and colorized format
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
			FieldsExclude: []string{
				"user_agent",
				"git_revision",
				"go_version",
			},
		}

		// Configure an auto rotating file for storing JSON-formatted records
		fileWriter := &lumberjack.Logger{
			Filename:   "logs/app.log",
			MaxSize:    5,
			MaxBackups: 10,
			MaxAge:     14,
			Compress:   true,
		}

		// Allows logging to multiple destinations at once
		output := zerolog.MultiLevelWriter(consoleWriter, fileWriter)

		var gitRevision string

		buildInfo, ok := debug.ReadBuildInfo()
		if ok {
			for _, v := range buildInfo.Settings {
				if v.Key == "vcs.revision" {
					gitRevision = v.Value
					break
				}
			}
		}

		// Create a new logger with some global metadata
		zeroLogger = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Str("git_revision", gitRevision).
			Str("go_version", buildInfo.GoVersion).
			Logger()

		zerolog.DefaultContextLogger = &zeroLogger
	})

	return zeroLogger
}
