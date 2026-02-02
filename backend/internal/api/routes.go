package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"simvex/internal/api/handlers"
	"simvex/internal/api/middleware"
	"simvex/internal/services"
)

func RegisterRoutes(router *gin.Engine, objectService *services.ObjectService, aiService *services.AIService) {
	objectHandler := handlers.NewObjectHandler(objectService)
	partHandler := handlers.NewPartHandler(objectService)
	aiHandler := handlers.NewAIHandler(aiService)

	objectLimiter := middleware.NewRateLimiter(100, time.Minute)
	aiLimiter := middleware.NewRateLimiter(20, time.Minute)

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

		api.GET("/objects", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.ListObjects)
		api.GET("/objects/:id", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.GetObject)
		api.GET("/objects/:id/model", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.GetObjectModel)
		api.GET("/objects/:id/parts", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.GetPartsByObject)
		api.GET("/parts/:partId", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), partHandler.GetPart)

		api.POST("/ai/chat", aiLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), aiHandler.Chat)
	}
}
