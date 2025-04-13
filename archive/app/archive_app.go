package app

import (
	"fmt"
	"io"
	"time"

	"archive/playlist"
)

// ArchiveRepository defines the interface for archive storage operations
type ArchiveRepository interface {
	// ReadPlaylist reads a playlist for a specific time
	ReadPlaylist(time time.Time) (*playlist.Playlist, error)

	// WritePlaylist writes a playlist for a specific time
	WritePlaylist(time time.Time, playlist *playlist.Playlist) error

	// WriteSegment writes a segment file for a specific time
	WriteSegment(time time.Time, filename string, content io.ReadCloser) error
}

// StreamRepository defines the interface for reading a playlist and segments
type StreamRepository interface {
	// GetPlaylist reads the playlist for the stream
	GetPlaylist() (*playlist.Playlist, error)
	// GetSegment reads a segment from the stream
	GetSegment(filename string) (io.ReadCloser, error)
}

// ArchiveApp coordinates the archive process
type ArchiveApp struct {
	streamRepo  StreamRepository
	archiveRepo ArchiveRepository
}

// NewArchiveApp creates a new ArchiveApp
func NewArchiveApp(streamRepo StreamRepository, archiveRepo ArchiveRepository) *ArchiveApp {
	return &ArchiveApp{
		streamRepo:  streamRepo,
		archiveRepo: archiveRepo,
	}
}

// ArchiveResult represents the result of an archive operation
type ArchiveResult struct {
	Error            error
	ArchivedSegments int
}

// Archive performs the archive operation
func (app *ArchiveApp) Archive() ArchiveResult {
	recorderPlaylist, err := app.streamRepo.GetPlaylist()
	if err != nil {
		return ArchiveResult{Error: fmt.Errorf("failed to get recorder playlist: %w", err)}
	}

	backedUp := 0
	for _, segment := range recorderPlaylist.Segments {
		// Get archive playlist for this segment's time
		archivePlaylist, err := app.archiveRepo.ReadPlaylist(segment.DateTime)
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
		content, err := app.streamRepo.GetSegment(segment.Filename)
		if err != nil {
			fmt.Printf("Failed to get segment %s: %v\n", segment.Filename, err)
			continue
		}

		// Write segment to archive
		newFilename := fmt.Sprintf("segment_%03d.ts", len(archivePlaylist.Segments))
		if err := app.archiveRepo.WriteSegment(segment.DateTime, newFilename, content); err != nil {
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
		if err := app.archiveRepo.WritePlaylist(segment.DateTime, archivePlaylist); err != nil {
			fmt.Printf("Failed to write archive playlist for segment %s: %v\n", newFilename, err)
			continue
		}

		backedUp++
	}

	fmt.Printf("Archive complete. Archived %d segments.\n", backedUp)
	return ArchiveResult{ArchivedSegments: backedUp}
}
