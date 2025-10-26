package services

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"math/big"

	"github.com/Qovix/Qovix-go/internal/repository"
	"github.com/Qovix/Qovix-go/pkg/logger"
	"github.com/sirupsen/logrus"
)

type AuthService struct {
	userRepo        repository.UserRepository
	jwtService      *JWTService
	passwordService *PasswordService
	emailService    *EmailService
	log             *logrus.Logger
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtService *JWTService,
	passwordService *PasswordService,
	emailService *EmailService,
) *AuthService {
	return &AuthService{
		userRepo:        userRepo,
		jwtService:      jwtService,
		passwordService: passwordService,
		emailService:    emailService,
		log:             logger.GetLogger(),
	}
}

func (s *AuthService) GenerateVerificationCode() (string, error) {
	max := big.NewInt(1000000)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *AuthService) GenerateSecureToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
