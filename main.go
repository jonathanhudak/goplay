package main

import (
	"fmt"
	"goplay/api"
	"log"
	"net/http"
	"os"
	"time"

	jwtmiddleware "github.com/auth0/go-jwt-middleware"
	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"github.com/urfave/negroni"
)

func main() {
	r := mux.NewRouter()
	r.Use(mux.CORSMethodMiddleware(r))

	authenticatedRouter := mux.NewRouter().PathPrefix("/api").Subrouter().StrictSlash(true)

	authenticatedRouter.HandleFunc("/profile", api.ProfileHandler).Methods(http.MethodGet, http.MethodOptions)
	authenticatedRouter.HandleFunc("/logs/create", api.CreateLogHandler).Methods(http.MethodPost, http.MethodOptions, http.MethodGet, http.MethodOptions)
	authenticatedRouter.HandleFunc("/logs", api.GetLogsHandler).Methods(http.MethodPost, http.MethodOptions, http.MethodGet, http.MethodOptions)
	authenticatedRouter.HandleFunc("/logs/{_id}", api.GetLogHandler).Methods(http.MethodGet, http.MethodOptions)
	authenticatedRouter.HandleFunc("/logs/{_id}", api.UpdateLogHandler).Methods(http.MethodPut, http.MethodOptions)
	authenticatedRouter.HandleFunc("/logs/{_id}", api.DeleteLogHandler).Methods(http.MethodDelete, http.MethodOptions)

	r.HandleFunc("/register", api.RegisterHandler).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/login", api.LoginHandler).Methods(http.MethodPost, http.MethodOptions)

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
		negroni.Wrap(authenticatedRouter))

	r.PathPrefix("/api").Handler(n)

	port := os.Getenv("SERVER_PORT")

	if len(port) == 0 {
		port = "5000"
	}

	fmt.Println("Listening on: http://localhost:" + port)

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080", "http://frontend:8080"},
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{"GET", "PUT", "POST", "DELETE"},
		Debug:            false,
	})

	srv := &http.Server{
		Handler: c.Handler(r),
		Addr:    ":" + port,
		// Good practice: enforce timeouts for servers you create!
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	log.Fatal(srv.ListenAndServe())
}
