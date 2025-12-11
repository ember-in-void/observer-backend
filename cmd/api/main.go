// cmd/api/main.go
package main

import (
	"log"

	"steam-observer/internal/app"
	"steam-observer/internal/shared/config"
)

func main() {
	cfg := config.Load()

	server := app.NewServer(cfg)

	if err := server.Run(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
