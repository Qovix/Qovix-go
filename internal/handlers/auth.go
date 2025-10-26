package handlers

import (
	"github.com/Qovix/Qovix-go/internal/services"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/sirupsen/logrus"
)

type AuthHandler struct {
	authService *services.AuthService
	userService *services.UserService
	log         *logrus.Logger
}

func NewAuthHandler(authService *services.AuthService, userService *services.UserService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
		log:         logger.GetLogger(),
	}
}
