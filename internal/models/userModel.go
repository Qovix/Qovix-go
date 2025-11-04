package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email              string             `bson:"email" json:"email"`
	Password           string             `bson:"password,omitempty" json:"-"`
	FirstName          string             `bson:"first_name" json:"first_name"`
	LastName           string             `bson:"last_name" json:"last_name"`
	Username           string             `bson:"username" json:"username"`
	Avatar             string             `bson:"avatar,omitempty" json:"avatar,omitempty"`
	IsVerified         bool               `bson:"is_verified" json:"is_verified"`
	VerificationCode   string             `bson:"verification_code,omitempty" json:"-"`
	VerificationExpiry time.Time          `bson:"verification_expiry,omitempty" json:"-"`
	ResetCode          string             `bson:"reset_code,omitempty" json:"-"`
	ResetExpiry        time.Time          `bson:"reset_expiry,omitempty" json:"-"`
	GoogleID           string             `bson:"google_id,omitempty" json:"google_id,omitempty"`
	Provider           string             `bson:"provider" json:"provider"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt          time.Time          `bson:"updated_at" json:"updated_at"`
}

type ErrorResponse struct {
	Error   string                 `json:"error"`
	Message string                 `json:"message"`
	Details map[string]interface{} `json:"details,omitempty"`
}

type QueryHistory struct {
	ID           primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID       primitive.ObjectID `bson:"user_id" json:"user_id"`
	Database     string             `bson:"database" json:"database"`
	Tables       []string           `bson:"tables" json:"tables"`
	UserQuery    string             `bson:"user_query" json:"user_query"`
	GeneratedSQL string             `bson:"generated_sql" json:"generated_sql"`
	Explanation  string             `bson:"explanation" json:"explanation"`
	Confidence   float64            `bson:"confidence" json:"confidence"`
	IsValid      bool               `bson:"is_valid" json:"is_valid"`
	ExecutedAt   *time.Time         `bson:"executed_at,omitempty" json:"executed_at,omitempty"`
	CreatedAt    time.Time          `bson:"created_at" json:"created_at"`
}
