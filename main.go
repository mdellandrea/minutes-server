package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mdellandrea/minutes-server/lib/server"
	"github.com/rs/zerolog"
)

func main() {
	log := zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	s := server.Init(log)

	go func() {
		stopChan := make(chan os.Signal, 1)
		signal.Notify(stopChan, os.Interrupt, syscall.SIGTERM)
		<-stopChan

		log.Info().Msg("shutting down server")
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		if err := s.Shutdown(ctx); err != nil {
			log.Info().
				Err(err).
				Msg("http server error during shutdown")
		}
	}()

	if err := s.ListenAndServe(); err != http.ErrServerClosed {
		log.Info().
			Err(err).
			Msg("http server terminated unexpectedly")
	}
}
