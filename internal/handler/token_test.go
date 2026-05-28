package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cloudalbum/internal/config"
	"cloudalbum/internal/model"
	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTokenHandlerCreateAcceptsExpiresIn(t *testing.T) {
	gin.SetMode(gin.TestMode)
	db, userID := newTokenHandlerTestDB(t)
	repo := repository.NewTokenRepository(db)
	provider := config.NewProvider(config.Config{Token: config.TokenPolicyConfig{AllowNoExpiry: true, DefaultExpiresIn: 7 * 24 * time.Hour}}, config.Overrides{})
	h := NewTokenHandler(service.NewTokenService(repo, provider))

	rec := httptest.NewRecorder()
	_, engine := gin.CreateTestContext(rec)
	engine.POST("/tokens", func(c *gin.Context) {
		c.Set("user_id", userID)
		h.Create(c)
	})

	body, _ := json.Marshal(gin.H{"name": "cli", "scope": "upload", "expires_in": 3600})
	req := httptest.NewRequest(http.MethodPost, "/tokens", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	engine.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d body=%s", rec.Code, rec.Body.String())
	}
	if !bytes.Contains(rec.Body.Bytes(), []byte(`"expires_at"`)) {
		t.Fatalf("response missing expires_at: %s", rec.Body.String())
	}
}

func newTokenHandlerTestDB(t *testing.T) (*gorm.DB, uint) {
	t.Helper()
	dsn := "file:" + t.Name() + "?mode=memory&cache=shared"
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&model.User{}, &model.APIToken{}); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	user := &model.User{Username: "admin", PasswordHash: "hash", Role: "admin"}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return db, user.ID
}
