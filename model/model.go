package model

import "go.mongodb.org/mongo-driver/bson/primitive"

// User
type User struct {
	OID       primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Username  string             `json:"username"`
	FirstName string             `json:"firstname"`
	LastName  string             `json:"lastname"`
	Password  string             `json:"password"`
	Token     string             `json:"token"`
}

// ResponseResult
type ResponseResult struct {
	Error  string `json:"error"`
	Result string `json:"result"`
}

// Log is the type for collection item
type Log struct {
	ID     *primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Entry  string              `json:"entry"`
	UserID primitive.ObjectID  `json:"user_id" bson:"user_id,omitempty"`
}
