package handler

import (
	"errors"
	"net/http"
	"strconv"

	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TokenHandler struct {
	tokenSvc *service.TokenService
}

func NewTokenHandler(tokenSvc *service.TokenService) *TokenHandler {
	return &TokenHandler{tokenSvc: tokenSvc}
}

type CreateTokenRequest struct {
	Name      string `json:"name" binding:"required"`
	Scope     string `json:"scope" binding:"required,oneof=read upload full"`
	ExpiresIn *int64 `json:"expires_in,omitempty"`
}

func (h *TokenHandler) List(c *gin.Context) {
	tokens, err := h.tokenSvc.List(c.GetUint("user_id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tokens": tokens})
}

func (h *TokenHandler) Create(c *gin.Context) {
	var req CreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, rawToken, err := h.tokenSvc.Create(c.GetUint("user_id"), req.Name, req.Scope, req.ExpiresIn)
	if err != nil {
		if err.Error() == "invalid expires_in" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid_request"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"token":     token,
		"raw_token": rawToken,
	})
}

func (h *TokenHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	if err := h.tokenSvc.Delete(uint(id), c.GetUint("user_id")); err != nil {
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "token not found"})
		case errors.Is(err, service.ErrTokenForbidden):
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "delete failed"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
