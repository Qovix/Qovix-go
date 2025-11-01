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

type QueryHistoryRepository interface {
	SaveQuery(ctx context.Context, history *models.QueryHistory) error
	GetUserQueryHistory(ctx context.Context, userID primitive.ObjectID, limit int, offset int) ([]*models.QueryHistory, error)
	GetQueryByID(ctx context.Context, id primitive.ObjectID) (*models.QueryHistory, error)
	UpdateQueryExecution(ctx context.Context, id primitive.ObjectID, executedAt time.Time) error
	DeleteQuery(ctx context.Context, id primitive.ObjectID) error
	GetQueryStats(ctx context.Context, userID primitive.ObjectID) (*QueryHistoryStats, error)
}

type DatabaseConnectionRepository interface {
	CreateConnection(ctx context.Context, conn *models.DatabaseConnection) error
	GetConnectionByID(ctx context.Context, id primitive.ObjectID) (*models.DatabaseConnection, error)
	GetUserConnections(ctx context.Context, userID primitive.ObjectID) ([]*models.DatabaseConnection, error)
	UpdateConnection(ctx context.Context, conn *models.DatabaseConnection) error
	DeleteConnection(ctx context.Context, id, userID primitive.ObjectID) error
	UpdateConnectionStatus(ctx context.Context, id primitive.ObjectID, status models.ConnectionStatus, version string, schemas []string) error
}

type Repository struct {
	User               UserRepository
	QueryHistory       QueryHistoryRepository
	DatabaseConnection DatabaseConnectionRepository
}

func NewRepository(db *database.MongoDB) *Repository {
	return &Repository{
		User:               NewUserRepository(db),
		QueryHistory:       NewQueryHistoryRepository(db),
		DatabaseConnection: NewDatabaseConnectionRepository(db),
	}
}
