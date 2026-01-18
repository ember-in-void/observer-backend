// cmd/api/main.go
package main

import (
	"steam-observer/internal/app"
	"steam-observer/internal/shared/config"
	"steam-observer/internal/shared/logger"
)

func main() {
	log, err := logger.NewCustomLogger()
	if err != nil {
		panic("failed to initialize logger: " + err.Error())
	}

	log.Info("starting steam-observer API...")

	cfg := config.Load()
	log.Infof("config loaded, http_addr=%s", cfg.HTTPAddr)

	server := app.NewServer(cfg, log)

	defer func() {
		log.Info("shutting down...")
		server.Container.DB.Close()
	}()

	if err := server.Run(); err != nil {
		log.Errorf("server error: %v", err)
	}
}
