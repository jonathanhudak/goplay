package main

import (
	"fmt"
	"log"
	"encoding/json"
	"net/http"
	"time"
	"context"
	"go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
		"go.mongodb.org/mongo-driver/mongo/readpref"
		"go.mongodb.org/mongo-driver/bson"
)

func hello(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "hello\n");
}

func headers(w http.ResponseWriter, req *http.Request) {
	for name, headers := range req.Header {
		for _, h:= range headers {
			fmt.Fprintf(w, "%v: %v\n", name, h)
		}
	}
}


type Log struct {
		Entry string
		CreatedAt time.Time
}

func main() {
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, connectErr := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if connectErr != nil {
	  log.Fatal("Could not connect", connectErr)
	}
	err := client.Ping(ctx, readpref.Primary())
	if err != nil {
          log.Fatal("Couldn't connect to the database", err)
	} else {
	  log.Println("Connected!")
	}
	db := client.Database("jonapi")
	collection := db.Collection("logs")
	firstLog := Log{"Made this entry", time.Now()}

	// Insert a single document
	insertResult, err := collection.InsertOne(context.TODO(), firstLog)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Inserted a single document: ", insertResult.InsertedID)
	

	http.HandleFunc("/hello", hello);

	http.HandleFunc("headers", headers);
	fs := http.FileServer(http.Dir("static"))
	http.HandleFunc("/logs", func(w http.ResponseWriter, req *http.Request) {
		var results []*Log
		findOptions := options.Find()
		findOptions.SetLimit(10)

		cur, err := collection.Find(context.TODO(), bson.D{{}}, findOptions)
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

		logsJson, err := json.Marshal(results)

		w.Header().Set("Content-Type", "application/json")
  	w.Write(logsJson)
	})
	http.Handle("/", fs)
	fmt.Println("Listening at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}
