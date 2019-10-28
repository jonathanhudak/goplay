package main

import (
	"fmt"
	"golog/api"
	"golog/spa"
	"log"
	"net/http"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/urfave/negroni"
)

func main() {
	r := mux.NewRouter()

	apiRouter := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)

	apiRouter.HandleFunc("/profile", api.ProfileHandler).Methods("GET")
	apiRouter.HandleFunc("/logs/create", api.CreateLogHandler).Methods("POST")
	apiRouter.HandleFunc("/logs", api.GetLogsHandler).Methods("POST")
	apiRouter.HandleFunc("/logs/{_id}", api.GetLogHandler).Methods("GET")
	apiRouter.HandleFunc("/logs/{_id}", api.UpdateLogHandler).Methods("PUT")
	apiRouter.HandleFunc("/logs/{_id}", api.DeleteLogHandler).Methods("DELETE")

	r.HandleFunc("/register", api.RegisterHandler).Methods("POST")
	r.HandleFunc("/login", api.LoginHandler).Methods("POST")

	// Middleware: https://github.com/urfave/negroni

	jwtMiddleware := jwtmiddleware.New(jwtmiddleware.Options{
		Debug: true,
		ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
			return []byte("jonapi"), nil
		},
		// When set, the middleware verifies that tokens are signed with the specific signing algorithm
		// If the signing method is not constant the ValidationKeyGetter callback can be used to implement additional checks
		// Important to avoid security issues described here: https://auth0.com/blog/2015/03/31/critical-vulnerabilities-in-json-web-token-libraries/
		SigningMethod: jwt.SigningMethodHS256,
	})

	n := negroni.New(
		negroni.HandlerFunc(jwtMiddleware.HandlerWithNext),
		negroni.Wrap(apiRouter))

	r.PathPrefix("/api").Handler(n)
	r.PathPrefix("/").Handler(spa.CreateSpa("static", "index.html"))

	srv := &http.Server{
		Handler: r,
		Addr:    "127.0.0.1:8000",
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Println("Listening on: http://localhost:8000")

	log.Fatal(srv.ListenAndServe())
}
