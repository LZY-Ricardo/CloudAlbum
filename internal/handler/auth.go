package handler

import (
	"errors"
	"log"
	"net/http"
	"time"

	"cloudalbum/internal/repository"
	"cloudalbum/internal/service"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

const defaultAdminPassword = "admin123"

type AuthHandler struct {
	authSvc  *service.AuthService
	userRepo *repository.UserRepository
}

func NewAuthHandler(authSvc *service.AuthService, userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{authSvc: authSvc, userRepo: userRepo}
}

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "username and password required"})
		return
	}

	token, err := h.authSvc.Login(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func (h *AuthHandler) Me(c *gin.Context) {
	response := gin.H{
		"user_id": c.GetUint("user_id"),
	}
	if username := c.GetString("username"); username != "" {
		response["username"] = username
	}
	authType := c.GetString("auth_type")
	if authType != "" {
		response["auth_type"] = authType
	}
	if tokenScope := c.GetString("token_scope"); tokenScope != "" {
		response["token_scope"] = tokenScope
	}

	if authType == "jwt" {
		if user, err := h.userRepo.FindByID(c.GetUint("user_id")); err == nil {
			response["created_at"] = user.CreatedAt
			response["password_changed_at"] = user.PasswordChangedAt
			if user.Username == "admin" {
				response["uses_default_password"] = bcrypt.CompareHashAndPassword(
					[]byte(user.PasswordHash), []byte(defaultAdminPassword),
				) == nil
			}
		}
	}

	c.JSON(http.StatusOK, response)
}

type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	if c.GetString("auth_type") != "jwt" {
		c.JSON(http.StatusForbidden, gin.H{"error": "api_token_forbidden"})
		return
	}
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
		return
	}
	userID := c.GetUint("user_id")
	token, changedAt, err := h.authSvc.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		log.Printf("[settings] password change failed for user_id=%d ip=%s reason=%v", userID, c.ClientIP(), err)
		switch {
		case errors.Is(err, service.ErrInvalidCredentials):
			c.JSON(http.StatusUnauthorized, gin.H{"error": "wrong_old_password"})
		case errors.Is(err, service.ErrPasswordTooShort):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request", "detail": "min 8 characters"})
		case errors.Is(err, service.ErrPasswordSameAsOld):
			c.JSON(http.StatusBadRequest, gin.H{"error": "same_as_old"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "change_password_failed"})
		}
		return
	}
	log.Printf("[settings] password changed for user_id=%d ip=%s ts=%s", userID, c.ClientIP(), changedAt.UTC().Format(time.RFC3339))
	c.JSON(http.StatusOK, gin.H{
		"token":               token,
		"password_changed_at": changedAt,
	})
}
