package log

import (
	"os"
	"path/filepath"

	"github.com/bodenr/vehicle-api/config"
	"github.com/rs/zerolog"
)

var Log zerolog.Logger

func init() {
	level, err := zerolog.ParseLevel(config.GetEnv("LOG_LEVEL", "info"))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)
	binary := filepath.Base(os.Args[0])
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}
	//zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	Log = zerolog.New(os.Stdout).With().Timestamp().Caller().Str(
		Binary, binary).Str(Hostname, hostname).Logger()

	if err != nil {
		Log.Warn().Msg("failed to parse log level, falling back to info")
	}
}
