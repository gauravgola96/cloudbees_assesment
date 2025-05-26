package server

import (
	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"
	"net/http"
	"time"
)

var JENKINBASELOGURL = "https://ci.jenkins.io/job/Core/job/jenkins/job/master/"

func NewHttpServer() (*http.Server, error) {
	subLogger := log.With().Str("module", "server.server").Logger()
	router := chi.NewRouter()

	addr := "localhost:8080"
	subLogger.Info().Msgf("Starting server on %s", addr)
	httpServer := http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 20 * time.Second,
		ReadTimeout:       1 * time.Minute,
		WriteTimeout:      2 * time.Minute,
	}

	router.Mount("/", LogProxyRoutes())

	go func() {
		subLogger.Info().Msgf("Listening server on %s", addr)
		err := httpServer.ListenAndServe()
		if err != nil {
			subLogger.Error().Err(err).Msgf("Error in starting server")
			return
		}
	}()

	return &httpServer, nil
}
