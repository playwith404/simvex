package api

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"simvex/internal/api/handlers"
	"simvex/internal/api/middleware"
	"simvex/internal/services"
)

func RegisterRoutes(
	router *gin.Engine,
	objectService *services.ObjectService,
	aiService *services.AIService,
	authService *services.AuthService,
	redisClient *redis.Client,
	cookieName string,
	cookieSecure bool,
) {
	objectHandler := handlers.NewObjectHandler(objectService)
	partHandler := handlers.NewPartHandler(objectService)
	aiHandler := handlers.NewAIHandler(aiService)
	authHandler := handlers.NewAuthHandler(authService, cookieName, cookieSecure)

	objectLimiter := middleware.NewRateLimiter(100, time.Minute)
	aiLimiter := middleware.NewRateLimiter(20, time.Minute)

	api := router.Group("/api")
	{
		api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })

		api.POST("/auth/register", authHandler.Register)
		api.POST("/auth/verify", authHandler.Verify)
		api.POST("/auth/login", authHandler.Login)
		api.POST("/auth/logout", authHandler.Logout)
		api.POST("/auth/password/reset-request", authHandler.RequestReset)
		api.POST("/auth/password/reset-confirm", authHandler.ConfirmReset)
		api.GET("/auth/me", middleware.AuthMiddleware(redisClient, cookieName), authHandler.Me)

		api.GET("/objects", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.ListObjects)
		api.GET("/objects/:id", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.GetObject)
		api.GET("/objects/:id/model", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.GetObjectModel)
		api.GET("/objects/:id/parts", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), objectHandler.GetPartsByObject)
		api.GET("/parts/:partId", objectLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), partHandler.GetPart)

		api.POST("/ai/chat", aiLimiter.Middleware("RATE_LIMIT_EXCEEDED", "잠시 후 다시 시도해주세요"), aiHandler.Chat)
	}
}
