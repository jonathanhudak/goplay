package database

import (
	"context"
	"fmt"
	"goplay/model"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var (
	DB     *mongo.Database
	Logs   *mongo.Collection
	Users  *mongo.Collection
	Habits *mongo.Collection
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
	Habits = DB.Collection("habits")
}

// Creates a new Log
func CreateLog(logEntry model.Log) *mongo.InsertOneResult {
	result, err := Logs.InsertOne(context.TODO(), logEntry)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

var logsLookup = bson.D{
	{"from", "habits"},
	{"localField", "habits"},
	{"foreignField", "_id"},
	{"as", "habits_info"},
}

func GetLog(id primitive.ObjectID, ownerId primitive.ObjectID) model.Log {
	var logEntry model.Log

	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"_id", id}, {"user_id", ownerId}}}},
		{{"$lookup", logsLookup}},
	}
	cursor, err := Logs.Aggregate(context.Background(), pipeline)
	if err != nil {
		log.Fatal(err)
	}

	defer cursor.Close(context.Background())
	for cursor.Next(context.Background()) {
		err := cursor.Decode(&logEntry)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(logEntry)
	}

	return logEntry
}

func GetLogs(ownerId primitive.ObjectID) []*model.Log {
	var results []*model.Log

	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"user_id", ownerId}}}},
		{{"$lookup", logsLookup}},
	}

	cursor, err := Logs.Aggregate(context.Background(), pipeline)
	defer cursor.Close(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through the cursor
	for cursor.Next(context.Background()) {
		var logEntry model.Log
		err := cursor.Decode(&logEntry)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, &logEntry)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return results
}

func CreateHabit(habit model.Habit) *mongo.InsertOneResult {
	result, err := Habits.InsertOne(context.TODO(), habit)
	if err != nil {
		log.Fatal(err)
	}
	return result
}

func GetHabits(ownerId primitive.ObjectID) []*model.Habit {
	var results []*model.Habit

	pipeline := mongo.Pipeline{
		{{"$match", bson.D{{"user_id", ownerId}}}},
		// {{"$lookup", logsLookup}},
	}

	cursor, err := Habits.Aggregate(context.Background(), pipeline)
	defer cursor.Close(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through the cursor
	for cursor.Next(context.Background()) {
		var habit model.Habit
		err := cursor.Decode(&habit)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, &habit)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	return results
}
