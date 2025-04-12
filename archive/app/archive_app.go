package app

import (
	"fmt"

	"archive/gateway"
	"archive/playlist"
	"archive/repository"
)

// ArchiveApp coordinates the archive process
type ArchiveApp struct {
	gateway    gateway.PlaylistGateway
	repository repository.ArchiveRepository
}

// NewArchiveApp creates a new ArchiveApp
func NewArchiveApp(gateway gateway.PlaylistGateway, repository repository.ArchiveRepository) *ArchiveApp {
	return &ArchiveApp{
		gateway:    gateway,
		repository: repository,
	}
}

// ArchiveResult represents the result of an archive operation
type ArchiveResult struct {
	Error            error
	ArchivedSegments int
}

// Archive performs the archive operation
func (app *ArchiveApp) Archive() ArchiveResult {
	recorderPlaylist, err := app.gateway.GetPlaylist()
	if err != nil {
		return ArchiveResult{Error: fmt.Errorf("failed to get recorder playlist: %w", err)}
	}

	backedUp := 0
	for _, segment := range recorderPlaylist.Segments {
		// Get archive playlist for this segment's time
		archivePlaylist, err := app.repository.ReadPlaylist(segment.DateTime)
		if err != nil {
			fmt.Printf("Failed to read archive playlist for segment %s: %v\n", segment.Filename, err)
			continue
		}

		if archivePlaylist == nil {
			fmt.Printf("Creating new archive playlist for time %s\n", segment.DateTime.Format("2006-01-02T15:04:05Z"))
			archivePlaylist = &playlist.Playlist{
				Version:        recorderPlaylist.Version,
				TargetDuration: recorderPlaylist.TargetDuration,
				MediaSequence:  recorderPlaylist.MediaSequence,
				Segments:       []playlist.Segment{},
			}
		}

		// Check if segment already exists in archive playlist based on DateTime
		segmentExists := false
		for _, existingSegment := range archivePlaylist.Segments {
			if existingSegment.DateTime.Equal(segment.DateTime) {
				segmentExists = true
				break
			}
		}

		if segmentExists {
			fmt.Printf("Segment with DateTime %s already exists in archive, skipping\n", segment.DateTime.Format("2006-01-02T15:04:05Z"))
			continue
		}

		// Get segment content from recorder
		content, err := app.gateway.GetSegment(segment.Filename)
		if err != nil {
			fmt.Printf("Failed to get segment %s: %v\n", segment.Filename, err)
			continue
		}

		// Write segment to archive
		newFilename := fmt.Sprintf("segment_%03d.ts", len(archivePlaylist.Segments))
		if err := app.repository.WriteSegment(segment.DateTime, newFilename, content); err != nil {
			fmt.Printf("Failed to write segment %s: %v\n", newFilename, err)
			continue
		}

		// Create new segment with updated filename
		newSegment := playlist.Segment{
			Filename: newFilename,
			Duration: segment.Duration,
			DateTime: segment.DateTime,
		}

		// Add segment to archive playlist
		archivePlaylist = playlist.Concat(archivePlaylist, newSegment)

		// Write updated playlist
		if err := app.repository.WritePlaylist(segment.DateTime, archivePlaylist); err != nil {
			fmt.Printf("Failed to write archive playlist for segment %s: %v\n", newFilename, err)
			continue
		}

		backedUp++
	}

	fmt.Printf("Archive complete. Archived %d segments.\n", backedUp)
	return ArchiveResult{ArchivedSegments: backedUp}
}
