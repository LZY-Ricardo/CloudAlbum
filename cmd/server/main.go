package main

import (
	"fmt"
	"log"

	"cloudalbum/internal/config"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	fmt.Printf("CloudAlbum starting on :%d\n", cfg.Server.Port)
	fmt.Printf("Database: %s (%s)\n", cfg.Database.Driver, cfg.Database.DSN)
	fmt.Printf("Storage: %s\n", cfg.Storage.Driver)
}
