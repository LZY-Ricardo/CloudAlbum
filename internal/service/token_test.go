package service

import (
	"strings"
	"testing"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTokenServiceCreateAndValidate(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	svc := NewTokenService(tokenRepo, testTokenProvider())

	created, rawToken, err := svc.Create(7, "cli", "upload", nil)
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

func TestTokenServiceValidateRejectsExpiredToken(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	svc := NewTokenService(tokenRepo, testTokenProvider())

	expiresIn := int64(1)
	created, rawToken, err := svc.Create(7, "cli", "upload", &expiresIn)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ExpiresAt == nil {
		t.Fatal("Create() should set expires_at when expires_in provided")
	}

	validated, err := svc.Validate(rawToken)
	if err != nil {
		t.Fatalf("Validate() immediate error = %v", err)
	}
	if validated.LastUsedAt == nil {
		t.Fatal("LastUsedAt should be updated before expiry")
	}

	expiredAt := created.CreatedAt.Add(-time.Minute)
	if err := db.Model(&model.APIToken{}).Where("id = ?", created.ID).Update("expires_at", expiredAt).Error; err != nil {
		t.Fatalf("expire token: %v", err)
	}
	before := *validated.LastUsedAt
	_, err = svc.Validate(rawToken)
	if err == nil {
		t.Fatal("Validate() error = nil, want invalid token")
	}

	reloaded, err := tokenRepo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("FindByID() error = %v", err)
	}
	if reloaded.LastUsedAt == nil || !reloaded.LastUsedAt.Equal(before) {
		t.Fatalf("LastUsedAt changed on expired token: before=%v after=%v", before, reloaded.LastUsedAt)
	}
}

func TestTokenServiceCreateUsesDefaultExpiry(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	svc := NewTokenService(tokenRepo, testTokenProvider())

	created, _, err := svc.Create(7, "cli", "upload", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ExpiresAt == nil {
		t.Fatal("default expiry should be applied")
	}
}

func TestTokenServiceCreateAllowsNoExpiryWhenConfigured(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	provider := config.NewProvider(config.Config{Token: config.TokenPolicyConfig{AllowNoExpiry: true, DefaultExpiresIn: 0}}, config.Overrides{})
	svc := NewTokenService(tokenRepo, provider)

	created, _, err := svc.Create(7, "cli", "upload", nil)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.ExpiresAt != nil {
		t.Fatalf("ExpiresAt = %v, want nil", created.ExpiresAt)
	}
}

func TestTokenServiceCreateRejectsNoExpiryWhenDisallowed(t *testing.T) {
	db := newTestTokenServiceDB(t)
	tokenRepo := repository.NewTokenRepository(db)
	provider := config.NewProvider(config.Config{Token: config.TokenPolicyConfig{AllowNoExpiry: false, DefaultExpiresIn: 0}}, config.Overrides{})
	svc := NewTokenService(tokenRepo, provider)

	_, _, err := svc.Create(7, "cli", "upload", nil)
	if err == nil {
		t.Fatal("Create() error = nil, want invalid expires_in")
	}
	if err.Error() != "invalid expires_in" {
		t.Fatalf("Create() error = %v, want invalid expires_in", err)
	}
}

func testTokenProvider() *config.Provider {
	base := config.Config{Token: config.TokenPolicyConfig{AllowNoExpiry: true, DefaultExpiresIn: 7 * 24 * time.Hour}}
	return config.NewProvider(base, config.Overrides{})
}

func newTestTokenServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := "file:" + strings.ReplaceAll(t.Name(), "/", "_") + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
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
