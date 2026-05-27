package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"cloudalbum/internal/config"
	"cloudalbum/internal/handler"
	imgpkg "cloudalbum/internal/image"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/router"
	"cloudalbum/internal/service"
	"cloudalbum/internal/storage"
	"github.com/gin-gonic/gin"
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

	userRepo := repository.NewUserRepository(db)
	imageRepo := repository.NewImageRepository(db)
	albumRepo := repository.NewAlbumRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	settingsRepo := repository.NewSettingsRepository(db)

	overrides, err := settingsRepo.LoadOrBootstrap()
	if err != nil {
		log.Printf("[config] WARNING: settings load failed, falling back to empty overrides: %v", err)
		overrides = config.Overrides{}
	}
	provider := config.NewProvider(*cfg, overrides)

	if info, statErr := os.Stat("configs/config.yaml"); statErr == nil {
		if updated, uErr := settingsRepo.UpdatedAt(); uErr == nil && info.ModTime().After(updated) {
			log.Printf("[config] yaml file is newer than settings table (yaml=%s settings=%s); runtime still uses values from settings DB. To re-seed from YAML, delete the row in settings table and restart.",
				info.ModTime().Format("2006-01-02T15:04:05Z07:00"),
				updated.Format("2006-01-02T15:04:05Z07:00"))
		}
	}
	authSvc := service.NewAuthService(userRepo, tokenRepo, provider)
	tokenSvc := service.NewTokenService(tokenRepo)
	processor := imgpkg.NewProcessor(provider)
	imageSvc := service.NewImageService(imageRepo, store, processor, provider)
	albumSvc := service.NewAlbumService(albumRepo, imageRepo)

	authHandler := handler.NewAuthHandler(authSvc)
	tokenHandler := handler.NewTokenHandler(tokenSvc)
	imageHandler := handler.NewImageHandler(imageSvc)
	albumHandler := handler.NewAlbumHandler(albumSvc)
	publicHandler := handler.NewPublicHandler(store, processor)

	if err := authSvc.EnsureAdmin("admin", "admin123"); err != nil {
		log.Fatalf("Failed to ensure admin user: %v", err)
	}

	r := gin.Default()
	router.Setup(r, WebFS(), authSvc, tokenSvc, authHandler, tokenHandler, imageHandler, albumHandler, publicHandler)

	fmt.Printf("CloudAlbum starting on :%d\n", cfg.Server.Port)
	fmt.Printf("Database: %s (%s)\n", cfg.Database.Driver, cfg.Database.DSN)
	if localStore, ok := store.(*storage.LocalStorage); ok {
		fmt.Printf("Storage: %s (%s)\n", cfg.Storage.Driver, localStore.String())
	} else {
		fmt.Printf("Storage: %s\n", cfg.Storage.Driver)
	}

	if err := r.Run(fmt.Sprintf(":%d", cfg.Server.Port)); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func initDB(cfg *config.Config) (*gorm.DB, error) {
	switch cfg.Database.Driver {
	case "sqlite":
		dbPath, err := sqliteFilesystemPath(cfg.Database.DSN)
		if err != nil {
			return nil, fmt.Errorf("parse sqlite dsn: %w", err)
		}
		if err := ensureParentDir(dbPath); err != nil {
			return nil, fmt.Errorf("prepare database path: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", cfg.Database.Driver)
	}

	dsn := cfg.Database.DSN
	if cfg.Database.Driver == "sqlite" {
		dsn = sqliteDSNWithPragmas(cfg.Database.DSN)
	}

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.Image{}, &model.Album{}, &model.APIToken{}, &model.Settings{}); err != nil {
		return nil, fmt.Errorf("auto migrate: %w", err)
	}

	return db, nil
}

func initStorage(cfg *config.Config) (storage.Storage, error) {
	switch cfg.Storage.Driver {
	case "local":
		return storage.NewLocalStorage(cfg.Storage.Local.Path)
	case "s3":
		return storage.NewS3Storage(cfg.Storage.S3)
	default:
		return nil, fmt.Errorf("unsupported storage driver: %s", cfg.Storage.Driver)
	}
}

func ensureParentDir(path string) error {
	if path == "" || path == ":memory:" {
		return nil
	}
	dir := filepath.Dir(path)
	if dir == "." || dir == "" {
		return nil
	}
	return os.MkdirAll(dir, 0755)
}

func sqliteFilesystemPath(dsn string) (string, error) {
	if dsn == "" || dsn == ":memory:" {
		return "", nil
	}
	if strings.HasPrefix(dsn, "file:") {
		u, err := url.Parse(dsn)
		if err != nil {
			return "", err
		}
		path := u.Path
		if path == "" {
			path = strings.TrimPrefix(dsn, "file:")
			if idx := strings.Index(path, "?"); idx >= 0 {
				path = path[:idx]
			}
		}
		if path == "" || path == ":memory:" || strings.Contains(path, "mode=memory") {
			return "", nil
		}
		return path, nil
	}
	if idx := strings.Index(dsn, "?"); idx >= 0 {
		return dsn[:idx], nil
	}
	return dsn, nil
}

func sqliteDSNWithPragmas(dsn string) string {
	if dsn == "" {
		return dsn
	}

	if strings.HasPrefix(dsn, "file:") {
		return appendSQLitePragmas(dsn)
	}

	if strings.Contains(dsn, "?") {
		return appendSQLitePragmas("file:" + dsn)
	}

	return appendSQLitePragmas(dsn)
}

func appendSQLitePragmas(dsn string) string {
	separator := "?"
	if strings.Contains(dsn, "?") {
		separator = "&"
	}

	values := url.Values{}
	values.Set("_foreign_keys", "on")
	values.Set("_busy_timeout", "5000")
	values.Set("_journal_mode", "WAL")

	return dsn + separator + values.Encode()
}
