package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	Client   *mongo.Client
	Database *mongo.Database
)

// Collections
var (
	UsersCollection      *mongo.Collection
	RecipesCollection    *mongo.Collection
	CategoriesCollection *mongo.Collection
)

// Connect establishes connection to MongoDB Atlas
func Connect(uri string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return err
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return err
	}

	Client = client
	Database = client.Database("culinaryshare")

	// Initialize collections
	UsersCollection = Database.Collection("users")
	RecipesCollection = Database.Collection("recipes")
	CategoriesCollection = Database.Collection("categories")

	// Create indexes
	if err := createIndexes(ctx); err != nil {
		log.Printf("Warning: Failed to create indexes: %v", err)
	}

	log.Println("Connected to MongoDB Atlas successfully")
	return nil
}

// createIndexes creates all required indexes
func createIndexes(ctx context.Context) error {
	// Unique index on users.email
	// This ensures no duplicate email addresses in the system
	_, err := UsersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	log.Println("Created unique index on users.email")

	// Unique index on users.username
	_, err = UsersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "username", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return err
	}
	log.Println("Created unique index on users.username")

	// Compound index on recipes: categoryId + averageRating
	// Purpose: Optimizes queries that filter by category and sort by rating
	// Common query pattern: "Show me the best recipes in category X"
	// The index supports:
	// 1. Filtering by categoryId (equality match)
	// 2. Sorting by averageRating in descending order
	// 3. Combined filter + sort operations efficiently
	_, err = RecipesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "categoryId", Value: 1},
			{Key: "averageRating", Value: -1},
		},
	})
	if err != nil {
		return err
	}
	log.Println("Created compound index on recipes(categoryId, averageRating)")

	// Index on recipes.userId for efficient user recipe lookups
	_, err = RecipesCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "userId", Value: 1}},
	})
	if err != nil {
		return err
	}
	log.Println("Created index on recipes.userId")

	return nil
}

// Disconnect closes the MongoDB connection
func Disconnect() {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := Client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from MongoDB: %v", err)
		}
	}
}
