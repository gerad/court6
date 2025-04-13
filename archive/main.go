package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"archive/app"
	"archive/archiverepo"
	"archive/streamrepo"
)

func main() {
	inputDir, found := os.LookupEnv("INPUT_DIR")
	if !found {
		log.Fatalln("Error: INPUT_DIR environment variable is not set")
	}
	streamRepo := streamrepo.New(inputDir)

	outputDir, found := os.LookupEnv("OUTPUT_DIR")
	if !found {
		log.Fatalln("Error: OUTPUT_DIR environment variable is not set")
	}
	archiveRepo := archiverepo.New(outputDir)

	archiveApp := app.NewArchiveApp(streamRepo, archiveRepo)

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
