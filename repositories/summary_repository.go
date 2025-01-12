package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoSummaryRepository handles operations related to summaries and users
type MongoSummaryRepository struct {
	collection     *mongo.Collection
	userCollection *mongo.Collection
}

// NewMongoSummaryRepository initializes the repository with MongoDB collections
func NewMongoSummaryRepository(db *mongo.Database) (*MongoSummaryRepository, error) {
	usersCollection := db.Collection("users")
	summariesCollection := db.Collection("summaries")

	// Create unique index on user_id in the users collection
	if _, err := usersCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	}); err != nil {
		return nil, errors.New("failed to create index on users collection: " + err.Error())
	}

	// Create index on user_id and server_id in the summaries collection
	if _, err := summariesCollection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
		Keys: bson.D{{Key: "user_id", Value: 1}, {Key: "server_id", Value: 1}},
	}); err != nil {
		return nil, errors.New("failed to create index on summaries collection: " + err.Error())
	}

	return &MongoSummaryRepository{
		collection:     summariesCollection,
		userCollection: usersCollection,
	}, nil
}

// AddSummary inserts a new summary into the summaries collection
func (r *MongoSummaryRepository) AddSummary(summaryID, userID, serverID string, isPrivate bool, summaryContent, createdAt string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	summary := bson.M{
		"summary_id": summaryID,
		"user_id":    userID,
		"server_id":  serverID,
		"is_private": isPrivate,
		"summary":    summaryContent,
		"created_at": createdAt,
		"updated_at": createdAt,
	}

	if _, err := r.collection.InsertOne(ctx, summary); err != nil {
		return fmt.Errorf("failed to add summary: %w", err)
	}
	return nil
}

// GetSummaries retrieves summaries matching the provided filter
func (r *MongoSummaryRepository) GetSummaries(filter bson.M) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, errors.New("failed to retrieve summaries: " + err.Error())
	}
	defer cursor.Close(ctx)

	var summaries []bson.M
	if err := cursor.All(ctx, &summaries); err != nil {
		return nil, errors.New("failed to decode summaries: " + err.Error())
	}
	return summaries, nil
}

// UpdateSummary modifies the summary content for the given filter
func (r *MongoSummaryRepository) UpdateSummary(userID, serverID string, isPrivate bool, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID, "server_id": serverID, "is_private": isPrivate}
	update := bson.M{
		"$set": bson.M{
			"summary":    content,
			"updated_at": time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return errors.New("failed to update summary: " + err.Error())
	}
	if result.MatchedCount == 0 {
		return errors.New("no matching summary found")
	}
	return nil
}

func (r *MongoSummaryRepository) DeleteSummary(userID, summaryID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{
		"user_id":    userID,
		"summary_id": summaryID,
	}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete summary: %w", err)
	}

	if result.DeletedCount == 0 {
		return errors.New("no matching summary found")
	}
	return nil
}

func (r *MongoSummaryRepository) CheckUserExists(userID string) (bool, error) {
	filter := bson.M{"id": userID} // Use "id" field instead of "_id"
	var result bson.M
	err := r.userCollection.FindOne(context.Background(), filter).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil // User not found
		}
		return false, err // Other errors
	}
	return true, nil
}
