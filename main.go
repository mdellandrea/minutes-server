package main

import (
	"os"

	"github.com/mdellandrea/minutes-server/lib/server"
	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()
	s := server.Init(logger)
	if err := s.ListenAndServe(); err != nil {
		logger.Fatal().Err(err)
	}
}
