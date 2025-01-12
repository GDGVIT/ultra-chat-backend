package repositories

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"ultra-chat-backend/models"
)

type UserRepository interface {
	FindUserByID(id string) (*models.User, error)
	CreateUser(user *models.User) error
	UpdateUser(id string, update bson.M) error
	AddSummary(userID string, summary bson.M) error
	GetSummaries(userID string) ([]bson.M, error)
	UpdateSummary(userID, summaryID, content string) error
	DeleteSummary(userID, summaryID string) error
	IsAuthenticated(userID string) (bool, error)
}

type userRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

func (r *userRepository) FindUserByID(id string) (*models.User, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user models.User
	err := r.collection.FindOne(ctx, bson.M{"id": id}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) CreateUser(user *models.User) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.InsertOne(ctx, user)
	return err
}

func (r *userRepository) UpdateUser(id string, update bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(ctx, bson.M{"id": id}, bson.M{"$set": update})
	return err
}

func (r *userRepository) AddSummary(userID string, summary bson.M) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := r.collection.UpdateOne(ctx, bson.M{"id": userID}, bson.M{"$push": bson.M{"summaries": summary}})
	return err
}

func (r *userRepository) GetSummaries(userID string) ([]bson.M, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var user struct {
		Summaries []bson.M `bson:"summaries"`
	}
	err := r.collection.FindOne(ctx, bson.M{"id": userID}).Decode(&user)
	if err != nil {
		return nil, err
	}
	return user.Summaries, nil
}

func (r *userRepository) UpdateSummary(userID, summaryID, content string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": userID, "summaries.id": summaryID}
	update := bson.M{"$set": bson.M{"summaries.$.content": content}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepository) DeleteSummary(userID, summaryID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"id": userID}
	update := bson.M{"$pull": bson.M{"summaries": bson.M{"id": summaryID}}}

	_, err := r.collection.UpdateOne(ctx, filter, update)
	return err
}

func (r *userRepository) IsAuthenticated(userID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	count, err := r.collection.CountDocuments(ctx, bson.M{"id": userID})
	return count > 0, err
}
