package main

import (
	"log"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"simvex/internal/api"
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

	router := gin.Default()
	allowedOrigins := []string{"https://simvex.com", "http://localhost:5173"}
	if extra := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); extra != "" {
		allowedOrigins = append(allowedOrigins, strings.Split(extra, ",")...)
	}
	router.Use(middleware.CORS(allowedOrigins))

	router.Static("/assets", "./assets")

	api.RegisterRoutes(router, objectService, aiService)

	if err := router.Run(":" + port); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
