package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"simvex/internal/api"
	"simvex/internal/api/handlers"
	"simvex/internal/api/middleware"
	"simvex/internal/repository"
	"simvex/internal/services"
	openaisvc "simvex/pkg/openai"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		log.Println("WARNING: OPENAI_API_KEY is not set; /api/ai/chat will fail.")
	}
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL is not set")
	}

	var repo *repository.PostgresRepository
	var err error
	for attempt := 1; attempt <= 10; attempt++ {
		repo, err = repository.NewPostgresRepository(dsn)
		if err == nil {
			break
		}
		log.Printf("database connection failed (attempt %d/10): %v", attempt, err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Fatalf("failed to init repository: %v", err)
	}
	defer repo.Close()

	openaiClient := openaisvc.NewClient(apiKey, model)
	objectService := services.NewObjectService(repo)
	aiService := services.NewAIService(repo, openaiClient)
	assetService := services.NewAssetService(repo)

	redisAddr := strings.TrimSpace(os.Getenv("REDIS_ADDR"))
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB := 0
	if raw := strings.TrimSpace(os.Getenv("REDIS_DB")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			redisDB = parsed
		}
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	smtpHost := strings.TrimSpace(os.Getenv("SMTP_HOST"))
	smtpPort := 587
	if raw := strings.TrimSpace(os.Getenv("SMTP_PORT")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil {
			smtpPort = parsed
		}
	}
	smtpUser := strings.TrimSpace(os.Getenv("SMTP_USER"))
	smtpPass := os.Getenv("SMTP_PASS")
	smtpFrom := strings.TrimSpace(os.Getenv("SMTP_FROM"))
	var mailer services.Mailer
	if smtpHost == "" || smtpFrom == "" {
		log.Println("WARNING: SMTP settings not configured; auth email sending will fail.")
	} else {
		mailer = services.NewSMTPMailer(smtpHost, smtpPort, smtpUser, smtpPass, smtpFrom)
	}

	sessionTTL := 14 * 24 * time.Hour
	codeTTL := 10 * time.Minute
	authService := services.NewAuthService(repo, redisClient, mailer, sessionTTL, codeTTL)

	router := gin.Default()
	allowedOrigins := []string{"https://simvex.com", "http://localhost:5173"}
	if extra := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); extra != "" {
		allowedOrigins = append(allowedOrigins, strings.Split(extra, ",")...)
	}
	router.Use(middleware.CORS(allowedOrigins))

	assetPath := os.Getenv("ASSET_IMPORT_PATH")
	if assetPath == "" {
		assetPath = "./assets/models"
	}
	if err := repo.SeedAssetsFromDir(assetPath); err != nil {
		log.Printf("asset seed failed: %v", err)
	}

	assetHandler := handlers.NewAssetHandler(assetService)
	router.GET("/assets/models/*filepath", assetHandler.GetAsset)
	router.HEAD("/assets/models/*filepath", assetHandler.GetAsset)

	cookieName := strings.TrimSpace(os.Getenv("SESSION_COOKIE_NAME"))
	if cookieName == "" {
		cookieName = "simvex_session"
	}
	cookieSecure := strings.TrimSpace(os.Getenv("COOKIE_SECURE")) == "true"

	api.RegisterRoutes(router, objectService, aiService, authService, cookieName, cookieSecure)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
