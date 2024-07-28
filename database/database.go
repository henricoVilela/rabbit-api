package database

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var db *mongo.Database

type AuditLog struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	UserId      string             `bson:"userId"`
	UserAgent   string             `bson:"UserAgent"`
	Application string             `bson:"application"`
	Message     string             `bson:"message"`
	Timestamp   time.Time          `bson:"timestamp"`
	IP          string             `bson:"ip"`
}

func ConnectDB() error {
	var err error

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatalf("Falha ao conectar ao MongoDB: %v", err)
	}

	// Ping the database to verify the connection
	err = client.Ping(ctx, nil)
	if err != nil {
		return nil
	}

	db = client.Database("notificationDB")
	log.Println("Connected to MongoDB")

	return nil
}

func DisconnectDB() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return db.Client().Disconnect(ctx)
}

func InsertAuditLog(log AuditLog) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	auditCollection := db.Collection("auditLogs")

	_, err := auditCollection.InsertOne(ctx, bson.M{
		"userId":      log.UserId,
		"application": log.Application,
		"message":     log.Message,
		"timestamp":   log.Timestamp,
		"ip":          log.IP,
		"userAgent":   log.UserAgent,
	})

	return err
}
