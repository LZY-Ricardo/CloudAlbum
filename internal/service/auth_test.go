package service

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestAuthServiceLoginAndParseJWT(t *testing.T) {
	db := newTestAuthServiceDB(t)
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authSvc := NewAuthService(userRepo, tokenRepo, config.AuthConfig{
		JWTSecret:   "test-secret",
		TokenExpire: time.Hour,
	})

	token, err := authSvc.Login("auth-user", "password123")
	if err != nil {
		t.Fatalf("Login() error = %v", err)
	}

	claims, err := authSvc.ParseJWT(token)
	if err != nil {
		t.Fatalf("ParseJWT() error = %v", err)
	}
	if claims.Username != "auth-user" {
		t.Fatalf("ParseJWT() username = %q, want auth-user", claims.Username)
	}
	if claims.UserID == 0 {
		t.Fatal("ParseJWT() user id was not populated")
	}
}

func TestAuthServiceLoginInvalidCredentials(t *testing.T) {
	db := newTestAuthServiceDB(t)
	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authSvc := NewAuthService(userRepo, tokenRepo, config.AuthConfig{
		JWTSecret:   "test-secret",
		TokenExpire: time.Hour,
	})

	_, err := authSvc.Login("auth-user", "wrong-password")
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, want ErrInvalidCredentials", err)
	}
}

func TestAuthServiceLoginCorruptHashReturnsBackendError(t *testing.T) {
	db := newTestAuthServiceDB(t)
	if err := db.Model(&model.User{}).Where("username = ?", "auth-user").Update("password_hash", "not-a-bcrypt-hash").Error; err != nil {
		t.Fatalf("corrupt password hash: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	authSvc := NewAuthService(userRepo, tokenRepo, config.AuthConfig{
		JWTSecret:   "test-secret",
		TokenExpire: time.Hour,
	})

	_, err := authSvc.Login("auth-user", "password123")
	if err == nil {
		t.Fatal("Login() error = nil, want backend error")
	}
	if errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("Login() error = %v, should not be classified as invalid credentials", err)
	}
}

func newTestAuthServiceDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", t.Name())
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := &model.User{Username: "auth-user", PasswordHash: string(hash), Role: "admin"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return db
}
