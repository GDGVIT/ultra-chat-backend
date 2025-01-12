package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database
var mongoClient *mongo.Client

func ConnectDB() *mongo.Database {
	uri := os.Getenv("MONGO_URI")
	fmt.Println("MongoDB URI:", uri)

	if uri == "" {
		log.Fatal("MONGO_URI not set in environment")
	}

	clientOpts := options.Client().ApplyURI(uri).SetMaxPoolSize(100)
	fmt.Println("Connecting to MongoDB...")

	client, err := mongo.NewClient(clientOpts)
	if err != nil {
		log.Fatalf("Failed to create MongoDB client: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}

	fmt.Println("Pinging MongoDB...")
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	fmt.Println("Connected to MongoDB successfully")
	return client.Database("discord_oauth")
}

// DisconnectDB gracefully disconnects MongoDB
func DisconnectDB() {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := mongoClient.Disconnect(ctx); err != nil {
			log.Println("Failed to disconnect MongoDB: ", err)
		} else {
			log.Println("Disconnected from MongoDB")
		}
	}
}
