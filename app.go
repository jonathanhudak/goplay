package main

import (
	"fmt"
	"golog/api"
	"golog/spa"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/logs/create", api.CreateLog).Methods("POST")
	r.HandleFunc("/logs", api.GetLogs).Methods("POST")
	r.HandleFunc("/logs/{_id}", api.GetLog).Methods("GET")
	r.HandleFunc("/logs/{_id}", api.UpdateLog).Methods("PUT")
	r.HandleFunc("/logs/{_id}", api.DeleteLog).Methods("DELETE")
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
