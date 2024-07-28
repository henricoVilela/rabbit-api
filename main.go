package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"henricovilela.com/rabbit_api/database"
	"henricovilela.com/rabbit_api/rabbit"
)

func SendNotification(w http.ResponseWriter, r *http.Request) {
	var notification rabbit.Notification

	// Decode the JSON request body into the notification struct
	if err := json.NewDecoder(r.Body).Decode(&notification); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err := rabbit.SendMessage(notification)

	// Create the audit log entry
	auditLog := database.AuditLog{
		UserId:      notification.UserId,
		Application: notification.Application,
		Message:     notification.Message,
		Timestamp:   time.Now(),
		IP:          r.RemoteAddr,
		UserAgent:   r.UserAgent(),
		Success:     err == nil,
	}

	// Start a goroutine to insert the audit log into MongoDB asynchronously
	go func() {
		if err := database.InsertAuditLog(auditLog); err != nil {
			log.Printf("Failed to save audit log: %v", err)
		}
	}()

	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{"erro": err.Error()})
		return
	}

	// Send a response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func ListAuditLogs(w http.ResponseWriter, r *http.Request) {
	logs, err := database.GetAuditLogs()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func main() {
	var err error

	// Initialize MongoDB connection
	if err = database.ConnectDB(); err != nil {
		log.Fatal(err)
	}

	// Create a new mux router
	r := mux.NewRouter()

	// Register the /send endpoint with the SendNotification handler
	r.HandleFunc("/send", SendNotification).Methods("POST")
	r.HandleFunc("/auditlogs", ListAuditLogs).Methods("GET")
	// Start the HTTP server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		database.DisconnectDB()
		rabbit.Disconnect()
		log.Fatal(err)
	}
}
