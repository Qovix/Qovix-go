package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type userRepository struct {
	Collection *mongo.Collection
}

func NewUserRepository(db *database.MongoDB) UserRepository {
	return &userRepository{
		Collection: db.GetCollection("users"),
	}
}

func (r *userRepository) CreateUser(ctx context.Context, user *models.User) error {
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	result, err := r.Collection.InsertOne(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}
	user.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}
func (r *userRepository) GetByID(ctx context.Context, id primitive.ObjectID) (*models.User, error) {
	var user models.User
	err := r.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}
