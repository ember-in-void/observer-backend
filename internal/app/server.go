// internal/app/server.go
package app

import (
	"net/http"

	"steam-observer/internal/shared/config"
	"steam-observer/internal/shared/http/middleware"
	"steam-observer/internal/shared/logger"
)

type Server struct {
	addr      string
	router    http.Handler
	logger    logger.Logger
	Container *Container
}

func NewServer(cfg *config.Config, log logger.Logger) *Server {
	mux := http.NewServeMux()
	container := NewContainer(cfg, log)

	RegisterRoutes(mux, container)

	// Оборачиваем в logging middleware
	handler := middleware.Logging(log.WithField("component", "http"))(mux)

	return &Server{
		addr:      cfg.HTTPAddr,
		router:    handler,
		logger:    log,
		Container: container,
	}
}

func (s *Server) Run() error {
	s.logger.Infof("starting http server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}
