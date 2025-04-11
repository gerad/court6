package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"backup/playlist"
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

	mux := http.NewServeMux()
	mux.HandleFunc("POST /backup", func(w http.ResponseWriter, r *http.Request) {
		handleBackup(w, r)
	})

	fmt.Printf("Starting backup server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, mux); err != nil {
		fmt.Printf("Server error: %v\n", err)
	}
}

func handleBackup(w http.ResponseWriter, _ *http.Request) {
	recorderURL := os.Getenv("RECORDER_URL")
	if recorderURL == "" {
		log.Fatal("RECORDER_URL environment variable is not set")
	}

	playlistURL := fmt.Sprintf("%s/videos/playlist.m3u8", recorderURL)
	resp, err := http.Get(playlistURL)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, BackupResponse{Error: "Failed to fetch playlist"})
		return
	}
	defer resp.Body.Close()

	recorderPlaylist, err := playlist.Parse(resp.Body)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, BackupResponse{Error: "Failed to parse playlist"})
		return
	}

	backupPath := filepath.Join("volumes", "backup", "videos",
		fmt.Sprintf("%d", time.Now().Year()),
		fmt.Sprintf("%02d", time.Now().Month()),
		fmt.Sprintf("%02d", time.Now().Day()),
		fmt.Sprintf("%02d", time.Now().Hour()))

	if err := os.MkdirAll(backupPath, 0755); err != nil {
		sendJSON(w, http.StatusInternalServerError, BackupResponse{Error: "Failed to create backup directory"})
		return
	}

	// Read existing backup playlist if it exists
	backupPlaylistPath := filepath.Join(backupPath, "playlist.m3u8")
	var backupPlaylist *playlist.Playlist
	if content, err := os.ReadFile(backupPlaylistPath); err == nil {
		backupPlaylist, err = playlist.Parse(strings.NewReader(string(content)))
		if err != nil {
			sendJSON(w, http.StatusInternalServerError, BackupResponse{Error: "Failed to parse backup playlist"})
			return
		}
	} else {
		backupPlaylist = &playlist.Playlist{
			Version:        3,
			TargetDuration: 10,
			MediaSequence:  0,
			Segments:       make([]playlist.Segment, 0),
		}
	}

	// Find new segments to backup
	newSegments := playlist.FindNewSegments(recorderPlaylist, backupPlaylist)
	backedUp := 0

	for _, segment := range newSegments {
		// Download and save segment
		segmentURL := fmt.Sprintf("%s/videos/%s", recorderURL, segment.Filename)
		resp, err := http.Get(segmentURL)
		if err != nil {
			fmt.Printf("Failed to download segment %s: %v\n", segment.Filename, err)
			continue
		}

		// Create new segment name with sequential numbering
		newSegmentName := fmt.Sprintf("segment_%02d.ts", len(backupPlaylist.Segments)+1)
		segmentPath := filepath.Join(backupPath, newSegmentName)

		out, err := os.Create(segmentPath)
		if err != nil {
			resp.Body.Close()
			fmt.Printf("Failed to create segment file %s: %v\n", newSegmentName, err)
			continue
		}

		_, err = io.Copy(out, resp.Body)
		resp.Body.Close()
		out.Close()
		if err != nil {
			fmt.Printf("Failed to write segment file %s: %v\n", newSegmentName, err)
			continue
		}

		// Update segment filename and add to backup playlist
		segment.Filename = newSegmentName
		backupPlaylist = playlist.Concat(backupPlaylist, segment)
		backedUp++
	}

	// Write updated playlist
	if err := os.WriteFile(backupPlaylistPath, []byte(backupPlaylist.String()), 0644); err != nil {
		sendJSON(w, http.StatusInternalServerError, BackupResponse{Error: "Failed to write backup playlist"})
		return
	}

	sendJSON(w, http.StatusOK, BackupResponse{
		Message: fmt.Sprintf("Successfully backed up %d segments", backedUp),
	})
}

func sendJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
