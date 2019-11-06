package api

import (
	"context"
	"encoding/json"
	"goplay/database"
	"goplay/model"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// CreateLogHandler creates a log owned by the requester
func CreateLogHandler(w http.ResponseWriter, r *http.Request) {
	var logEntry model.Log
	owner, _, _ := getUserFromAuthToken(r)

	logEntry.UserID = owner.OID

	err := json.NewDecoder(r.Body).Decode(&logEntry)
	if err != nil {
		log.Fatal("Invalid params", err)
	}

	createLogResult := database.CreateLog(logEntry)

	resultJSON, err := json.Marshal(createLogResult)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

// CreateHabitHandler creates a habit owned by the requester
func CreateHabitHandler(w http.ResponseWriter, r *http.Request) {
	var habit model.Habit
	owner, _, _ := getUserFromAuthToken(r)

	habit.UserID = owner.OID

	err := json.NewDecoder(r.Body).Decode(&habit)
	if err != nil {
		log.Fatal("Invalid params", err)
	}

	result := database.CreateHabit(habit)

	json, err := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

// CreateIdentityHandler creates an identity owned by the requester
func CreateIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var identity model.Identity
	owner, _, _ := getUserFromAuthToken(r)

	identity.UserID = owner.OID

	err := json.NewDecoder(r.Body).Decode(&identity)
	if err != nil {
		log.Fatal("Invalid params", err)
	}

	result := database.CreateIdentity(identity)

	json, err := json.Marshal(result)
	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
}

// GetIdentitiesHandler retrieves identities from the database as json
func GetIdentitiesHandler(w http.ResponseWriter, r *http.Request) {
	owner, _, _ := getUserFromAuthToken(r)

	identitiesJSON, err := json.Marshal(database.GetIdentities(owner.OID))

	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(identitiesJSON)
}

// DeleteIdentityHandler removes an identity by id if the requester is the owner
func DeleteIdentityHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}

	if ensureIdentityOwner(w, r, filter) == false {
		return
	}

	deleteResult, err := database.Identities.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, _ := json.Marshal(deleteResult)

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

	deleteResult, err := database.Logs.DeleteOne(context.TODO(), filter)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, _ := json.Marshal(deleteResult)

	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

func DeleteHabitHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}

	if ensureHabitOwner(w, r, filter) == false {
		return
	}

	deleteResult, err := database.Habits.DeleteOne(context.TODO(), filter)
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
	findErr := database.Logs.FindOne(context.TODO(), filter).Decode(&logEntry)
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

func ensureHabitOwner(w http.ResponseWriter, r *http.Request, filter bson.D) bool {
	var habit model.Habit

	// Make sure the habit is owned by the same user
	owner, _, _ := getUserFromAuthToken(r)
	findErr := database.Habits.FindOne(context.TODO(), filter).Decode(&habit)
	if findErr != nil {
		log.Fatal("Habit not found", findErr)
	}

	if habit.UserID != owner.OID {
		w.WriteHeader(http.StatusForbidden)
		notFound, _ := json.Marshal("Computer says no")
		w.Write(notFound)
		return false
	}
	return true
}

func ensureIdentityOwner(w http.ResponseWriter, r *http.Request, filter bson.D) bool {
	var identity model.Identity

	// Make sure the identity is owned by the same user
	owner, _, _ := getUserFromAuthToken(r)
	findErr := database.Identities.FindOne(context.TODO(), filter).Decode(&identity)
	if findErr != nil {
		log.Fatal("Identity not found", findErr)
	}

	if identity.UserID != owner.OID {
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

	if ensureHabitOwner(w, r, filter) == false {
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

	updateResult, err := database.Logs.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, err := json.Marshal(updateResult)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

// UpdateHabitHandler updates a habit if the requester is the owner
func UpdateHabitHandler(w http.ResponseWriter, r *http.Request) {
	var habit model.Habit
	vars := mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}

	if ensureHabitOwner(w, r, filter) == false {
		return
	}

	err := json.NewDecoder(r.Body).Decode(&habit)
	if err != nil {
		log.Fatal("Invalid params", err)
		return
	}

	update := bson.D{
		{"$set", habit},
	}

	updateResult, err := database.Habits.UpdateOne(context.TODO(), filter, update)
	if err != nil {
		log.Fatal(err)
	}
	resultJSON, err := json.Marshal(updateResult)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resultJSON)
}

// UpdateIdentityHandler updates an identity if the requester is the owner
func UpdateIdentityHandler(w http.ResponseWriter, r *http.Request) {
	var identity model.Identity
	vars := mux.Vars(r)
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	filter := bson.D{{"_id", objID}}

	if ensureIdentityOwner(w, r, filter) == false {
		return
	}

	err := json.NewDecoder(r.Body).Decode(&identity)
	if err != nil {
		log.Fatal("Invalid params", err)
		return
	}

	update := bson.D{
		{"$set", identity},
	}

	updateResult, err := database.Identities.UpdateOne(context.TODO(), filter, update)
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
	objID, _ := primitive.ObjectIDFromHex(vars["_id"])
	logEntry := database.GetLog(objID, owner.OID)
	resultJSON, _ := json.Marshal(logEntry)
	w.Write(resultJSON)
}

// GetLogsHandler retrieves logs from the database as json
func GetLogsHandler(w http.ResponseWriter, r *http.Request) {
	owner, _, _ := getUserFromAuthToken(r)

	logsJSON, err := json.Marshal(database.GetLogs(owner.OID))

	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(logsJSON)
}

// GetHabitsHandler returns the owners habits
func GetHabitsHandler(w http.ResponseWriter, r *http.Request) {
	owner, _, _ := getUserFromAuthToken(r)
	json, err := json.Marshal(database.GetHabits(owner.OID))

	if err != nil {
		log.Fatal(err)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(json)
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
	err = database.Users.FindOne(context.TODO(), bson.D{{"username", user.Username}}).Decode(&result)

	if err != nil {
		if err.Error() == "mongo: no documents in result" {
			hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)

			if err != nil {
				res.Error = "Error While Hashing Password, Try Again"
				json.NewEncoder(w).Encode(res)
				return
			}
			user.Password = string(hash)

			_, err = database.Users.InsertOne(context.TODO(), user)
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

	err = database.Users.FindOne(context.TODO(), bson.D{{"username", user.Username}}).Decode(&result)

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
		err := database.Users.FindOne(context.TODO(), filter).Decode(&user)

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
