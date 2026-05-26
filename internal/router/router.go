package router

import (
	"cloudalbum/internal/handler"
	"cloudalbum/internal/middleware"
	"cloudalbum/internal/service"

	"github.com/gin-gonic/gin"
)

func Setup(r *gin.Engine, authSvc *service.AuthService, tokenSvc *service.TokenService, authHandler *handler.AuthHandler, tokenHandler *handler.TokenHandler, imageHandler *handler.ImageHandler, albumHandler *handler.AlbumHandler, publicHandler *handler.PublicHandler) {
	r.Use(middleware.CORS())

	r.GET("/i/*key", publicHandler.Image)
	r.GET("/t/*key", publicHandler.Thumbnail)

	auth := r.Group("/api/v1/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/logout", authHandler.Logout)

	api := r.Group("/api/v1")
	api.Use(middleware.AuthMiddleware(authSvc, tokenSvc))
	api.GET("/auth/me", authHandler.Me)

	images := api.Group("/images")
	images.POST("", middleware.RequireScope("upload"), imageHandler.Upload)
	images.POST("/upload-url", middleware.RequireScope("upload"), imageHandler.UploadURL)
	images.GET("", middleware.RequireScope("read", "upload"), imageHandler.List)
	images.GET("/stats", middleware.RequireScope("read", "upload"), imageHandler.Stats)
	images.GET("/:id", middleware.RequireScope("read", "upload"), imageHandler.Get)
	images.PUT("/:id", middleware.RequireScope("upload"), imageHandler.Update)
	images.DELETE("/:id", middleware.RequireScope("upload"), imageHandler.Delete)
	images.POST("/batch", middleware.RequireScope("upload"), imageHandler.Batch)
	images.POST("/:id/restore", middleware.RequireScope("upload"), imageHandler.Restore)
	images.DELETE("/:id/permanent", middleware.RequireScope("upload"), imageHandler.HardDelete)

	albums := api.Group("/albums")
	albums.GET("", albumHandler.List)
	albums.POST("", albumHandler.Create)
	albums.PUT("/:id", albumHandler.Update)
	albums.DELETE("/:id", albumHandler.Delete)

	tokens := api.Group("/tokens")
	tokens.GET("", tokenHandler.List)
	tokens.POST("", tokenHandler.Create)
	tokens.DELETE("/:id", tokenHandler.Delete)

}
