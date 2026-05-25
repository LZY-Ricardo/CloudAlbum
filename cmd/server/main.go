package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/storage"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := initDB(cfg)
	if err != nil {
		log.Fatalf("Failed to init database: %v", err)
	}

	store, err := initStorage(cfg)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}

	fmt.Printf("CloudAlbum starting on :%d\n", cfg.Server.Port)
	fmt.Printf("Database: %s (%s)\n", cfg.Database.Driver, cfg.Database.DSN)
	if localStore, ok := store.(*storage.LocalStorage); ok {
		fmt.Printf("Storage: %s (%s)\n", cfg.Storage.Driver, localStore.String())
	} else {
		fmt.Printf("Storage: %s\n", cfg.Storage.Driver)
	}

	_ = db
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	switch cfg.Database.Driver {
	case "sqlite":
		if err := ensureParentDir(cfg.Database.DSN); err != nil {
			return nil, fmt.Errorf("prepare database path: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	db, err := gorm.Open(sqlite.Open(cfg.Database.DSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Image{}, &model.Album{}, &model.APIToken{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Driver {
	case "local":
		return storage.NewLocalStorage(cfg.Storage.Local.Path)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}
}

func ensureParentDir(path string) error {
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}
