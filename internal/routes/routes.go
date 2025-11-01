package routes

import (
	"net/http"
	"time"

	"github.com/Qovix/Qovix-go/internal/handlers"
	"github.com/Qovix/Qovix-go/internal/middleware"
	"github.com/gin-gonic/gin"
)

func SetupRoutes(
	r *gin.Engine,
	authHandler *handlers.AuthHandler,
	databaseHandler *handlers.DatabaseHandler,
	aiHandler *handlers.AIHandler,
	authMiddleware *middleware.AuthMiddleware,
	securityMiddleware *middleware.SecurityMiddleware,
) {
	r.Use(securityMiddleware.ErrorHandler())
	r.Use(securityMiddleware.RequestID())
	r.Use(securityMiddleware.RequestLogger())
	r.Use(securityMiddleware.CORS())
	r.Use(securityMiddleware.SecurityHeaders())
	r.Use(securityMiddleware.MaxBodySize(10 << 20))
	r.Use(securityMiddleware.ValidateContentType())
	r.Use(securityMiddleware.RateLimit())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"version":   "1.0.0",
		})
	})

	auth := r.Group("/api/auth")
	{
		auth.POST("/signup", authHandler.SignUp)
		auth.POST("/login", authHandler.Login)
		auth.POST("/verify-email", authHandler.VerifyEmail)
		auth.POST("/resend-verification", authHandler.ResendVerification)
		auth.POST("/forgot-password", authHandler.ForgotPassword)
		auth.POST("/reset-password", authHandler.ResetPassword)
	}

	protected := r.Group("/api")
	protected.Use(authMiddleware.Authenticate())
	{
		protected.GET("/profile", authHandler.GetProfile)
		protected.PUT("/profile", authHandler.UpdateProfile)

		// Database connection routes
		database := protected.Group("/database")
		{
			database.POST("/test", databaseHandler.TestConnection)
			database.POST("/connect", databaseHandler.Connect)
			database.DELETE("/connections/:connection_id", databaseHandler.Disconnect)
			database.GET("/connections/:connection_id/status", databaseHandler.GetConnectionStatus)
			database.GET("/connections/:connection_id/schema", databaseHandler.GetSchema)
			database.POST("/connections/:connection_id/query", databaseHandler.ExecuteQuery)
		}

		// AI SQL Assistant routes
		ai := protected.Group("/ai")
		{
			ai.POST("/generate-sql", aiHandler.GenerateSQL)
			ai.POST("/validate-sql", aiHandler.ValidateSQL)
			ai.POST("/explain-sql", aiHandler.ExplainSQL)
			ai.GET("/history", aiHandler.GetQueryHistory)
		}
	}
}
