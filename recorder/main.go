package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"recorder/ffmpeg"
	"recorder/middleware"
)

//go:embed index.html
var content embed.FS

func main() {
	// Get port from environment variable, default to 6001 if not set
	port := os.Getenv("PORT")
	if port == "" {
		port = "6001"
	}

	// Check if RECORDING_URL is set
	recordingURL := os.Getenv("RECORDING_URL")
	if recordingURL == "" {
		log.Fatal("RECORDING_URL environment variable is not set")
	}

	// Set output directory
	outputDir := "/recorder/recordings"

	// Create and start the FFmpeg recorder
	recorder := ffmpeg.New(recordingURL, outputDir)
	if err := recorder.Start(); err != nil {
		log.Fatalf("Failed to start recorder: %v", err)
	}

	// Create a channel to listen for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to handle server errors
	serverErrChan := make(chan error, 1)

	// Create a new mux router
	mux := http.NewServeMux()

	// Register routes with their handlers
	mux.Handle("/", middleware.NoCache(rootHandler()))
	mux.Handle("/videos/", http.StripPrefix("/videos", middleware.NoCache(videoHandler(outputDir))))

	// Start the server in a goroutine
	serverAddr := fmt.Sprintf(":%s", port)
	log.Printf("Starting server on port %s", port)
	go func() {
		if err := http.ListenAndServe(serverAddr, mux); err != nil {
			serverErrChan <- err
		}
	}()

	// Wait for either a shutdown signal or server error
	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating shutdown...", sig)
	case err := <-serverErrChan:
		log.Printf("Server error: %v, initiating shutdown...", err)
	}

	// Stop the recorder
	if err := recorder.Stop(); err != nil {
		log.Printf("Error stopping recorder: %v", err)
	}

	log.Println("Shutdown complete")
}

// rootHandler returns a handler that serves the embedded index.html file
func rootHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		indexContent, err := content.ReadFile("index.html")
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		w.Write(indexContent)
	})
}

// videoHandler returns a handler that serves video files from the specified directory
func videoHandler(outputDir string) http.Handler {
	return http.FileServer(http.Dir(outputDir))
}
