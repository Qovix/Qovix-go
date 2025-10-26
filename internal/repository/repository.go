package repository

import (
	"context"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/pkg/database"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error)
}

type Repository struct {
	User UserRepository
}

func NewRepository(db *database.MongoDB) *Repository {
	return &Repository{
		User: NewUserRepository(db),
	}
}
