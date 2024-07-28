package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type Notification struct {
	UserId      string `json:"userId"`
	Application string `json:"application"`
	Message     string `json:"message"`
}

func SendNotification(w http.ResponseWriter, r *http.Request) {
	var notification Notification

	// Decode the JSON request body into the notification struct
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Log the received notification
	log.Printf("Received notification: %+v", notification)

	// Send a response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func main() {
	// Create a new mux router
	r := mux.NewRouter()

	// Register the /send endpoint with the SendNotification handler
	r.HandleFunc("/send", SendNotification).Methods("POST")

	// Start the HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
