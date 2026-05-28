package middleware

import (
	"net/http"
	"strings"

	"cloudalbum/internal/service"
	"github.com/gin-gonic/gin"
)

func AuthMiddleware(authSvc *service.AuthService, tokenSvc *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := extractBearerToken(c.GetHeader("Authorization"))
		if tokenStr == "" {
			tokenStr = strings.TrimSpace(c.Query("token"))
		}
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "authorization required"})
			return
		}

		if strings.HasPrefix(tokenStr, "ca_") {
			apiToken, err := tokenSvc.Validate(tokenStr)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
				return
			}

			c.Set("user_id", apiToken.UserID)
			c.Set("auth_type", "api_token")
			c.Set("token_id", apiToken.ID)
			c.Set("token_scope", apiToken.Scope)
			c.Next()
			return
		}

		claims, err := authSvc.ParseJWT(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		user, err := authSvc.LookupUser(claims.UserID)
		if err != nil || user.TokenVersion != claims.TokenVersion {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("auth_type", "jwt")
		c.Set("username", claims.Username)
		c.Next()
	}
}

func RequireScope(scopes ...string) gin.HandlerFunc {
	allowed := make(map[string]struct{}, len(scopes))
	for _, scope := range scopes {
		allowed[scope] = struct{}{}
	}

	return func(c *gin.Context) {
		authType := c.GetString("auth_type")
		if authType == "jwt" {
			c.Next()
			return
		}
		if authType != "api_token" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
			return
		}

		tokenScope := c.GetString("token_scope")
		if tokenScope == "full" {
			c.Next()
			return
		}
		if _, ok := allowed[tokenScope]; ok {
			c.Next()
			return
		}

		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "insufficient permissions"})
	}
}

func extractBearerToken(header string) string {
	header = strings.TrimSpace(header)
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
