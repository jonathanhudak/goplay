package database

import (
	"context"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	DB    *mongo.Database
	Logs  *mongo.Collection
	Users *mongo.Collection
)

func init() {
	dbPath := os.Getenv("MONGO_PATH")

	if len(dbPath) == 0 {
		dbPath = "localhost"
	}

	mongoURL := os.Getenv("MONGO_URL")

	if len(mongoURL) == 0 {
		mongoURL = "mongodb://" + dbPath + ":27017"
	}

	client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI(mongoURL))
	err := client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	} else {
		log.Println("Connected!")
	}

	DB = client.Database("jonapi")
	Logs = DB.Collection("logs")
	Users = DB.Collection("users")
}
