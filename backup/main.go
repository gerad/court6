package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"backup/app"
	"backup/gateway"
	"backup/repository"
)

type BackupResponse struct {
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "6002"
	}

	recorderURL := os.Getenv("RECORDER_URL")
	if recorderURL == "" {
		log.Fatal("RECORDER_URL environment variable is not set")
	}

	// Create dependencies
	gateway := gateway.NewHTTPPlaylistGateway(recorderURL)
	repository := repository.NewFileSystemBackupRepository("/backup")
	backupApp := app.NewBackupApp(gateway, repository)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /backup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			sendJSON(w, http.StatusMethodNotAllowed, BackupResponse{Error: "Method not allowed"})
			return
		}

		result := backupApp.Backup()
		if result.Error != nil {
			sendJSON(w, http.StatusInternalServerError, BackupResponse{Error: result.Error.Error()})
			return
		}

		sendJSON(w, http.StatusOK, BackupResponse{
			Message: fmt.Sprintf("Successfully backed up %d segments", result.BackedUpSegments),
		})
	})

	fmt.Printf("Starting backup server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
