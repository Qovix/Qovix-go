package services

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
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

type SignUpRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required,min=2"`
	LastName  string `json:"last_name" binding:"required,min=2"`
	Username  string `json:"username" binding:"required,min=3"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type VerifyEmailRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}

type AuthResponse struct {
	Token string       `json:"token"`
	User  *models.User `json:"user"`
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

func (s *AuthService) SignUp(ctx context.Context, req *SignUpRequest) (*AuthResponse, error) {
	existingUser, _ := s.userRepo.GetByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email already exists")
	}

	hashedPassword, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		s.log.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("failed to create user")
	}

	verificationCode, err := s.generateVerificationCode()
	if err != nil {
		s.log.WithError(err).Error("Failed to generate verification code")
		return nil, fmt.Errorf("failed to create user")
	}

	user := &models.User{
		Email:              strings.ToLower(req.Email),
		Password:           hashedPassword,
		FirstName:          req.FirstName,
		LastName:           req.LastName,
		Username:           req.Username,
		IsVerified:         false,
		VerificationCode:   verificationCode,
		VerificationExpiry: time.Now().Add(15 * time.Minute),
		Provider:           "email",
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		s.log.WithError(err).Error("Failed to create user")
		return nil, fmt.Errorf("failed to create user")
	}

	err = s.emailService.SendVerificationEmail(user.Email, verificationCode)
	if err != nil {
		s.log.WithError(err).Error("Failed to send verification email")
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.log.WithError(err).Error("Failed to generate token")
		return nil, fmt.Errorf("failed to create user")
	}

	user.Password = ""
	user.VerificationCode = ""

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, req *LoginRequest) (*AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err = s.passwordService.CheckPassword(user.Password, req.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	token, err := s.jwtService.GenerateToken(user.ID, user.Email)
	if err != nil {
		s.log.WithError(err).Error("Failed to generate token")
		return nil, fmt.Errorf("failed to login")
	}

	user.Password = ""
	user.VerificationCode = ""
	user.ResetCode = ""

	return &AuthResponse{
		Token: token,
		User:  user,
	}, nil
}

func (s *AuthService) VerifyEmail(ctx context.Context, req *VerifyEmailRequest) error {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(req.Email))
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsVerified {
		return fmt.Errorf("email already verified")
	}

	if user.VerificationCode != req.Code {
		return fmt.Errorf("invalid verification code")
	}

	if time.Now().After(user.VerificationExpiry) {
		return fmt.Errorf("verification code expired")
	}

	err = s.userRepo.MarkEmailAsVerified(ctx, user.Email)
	if err != nil {
		s.log.WithError(err).Error("Failed to mark email as verified")
		return fmt.Errorf("failed to verify email")
	}

	return nil
}

func (s *AuthService) ResendVerification(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.IsVerified {
		return fmt.Errorf("email already verified")
	}

	verificationCode, err := s.generateVerificationCode()
	if err != nil {
		s.log.WithError(err).Error("Failed to generate verification code")
		return fmt.Errorf("failed to resend verification")
	}

	err = s.userRepo.UpdateVerificationCode(ctx, user.Email, verificationCode, time.Now().Add(15*time.Minute))
	if err != nil {
		s.log.WithError(err).Error("Failed to update verification code")
		return fmt.Errorf("failed to resend verification")
	}

	err = s.emailService.SendVerificationEmail(user.Email, verificationCode)
	if err != nil {
		s.log.WithError(err).Error("Failed to send verification email")
		return fmt.Errorf("failed to send verification email")
	}

	return nil
}

func (s *AuthService) ForgotPassword(ctx context.Context, email string) error {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return nil
	}

	resetCode, err := s.generateVerificationCode()
	if err != nil {
		s.log.WithError(err).Error("Failed to generate reset code")
		return fmt.Errorf("failed to send reset email")
	}

	err = s.userRepo.UpdateResetCode(ctx, user.Email, resetCode, time.Now().Add(15*time.Minute))
	if err != nil {
		s.log.WithError(err).Error("Failed to update reset code")
		return fmt.Errorf("failed to send reset email")
	}

	err = s.emailService.SendPasswordResetEmail(user.Email, resetCode)
	if err != nil {
		s.log.WithError(err).Error("Failed to send reset email")
		return fmt.Errorf("failed to send reset email")
	}

	return nil
}

func (s *AuthService) ResetPassword(ctx context.Context, email, code, newPassword string) error {
	user, err := s.userRepo.GetByEmail(ctx, strings.ToLower(email))
	if err != nil {
		return fmt.Errorf("invalid reset code")
	}

	if user.ResetCode != code {
		return fmt.Errorf("invalid reset code")
	}

	if time.Now().After(user.ResetExpiry) {
		return fmt.Errorf("reset code expired")
	}

	hashedPassword, err := s.passwordService.HashPassword(newPassword)
	if err != nil {
		s.log.WithError(err).Error("Failed to hash password")
		return fmt.Errorf("failed to reset password")
	}

	err = s.userRepo.UpdatePassword(ctx, user.Email, hashedPassword)
	if err != nil {
		s.log.WithError(err).Error("Failed to update password")
		return fmt.Errorf("failed to reset password")
	}

	return nil
}

func (s *AuthService) generateVerificationCode() (string, error) {
	code := ""
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		code += num.String()
	}
	return code, nil
}
