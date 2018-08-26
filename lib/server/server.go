package server

import (
	"net/http"
	"time"

	"github.com/mdellandrea/minutes-server/lib/backend"
	"github.com/mdellandrea/minutes-server/lib/handlers"

	"github.com/go-chi/chi"
	"github.com/kelseyhightower/envconfig"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/hlog"
)

type serverConfig struct {
	ListenPort string `envconfig:"PORT0" default:":8080"`
	DbHost     string `envconfig:"DBHOST" default:"127.0.0.1"`
	DbPort     string `envconfig:"DBPORT" default:"6379"`
	Debug      bool   `envconfig:"DEBUG" default:"false"`
}

func setupMiddleware(log zerolog.Logger, mux *chi.Mux) {
	mux.Use(hlog.NewHandler(log))
	mux.Use(hlog.AccessHandler(func(r *http.Request, status, size int, duration time.Duration) {
		hlog.FromRequest(r).Info().
			Str("method", r.Method).
			Str("proto", r.Proto).
			Str("url", r.URL.String()).
			Int("status", status).
			Int("size", size).
			Dur("duration", duration).
			Msg("")
	}))
	mux.Use(hlog.RemoteAddrHandler("ip"))
	mux.Use(hlog.UserAgentHandler("user_agent"))
	mux.Use(hlog.RefererHandler("referer"))
	mux.Use(hlog.RequestIDHandler("req_id", "Request-Id"))
}

func Init(log zerolog.Logger) *http.Server {
	var c serverConfig
	err := envconfig.Process("", &c)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("environment variable configuration")
	}

	if c.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	client, err := backend.NewBackend(c.DbHost, c.DbPort)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("backend failure")
	}

	mux := chi.NewMux()
	setupMiddleware(log, mux)
	router := handlers.SetupRoutes(mux, client, log)

	return &http.Server{
		Addr:    c.ListenPort,
		Handler: router,
	}
}
