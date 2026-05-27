package repository

import (
	"fmt"
	"testing"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newSettingsTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.Settings{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func TestSettingsRepoLoadOrBootstrapEmpty(t *testing.T) {
	db := newSettingsTestDB(t)
	repo := NewSettingsRepository(db)
	o, err := repo.LoadOrBootstrap()
	if err != nil {
		t.Fatalf("bootstrap: %v", err)
	}
	if o.Server.BaseURL != nil || o.Image.Quality != nil {
		t.Fatalf("expected zero overrides, got %#v", o)
	}
	var row model.Settings
	if err := db.First(&row, 1).Error; err != nil {
		t.Fatalf("row not seeded: %v", err)
	}
	if row.Payload != "{}" && row.Payload != `{"server":{},"image":{}}` {
		t.Fatalf("payload not empty: %s", row.Payload)
	}
}

func TestSettingsRepoSaveAndLoad(t *testing.T) {
	db := newSettingsTestDB(t)
	repo := NewSettingsRepository(db)
	_, _ = repo.LoadOrBootstrap()

	var o config.Overrides
	q := 90
	o.Image.Quality = &q
	if err := repo.Save(o, 1); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := repo.LoadOrBootstrap()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Image.Quality == nil || *got.Image.Quality != 90 {
		t.Fatalf("reload quality: %#v", got.Image.Quality)
	}
}

func TestSettingsRepoCorruptedPayload(t *testing.T) {
	db := newSettingsTestDB(t)
	if err := db.Create(&model.Settings{ID: 1, Payload: "not json"}).Error; err != nil {
		t.Fatalf("seed: %v", err)
	}
	repo := NewSettingsRepository(db)
	_, err := repo.LoadOrBootstrap()
	if err == nil {
		t.Fatalf("expected error for corrupted payload")
	}
}
