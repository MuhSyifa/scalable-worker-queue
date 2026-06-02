package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func InitLogger(environment string) {
	// Set log level based on environment
	if environment == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		// Pretty print for development
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: zerolog.TimeFieldFormat})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		// JSON format for production
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	}
}
