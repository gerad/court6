package app

import (
	"fmt"

	"backup/gateway"
	"backup/playlist"
	"backup/repository"
)

// BackupApp coordinates the backup process
type BackupApp struct {
	gateway    gateway.PlaylistGateway
	repository repository.BackupRepository
}

// NewBackupApp creates a new BackupApp
func NewBackupApp(gateway gateway.PlaylistGateway, repository repository.BackupRepository) *BackupApp {
	return &BackupApp{
		gateway:    gateway,
		repository: repository,
	}
}

// BackupResult represents the result of a backup operation
type BackupResult struct {
	BackedUpSegments int
	Error            error
}

// Backup performs the backup operation
func (app *BackupApp) Backup() BackupResult {
	// Get recorder playlist
	recorderPlaylist, err := app.gateway.GetPlaylist()
	if err != nil {
		return BackupResult{Error: fmt.Errorf("failed to get recorder playlist: %w", err)}
	}

	// Process each segment in the recorder playlist
	backedUp := 0
	for _, segment := range recorderPlaylist.Segments {
		// Skip segments without a valid DateTime
		if segment.DateTime.IsZero() {
			fmt.Printf("Skipping segment %s: no DateTime available\n", segment.Filename)
			continue
		}

		// Get backup playlist for this segment's time
		backupPlaylist, err := app.repository.ReadBackupPlaylist(segment.DateTime)
		if err != nil {
			fmt.Printf("Failed to read backup playlist for segment %s: %v\n", segment.Filename, err)
			continue
		}

		// Initialize a new playlist if it doesn't exist
		if backupPlaylist == nil {
			backupPlaylist = &playlist.Playlist{
				Version:        recorderPlaylist.Version,
				TargetDuration: recorderPlaylist.TargetDuration,
				MediaSequence:  recorderPlaylist.MediaSequence,
				Segments:       make([]playlist.Segment, 0),
			}
		}

		// Check if segment already exists in backup playlist based on DateTime
		segmentExists := false
		for _, existingSegment := range backupPlaylist.Segments {
			if existingSegment.DateTime.Equal(segment.DateTime) {
				segmentExists = true
				break
			}
		}

		if segmentExists {
			fmt.Printf("Segment with DateTime %s already exists in backup, skipping\n", segment.DateTime.Format("2006-01-02T15:04:05Z"))
			continue
		}

		// Get segment content
		content, err := app.gateway.GetSegment(segment.Filename)
		if err != nil {
			fmt.Printf("Failed to download segment %s: %v\n", segment.Filename, err)
			continue
		}

		// Generate a new filename based on the number of existing segments
		newFilename := fmt.Sprintf("segment_%02d.ts", len(backupPlaylist.Segments))

		// Write segment to filesystem with the new filename
		if err := app.repository.WriteSegment(segment.DateTime, newFilename, content); err != nil {
			content.Close()
			fmt.Printf("Failed to write segment file %s: %v\n", newFilename, err)
			continue
		}
		content.Close()

		// Create a new segment with the updated filename
		newSegment := segment
		newSegment.Filename = newFilename

		// Add segment to backup playlist
		backupPlaylist = playlist.Concat(backupPlaylist, newSegment)
		backedUp++

		// Write updated playlist
		if err := app.repository.WriteBackupPlaylist(segment.DateTime, backupPlaylist); err != nil {
			fmt.Printf("Failed to write backup playlist for segment %s: %v\n", newFilename, err)
			continue
		}
	}

	return BackupResult{BackedUpSegments: backedUp}
}
