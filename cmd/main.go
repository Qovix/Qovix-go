package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Qovix/Qovix-go/internal/config"
	"github.com/Qovix/Qovix-go/internal/handlers"
	"github.com/Qovix/Qovix-go/internal/middleware"
	"github.com/Qovix/Qovix-go/internal/repository"
	"github.com/Qovix/Qovix-go/internal/routes"
	"github.com/Qovix/Qovix-go/internal/services"
	"github.com/Qovix/Qovix-go/pkg/database"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	logger.Init(cfg.Server.Env)
	loggerInstance := logger.GetLogger()

	loggerInstance.Info("Starting Schema Builder Backend...")

	db, err := database.New(cfg.Database.URI, cfg.Database.Database)
	if err != nil {
		loggerInstance.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	repos := repository.NewRepository(db)

	jwtService := services.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpiresIn)
	emailService := services.NewEmailService(&cfg.Email)
	passwordService := services.NewPasswordService(cfg.Security.BcryptCost)

	userService := services.NewUserService(repos.User)
	authService := services.NewAuthService(repos.User, jwtService, passwordService, emailService)
	databaseService := services.NewDatabaseService(cfg.JWT.Secret, repos.DatabaseConnection)
	aiService := services.NewAIService(cfg.AI.GeminiAPIKey)

	authMiddleware := middleware.NewAuthMiddleware(jwtService, userService)
	securityMiddleware := middleware.NewSecurityMiddleware(cfg)

	authHandler := handlers.NewAuthHandler(authService, userService)
	databaseHandler := handlers.NewDatabaseHandler(databaseService)
	aiHandler := handlers.NewAIHandler(aiService, databaseService)

	if cfg.IsProduction() {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	routes.SetupRoutes(r, authHandler, databaseHandler, aiHandler, authMiddleware, securityMiddleware)

	server := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: r,
	}

	go func() {
		loggerInstance.Infof("Server starting on port %s", cfg.Server.Port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			loggerInstance.Fatalf("Failed to start server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	loggerInstance.Info("Server shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		loggerInstance.Errorf("Server forced to shutdown: %v", err)
	}

	loggerInstance.Info("Server exited")
}
