package api

import (
	"context"
	"encoding/json"
	"fmt"
	"golog/model"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/crypto/bcrypt"
)

var db *mongo.Database
var logs *mongo.Collection
var users *mongo.Collection

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
	users = db.Collection("users")
}

// Log is the type for collection item
type Log struct {
	ID    *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Entry string              `json:"entry"`
}

// CreateLogHandler makes a log
func CreateLogHandler(w http.ResponseWriter, r *http.Request) {
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

// UpdateLogHandler updates a log
func UpdateLogHandler(w http.ResponseWriter, r *http.Request) {
	var logEntry Log
	vars := mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}
	err := json.NewDecoder(r.Body).Decode(&logEntry)
	if err != nil {
		log.Fatal("Invalid params", err)
	}

	update := bson.D{
		{"$set", logEntry},
	}

	updateResult, err := logs.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, err := json.Marshal(updateResult)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

// GetLogHandler retrieves a log by using the route param
func GetLogHandler(w http.ResponseWriter, r *http.Request) {
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

// DeleteLogHandler removes log by id
func DeleteLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
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

// GetLogsHandler retrieves logs from the database as json
func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
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

// Auth
func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	var res model.ResponseResult
	if err != nil {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	var result model.User
	err = users.FindOne(context.TODO(), bson.D{{"username", user.Username}}).Decode(&result)

	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

			if err != nil {
				res.Error = "Error While Hashing Password, Try Again"
				json.NewEncoder(w).Encode(res)
				return
			}
			user.Password = string(hash)

			_, err = users.InsertOne(context.TODO(), user)
			if err != nil {
				res.Error = "Error While Creating User, Try Again"
				json.NewEncoder(w).Encode(res)
				return
			}
			res.Result = "Registration Successful"
			json.NewEncoder(w).Encode(res)
			return
		}

		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

	res.Result = "Username already Exists!!"
	json.NewEncoder(w).Encode(res)
	return
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	var user model.User
	body, _ := ioutil.ReadAll(r.Body)
	err := json.Unmarshal(body, &user)
	if err != nil {
		log.Fatal(err)
	}

	var result model.User
	var res model.ResponseResult

	err = users.FindOne(context.TODO(), bson.D{{"username", user.Username}}).Decode(&result)

	if err != nil {
		res.Error = "Invalid username"
		json.NewEncoder(w).Encode(res)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))

	if err != nil {
		res.Error = "Invalid password"
		json.NewEncoder(w).Encode(res)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":  result.Username,
		"firstname": result.FirstName,
		"lastname":  result.LastName,
	})

	tokenString, err := token.SignedString([]byte("jonapi"))

	if err != nil {
		res.Error = "Error while generating token,Try again"
		json.NewEncoder(w).Encode(res)
		return
	}

	result.Token = tokenString
	result.Password = ""

	json.NewEncoder(w).Encode(result)

}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	tokenString := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("jonapi"), nil
	})
	var result model.User
	var res model.ResponseResult
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result.Username = claims["username"].(string)
		result.FirstName = claims["firstname"].(string)
		result.LastName = claims["lastname"].(string)

		json.NewEncoder(w).Encode(result)
		return
	} else {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

}
