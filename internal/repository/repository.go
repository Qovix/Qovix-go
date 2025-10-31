package repository

import (
	"context"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/pkg/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateVerificationCode(ctx context.Context, email, code string, expiry time.Time) error
	MarkEmailAsVerified(ctx context.Context, email string) error
	UpdateResetCode(ctx context.Context, email, code string, expiry time.Time) error
	UpdatePassword(ctx context.Context, email, hashedPassword string) error
}

type Repository struct {
	User UserRepository
}

func NewRepository(db *database.MongoDB) *Repository {
	return &Repository{
		User: NewUserRepository(db),
	}
}
