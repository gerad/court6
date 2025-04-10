package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
)

//go:embed index.html
var content embed.FS

// noCacheFileServer wraps a http.FileServer and adds no-cache headers
type noCacheFileServer struct {
	handler http.Handler
}

func (n noCacheFileServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Set headers to prevent caching
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	// Serve the file
	n.handler.ServeHTTP(w, r)
}

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
	playlistName := "playlist.m3u8"

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	// Start FFmpeg process
	ffmpegCmd := exec.Command("ffmpeg",
		"-y",
		"-i", recordingURL,
		"-c", "copy",
		"-f", "hls",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-hls_segment_filename", fmt.Sprintf("%s/segment_%%03d.ts", outputDir),
		fmt.Sprintf("%s/%s", outputDir, playlistName),
	)

	ffmpegCmd.Stdout = os.Stdout
	ffmpegCmd.Stderr = os.Stderr

	if err := ffmpegCmd.Start(); err != nil {
		log.Fatalf("Failed to start FFmpeg: %v", err)
	}

	// Create a channel to listen for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create a channel to handle server errors
	serverErrChan := make(chan error, 1)

	// Create a new mux router
	mux := http.NewServeMux()

	// Serve embedded index.html for the root path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
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

	// Serve video content under /videos/ using FileServer with no caching
	fileServer := noCacheFileServer{http.FileServer(http.Dir(outputDir))}
	mux.Handle("/videos/", http.StripPrefix("/videos", fileServer))

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

	// Gracefully shutdown FFmpeg
	log.Println("Shutting down FFmpeg...")
	if err := ffmpegCmd.Process.Signal(syscall.SIGTERM); err != nil {
		log.Printf("Error sending SIGTERM to FFmpeg: %v", err)
		// If SIGTERM fails, try SIGKILL
		if err := ffmpegCmd.Process.Kill(); err != nil {
			log.Printf("Error killing FFmpeg process: %v", err)
		}
	}

	// Wait for FFmpeg to finish
	if err := ffmpegCmd.Wait(); err != nil {
		log.Printf("FFmpeg process exited with error: %v", err)
	}

	log.Println("Shutdown complete")
}
