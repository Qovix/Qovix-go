package repository

import (
	"context"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type databaseConnectionRepository struct {
	db         *database.MongoDB
	collection *mongo.Collection
}

func NewDatabaseConnectionRepository(db *database.MongoDB) DatabaseConnectionRepository {
	return &databaseConnectionRepository{
		db:         db,
		collection: db.Database.Collection("database_connections"),
	}
}

func (r *databaseConnectionRepository) CreateConnection(ctx context.Context, conn *models.DatabaseConnection) error {
	conn.ID = primitive.NewObjectID()
	conn.CreatedAt = time.Now()
	conn.UpdatedAt = time.Now()
	conn.Status = models.StatusDisconnected

	_, err := r.collection.InsertOne(ctx, conn)
	return err
}

func (r *databaseConnectionRepository) GetConnectionByID(ctx context.Context, id primitive.ObjectID) (*models.DatabaseConnection, error) {
	var conn models.DatabaseConnection
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&conn)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *databaseConnectionRepository) GetUserConnections(ctx context.Context, userID primitive.ObjectID) ([]*models.DatabaseConnection, error) {
	filter := bson.M{"user_id": userID}
	opts := options.Find().SetSort(bson.D{{"created_at", -1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var connections []*models.DatabaseConnection
	for cursor.Next(ctx) {
		var conn models.DatabaseConnection
		if err := cursor.Decode(&conn); err != nil {
			continue
		}
		connections = append(connections, &conn)
	}

	return connections, cursor.Err()
}

func (r *databaseConnectionRepository) UpdateConnection(ctx context.Context, conn *models.DatabaseConnection) error {
	conn.UpdatedAt = time.Now()

	filter := bson.M{"_id": conn.ID}
	update := bson.M{
		"$set": bson.M{
			"name":        conn.Name,
			"type":        conn.Type,
			"host":        conn.Host,
			"port":        conn.Port,
			"database":    conn.Database,
			"username":    conn.Username,
			"password":    conn.Password,
			"status":      conn.Status,
			"last_tested": conn.LastTested,
			"version":     conn.Version,
			"schemas":     conn.Schemas,
			"updated_at":  conn.UpdatedAt,
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *databaseConnectionRepository) DeleteConnection(ctx context.Context, id, userID primitive.ObjectID) error {
	filter := bson.M{
		"_id":     id,
		"user_id": userID,
	}

	_, err := r.collection.DeleteOne(ctx, filter)
	return err
}

func (r *databaseConnectionRepository) UpdateConnectionStatus(ctx context.Context, id primitive.ObjectID, status models.ConnectionStatus, version string, schemas []string) error {
	filter := bson.M{"_id": id}
	update := bson.M{
		"$set": bson.M{
			"status":      status,
			"last_tested": time.Now(),
			"version":     version,
			"schemas":     schemas,
			"updated_at":  time.Now(),
		},
	}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}
