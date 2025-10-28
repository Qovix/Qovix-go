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

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.Collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	return &user, nil
}

func (r *userRepository) UpdateVerificationCode(ctx context.Context, email, code string, expiry time.Time) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"verification_code":   code,
			"verification_expiry": expiry,
			"updated_at":          time.Now(),
		},
	}

	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update verification code: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) MarkEmailAsVerified(ctx context.Context, email string) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"is_verified":         true,
			"verification_code":   "",
			"verification_expiry": time.Time{},
			"updated_at":          time.Now(),
		},
	}

	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to verify email: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) UpdateResetCode(ctx context.Context, email, code string, expiry time.Time) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"reset_code":   code,
			"reset_expiry": expiry,
			"updated_at":   time.Now(),
		},
	}

	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update reset code: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, email, hashedPassword string) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"password":     hashedPassword,
			"reset_code":   "",
			"reset_expiry": time.Time{},
			"updated_at":   time.Now(),
		},
	}

	result, err := r.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update password: %v", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}
