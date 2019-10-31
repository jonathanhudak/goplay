package api

import (
	"context"
	"encoding/json"
	"fmt"
	"golog/model"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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

	db = client.Database("jonapi")
	logs = db.Collection("logs")
	users = db.Collection("users")
}

// CreateLogHandler creates a log owned by the requester
func CreateLogHandler(w http.ResponseWriter, r *http.Request) {
	var logEntry model.Log
	owner, _, _ := getUserFromAuthToken(r)

	logEntry.UserID = owner.OID

	err := json.NewDecoder(r.Body).Decode(&logEntry)
	if err != nil {
		log.Fatal("Invalid params", err)
	}

	insertResult, err := logs.InsertOne(context.TODO(), logEntry)
	if err != nil {
		log.Fatal(err)
	}

	resultJSON, err := json.Marshal(insertResult)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

// DeleteLogHandler removes log by id if the requester is the owner
func DeleteLogHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}

	if ensureLogOwner(w, r, filter) == false {
		return
	}

	deleteResult, err := logs.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, _ := json.Marshal(deleteResult)

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func ensureLogOwner(w http.ResponseWriter, r *http.Request, filter bson.D) bool {
	var logEntry model.Log

	// Make sure the log is owned by the same user
	owner, _, _ := getUserFromAuthToken(r)
	findErr := logs.FindOne(context.TODO(), filter).Decode(&logEntry)
	if findErr != nil {
		log.Fatal("Log not found", findErr)
	}

	if logEntry.UserID != owner.OID {
		w.WriteHeader(http.StatusForbidden)
		notFound, _ := json.Marshal("Computer says no")
		w.Write(notFound)
		return false
	}
	return true
}

// UpdateLogHandler updates a log if the requester is the owner
func UpdateLogHandler(w http.ResponseWriter, r *http.Request) {

	var logEntry model.Log
	vars := mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}

	if ensureLogOwner(w, r, filter) == false {
		return
	}

	err := json.NewDecoder(r.Body).Decode(&logEntry)
	if err != nil {
		log.Fatal("Invalid params", err)
		return
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
	owner, _, _ := getUserFromAuthToken(r)
	vars := mux.Vars(r)
	var logEntry model.Log
	fmt.Printf("GetLog %s\n", vars["_id"])
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}, {"user_id", owner.OID}} // vet complains about this. Not sure composite literal uses unkeyed fields
	err := logs.FindOne(context.TODO(), filter).Decode(&logEntry)
	if err != nil {
		log.Fatal(err)
		notFound, _ := json.Marshal("Log entry not found")
		w.Write(notFound)
	}
	resultJSON, err := json.Marshal(logEntry)
	w.Write(resultJSON)
}

// GetLogsHandler retrieves logs from the database as json
func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	owner, _, _ := getUserFromAuthToken(r)
	var results []*model.Log
	findOptions := options.Find()
	findOptions.SetLimit(10)

	cur, err := logs.Find(context.TODO(), bson.D{{"user_id", owner.OID}}, findOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Iterate through the cursor
	for cur.Next(context.TODO()) {
		var elem model.Log
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

func getUserFromAuthToken(r *http.Request) (model.User, bool, error) {
	tokenString := strings.Split(r.Header.Get("Authorization"), "Bearer ")[1]
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte("jonapi"), nil
	})
	var user model.User
	var ok bool
	if claims, _ := token.Claims.(jwt.MapClaims); token.Valid {
		user.Username = claims["username"].(string)
		user.FirstName = claims["firstname"].(string)
		user.LastName = claims["lastname"].(string)
		ok = true

		filter := bson.D{{"username", user.Username}}
		err := users.FindOne(context.TODO(), filter).Decode(&user)

		if err != nil {
			log.Fatal(err)
		}
	}
	return user, ok, err
}

// ProfileHandler return the user's profile encoded in the jwt token claims
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	user, ok, err := getUserFromAuthToken(r)
	var res model.ResponseResult
	if ok {
		json.NewEncoder(w).Encode(user)
	} else {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
	}
}
