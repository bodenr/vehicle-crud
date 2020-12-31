// Package log provides logging for the application.
package log

import (
	"os"
	"path/filepath"

	"github.com/bodenr/vehicle-api/config"
	"github.com/rs/zerolog"
)

// Log is the singleton application wide logger.
var Log zerolog.Logger

func init() {
	// setup the singleton logger
	level, err := zerolog.ParseLevel(config.GetEnv("LOG_LEVEL", "info"))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	binary := filepath.Base(os.Args[0])
	hostname, hErr := os.Hostname()
	if hErr != nil {
		hostname = "unknown"
	}
	Log = zerolog.New(os.Stdout).With().Timestamp().Caller().Str(
		Binary, binary).Str(Hostname, hostname).Logger()

	if err != nil {
		Log.Warn().Msg("failed to parse log level, falling back to info")
	}
}
