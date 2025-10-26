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

}
