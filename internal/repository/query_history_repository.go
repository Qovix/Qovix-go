package repository

import (
	"context"
	"time"

	"github.com/Qovix/Qovix-go/internal/models"
	"github.com/Qovix/Qovix-go/pkg/database"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type QueryHistoryStats struct {
	TotalQueries      int64   `json:"total_queries"`
	SuccessfulQueries int64   `json:"successful_queries"`
	AverageConfidence float64 `json:"average_confidence"`
	MostUsedDatabase  string  `json:"most_used_database"`
}

type queryHistoryRepository struct {
	db *database.MongoDB
}

func NewQueryHistoryRepository(db *database.MongoDB) QueryHistoryRepository {
	return &queryHistoryRepository{
		db: db,
	}
}

func (r *queryHistoryRepository) SaveQuery(ctx context.Context, history *models.QueryHistory) error {
	history.CreatedAt = time.Now()

	collection := r.db.Database.Collection("query_history")
	result, err := collection.InsertOne(ctx, history)
	if err != nil {
		return err
	}

	history.ID = result.InsertedID.(primitive.ObjectID)
	return nil
}

func (r *queryHistoryRepository) GetUserQueryHistory(ctx context.Context, userID primitive.ObjectID, limit int, offset int) ([]*models.QueryHistory, error) {
	collection := r.db.Database.Collection("query_history")

	filter := bson.M{"user_id": userID}
	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(int64(limit)).
		SetSkip(int64(offset))

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var histories []*models.QueryHistory
	if err := cursor.All(ctx, &histories); err != nil {
		return nil, err
	}

	return histories, nil
}

func (r *queryHistoryRepository) GetQueryByID(ctx context.Context, id primitive.ObjectID) (*models.QueryHistory, error) {
	collection := r.db.Database.Collection("query_history")

	var history models.QueryHistory
	err := collection.FindOne(ctx, bson.M{"_id": id}).Decode(&history)
	if err != nil {
		return nil, err
	}

	return &history, nil
}

func (r *queryHistoryRepository) UpdateQueryExecution(ctx context.Context, id primitive.ObjectID, executedAt time.Time) error {
	collection := r.db.Database.Collection("query_history")

	_, err := collection.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": bson.M{"executed_at": executedAt}},
	)

	return err
}

func (r *queryHistoryRepository) DeleteQuery(ctx context.Context, id primitive.ObjectID) error {
	collection := r.db.Database.Collection("query_history")

	_, err := collection.DeleteOne(ctx, bson.M{"_id": id})
	return err
}

func (r *queryHistoryRepository) GetQueryStats(ctx context.Context, userID primitive.ObjectID) (*QueryHistoryStats, error) {
	collection := r.db.Database.Collection("query_history")

	pipeline := []bson.M{
		{"$match": bson.M{"user_id": userID}},
		{"$group": bson.M{
			"_id":           nil,
			"total_queries": bson.M{"$sum": 1},
			"successful_queries": bson.M{
				"$sum": bson.M{
					"$cond": bson.M{
						"if":   bson.M{"$eq": []interface{}{"$is_valid", true}},
						"then": 1,
						"else": 0,
					},
				},
			},
			"average_confidence": bson.M{"$avg": "$confidence"},
			"databases":          bson.M{"$push": "$database"},
		}},
	}

	cursor, err := collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}

	if len(results) == 0 {
		return &QueryHistoryStats{}, nil
	}

	result := results[0]
	stats := &QueryHistoryStats{
		TotalQueries:      result["total_queries"].(int64),
		SuccessfulQueries: result["successful_queries"].(int64),
		AverageConfidence: result["average_confidence"].(float64),
	}

	databases := result["databases"].(primitive.A)
	databaseCounts := make(map[string]int)
	for _, db := range databases {
		if dbStr, ok := db.(string); ok {
			databaseCounts[dbStr]++
		}
	}

	var mostUsed string
	var maxCount int
	for db, count := range databaseCounts {
		if count > maxCount {
			maxCount = count
			mostUsed = db
		}
	}
	stats.MostUsedDatabase = mostUsed

	return stats, nil
}
