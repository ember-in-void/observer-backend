// internal/app/server.go
package app

import (
	"log"
	"net/http"

	"steam-observer/internal/shared/config"
)

type Server struct {
	addr      string
	router    http.Handler
	Container *Container
}

func NewServer(cfg *config.Config) *Server {
	mux := http.NewServeMux()
	container := NewContainer(cfg)

	RegisterRoutes(mux, container)

	return &Server{
		addr:      cfg.HTTPAddr,
		router:    mux,
		Container: container,
	}
}

func (s *Server) Run() error {
	log.Printf("starting http server on %s", s.addr)
	return http.ListenAndServe(s.addr, s.router)
}
