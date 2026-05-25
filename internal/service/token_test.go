package service

import (
	"strings"
	"testing"

	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTokenServiceCreateAndValidate(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	svc := NewTokenService(tokenRepo)

	created, rawToken, err := svc.Create(7, "cli", "upload")
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.UserID != 7 {
		t.Fatalf("Create() user id = %d, want 7", created.UserID)
	}
	if created.Scope != "upload" {
		t.Fatalf("Create() scope = %q, want upload", created.Scope)
	}
	if strings.HasPrefix(created.TokenHash, "ca_") {
		t.Fatalf("Create() stored token hash leaked raw prefix: %q", created.TokenHash)
	}
	if len(created.TokenHash) != 64 {
		t.Fatalf("Create() token hash length = %d, want 64", len(created.TokenHash))
	}
	if !strings.HasPrefix(rawToken, "ca_") {
		t.Fatalf("Create() raw token prefix missing: %q", rawToken)
	}

	validated, err := svc.Validate(strings.TrimPrefix(rawToken, "ca_"))
	if err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if validated.ID != created.ID {
		t.Fatalf("Validate() token id = %d, want %d", validated.ID, created.ID)
	}
	if validated.LastUsedAt == nil {
		t.Fatal("Validate() did not update last_used_at")
	}
}

func newTestTokenServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	user := &model.User{Username: "token-user", PasswordHash: "hash", Role: "admin"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return db
}
