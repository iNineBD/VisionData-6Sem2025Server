package mongo

import (
	"context"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"go.mongodb.org/mongo-driver/bson"
)

// MongoInternal is a struct that contains a MongoDB client
type MongoInternal struct {
	client *mongo.Client
}

// NewMongoInternal is a function that returns a new MongoInternal struct
func NewMongoInternal() (*MongoInternal, error) {

	uri := os.Getenv("MONGO_URI")

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server
	client, err := mongo.Connect(context.TODO(), opts)
	if err != nil {
		return nil, err
	}

	if err := client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return nil, err
	}

	return &MongoInternal{
		client: client,
	}, nil

}
