package handlers

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

type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
	log         *logrus.Logger
}

type ResetPasswordRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Code        string `json:"code" binding:"required,len=6"`
	NewPassword string `json:"new_password" binding:"required,min=8"`
}

type ResendVerificationRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

func NewAuthHandler(authService *services.AuthService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		log:         logger.GetLogger(),
	}
}

func (h *AuthHandler) SignUp(c *gin.Context) {
	var req services.SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: map[string]interface{}{"validation": err.Error()},
		})
		return
	}

	response, err := h.authService.SignUp(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Error:   "user_exists",
				Message: err.Error(),
			})
			return
		}

		h.log.WithError(err).Error("SignUp failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "signup_failed",
			Message: "Failed to create account",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Account created successfully. Please check your email for verification code.",
		"data":    response,
	})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req services.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: map[string]interface{}{"validation": err.Error()},
		})
		return
	}

	response, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid credentials") {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error:   "invalid_credentials",
				Message: "Invalid email or password",
			})
			return
		}

		h.log.WithError(err).Error("Login failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "login_failed",
			Message: "Login failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"data":    response,
	})
}

func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	var req services.VerifyEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: map[string]interface{}{"validation": err.Error()},
		})
		return
	}

	err := h.authService.VerifyEmail(c.Request.Context(), &req)
	if err != nil {
		if strings.Contains(err.Error(), "invalid verification code") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_code",
				Message: "Invalid verification code",
			})
			return
		}

		if strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "code_expired",
				Message: "Verification code has expired",
			})
			return
		}

		if strings.Contains(err.Error(), "already verified") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "already_verified",
				Message: "Email is already verified",
			})
			return
		}

		h.log.WithError(err).Error("Email verification failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "verification_failed",
			Message: "Email verification failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Email verified successfully",
	})
}

func (h *AuthHandler) ResendVerification(c *gin.Context) {
	var req ResendVerificationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: map[string]interface{}{"validation": err.Error()},
		})
		return
	}

	err := h.authService.ResendVerification(c.Request.Context(), req.Email)
	if err != nil {
		if strings.Contains(err.Error(), "already verified") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "already_verified",
				Message: "Email is already verified",
			})
			return
		}

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, models.ErrorResponse{
				Error:   "user_not_found",
				Message: "User not found",
			})
			return
		}

		h.log.WithError(err).Error("ResendVerification failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "resend_failed",
			Message: "Failed to resend verification code",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Verification code sent successfully",
	})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: map[string]interface{}{"validation": err.Error()},
		})
		return
	}

	err := h.authService.ForgotPassword(c.Request.Context(), req.Email)
	if err != nil {
		h.log.WithError(err).Error("ForgotPassword failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "forgot_password_failed",
			Message: "Failed to send password reset email",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "If the email exists, a password reset code has been sent",
	})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request data",
			Details: map[string]interface{}{"validation": err.Error()},
		})
		return
	}

	err := h.authService.ResetPassword(c.Request.Context(), req.Email, req.Code, req.NewPassword)
	if err != nil {
		if strings.Contains(err.Error(), "invalid reset code") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "invalid_code",
				Message: "Invalid reset code",
			})
			return
		}

		if strings.Contains(err.Error(), "expired") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Error:   "code_expired",
				Message: "Reset code has expired",
			})
			return
		}

		h.log.WithError(err).Error("ResetPassword failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "reset_failed",
			Message: "Password reset failed",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password reset successfully",
	})
}

func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error:   "unauthorized",
			Message: "User not authenticated",
		})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_user_id",
			Message: "Invalid user ID",
		})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), objectID)
	if err != nil {
		h.log.WithError(err).Error("GetProfile failed")
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error:   "profile_fetch_failed",
			Message: "Failed to fetch profile",
		})
		return
	}

	user.Password = ""
	user.VerificationCode = ""
	user.ResetCode = ""

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, models.ErrorResponse{
		Error:   "not_implemented",
		Message: "Profile update not yet implemented",
	})
}
