package middleware

import (
	"net/http"
	"strings"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/internal/services"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type AuthMiddleware struct {
	jwtService  *services.JWTService
	userService *services.UserService
	log         *logrus.Logger
}

func NewAuthMiddleware(jwtService *services.JWTService, userService *services.UserService) *AuthMiddleware {
	return &AuthMiddleware{
		jwtService:  jwtService,
		userService: userService,
		log:         logger.GetLogger(),
	}
}

func GetUserFromContext(c *gin.Context) (*models.User, bool) {
	user, exists := c.Get("user")
	if !exists {
		return nil, false
	}

	userModel, ok := user.(*models.User)
	return userModel, ok
}

func (m *AuthMiddleware) Authenticate() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Authorization header is required",
			})
			c.Abort()
			return
		}

		tokenParts := strings.Split(authHeader, " ")
		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid authorization header format",
			})
			c.Abort()
			return
		}

		token := tokenParts[1]

		claims, err := m.jwtService.ValidateToken(token)
		if err != nil {
			m.log.Errorf("Token validation failed: %v", err)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		userID, err := primitive.ObjectIDFromHex(claims.UserID)
		if err != nil {
			m.log.Errorf("Invalid user ID in token: %v", err)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "Invalid token",
			})
			c.Abort()
			return
		}

		user, err := m.userService.GetUserByID(c.Request.Context(), userID)
		if err != nil {
			m.log.Errorf("User not found: %v", err)
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "unauthorized",
				Message: "User not found",
			})
			c.Abort()
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("user", user)
		c.Set("token", token)

		c.Next()
	})
}

func GetTokenFromContext(c *gin.Context) (string, bool) {
	token, exists := c.Get("token")
	if !exists {
		return "", false
	}

	tokenStr, ok := token.(string)
	return tokenStr, ok
}
