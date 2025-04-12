package main

import (
	"fmt"
	"log"
	"time"

	"archive/app"
	"archive/gateway"
	"archive/repository"
)

func main() {
	gateway := gateway.NewFileSystemPlaylistGateway("/stream")
	repository := repository.NewFileSystemArchiveRepository("/archive")
	archiveApp := app.NewArchiveApp(gateway, repository)

	// Create a ticker that runs every minute
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	// Run immediately on startup
	doArchive(archiveApp)

	// Then run every minute
	for range ticker.C {
		doArchive(archiveApp)
	}
}

func doArchive(archiveApp *app.ArchiveApp) {
	fmt.Println("Starting archive...")
	result := archiveApp.Archive()
	if result.Error != nil {
		log.Printf("Archive failed: %v\n", result.Error)
		return
	}
	log.Printf("Successfully archived %d segments\n", result.ArchivedSegments)
}
