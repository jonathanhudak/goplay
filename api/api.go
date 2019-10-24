package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var db *mongo.Database
var logs *mongo.Collection

// Log is the type for collection item
type Log struct {
	ID    *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Entry string              `json:"entry"`
}

// CreateLog makes a log
func CreateLog(w http.ResponseWriter, r *http.Request) {
	var logEntry Log
	err := json.NewDecoder(r.Body).Decode(&logEntry)
	if err != nil {
		log.Fatal("Invalid params", err)
	}

	insertResult, err := logs.InsertOne(context.TODO(), logEntry)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a log document: ", insertResult.InsertedID)
	resultJSON, err := json.Marshal(insertResult)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

// GetLog retrieves a log by using the route param
func GetLog(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	vars := mux.Vars(r)
	var logEntry Log
	fmt.Printf("GetLog %s\n", vars["_id"])
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}} // vet complains about this. Not sure composite literal uses unkeyed fields
	err := logs.FindOne(context.TODO(), filter).Decode(&logEntry)
	if err != nil {
		log.Fatal(err)
		notFound, _ := json.Marshal("Log entry not found")
		w.Write(notFound)
	}
	resultJSON, err := json.Marshal(logEntry)
	w.Write(resultJSON)
}

// DelteLog removes log by _id
func DeleteLog(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	fmt.Printf("DeleteLog %s\n", vars["_id"])
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}
	deleteResult, err := logs.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, _ := json.Marshal(deleteResult)

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

// GetLogs retrieves logs from the database as json
func GetLogs(w http.ResponseWriter, r *http.Request) {
	var results []*Log
	findOptions := options.Find()
	findOptions.SetLimit(10)

	cur, err := logs.Find(context.TODO(), bson.D{{}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through the cursor
	for cur.Next(context.TODO()) {
		var elem Log
		err := cur.Decode(&elem)
		if err != nil {
			log.Fatal(err)
		}

		results = append(results, &elem)
	}

	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}

	// Close the cursor once finished
	cur.Close(context.TODO())

	logsJSON, err := json.Marshal(results)

	w.Header().Set("Content-Type", "application/json")
	w.Write(logsJSON)
}

func init() {
	fmt.Println("api.go init")
	client, _ := mongo.Connect(context.TODO(), options.Client().ApplyURI("mongodb://localhost:27017"))
	err := client.Ping(context.TODO(), readpref.Primary())
	if err != nil {
		log.Fatal("Couldn't connect to the database", err)
	} else {
		log.Println("Connected!")
	}

	db = client.Database("jonapi")
	logs = db.Collection("logs")
}
