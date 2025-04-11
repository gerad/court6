package app_test

import (
	"io"
	"strings"
	"testing"
	"time"

	"backup/app"
	"backup/playlist"
)

func TestBackup(t *testing.T) {
	// Create test data with timestamps
	now := time.Now()
	nextMinute := now.Add(time.Minute)

	recorderPlaylist := &playlist.Playlist{
		Version:        3,
		TargetDuration: 10,
		MediaSequence:  0,
		Segments: []playlist.Segment{
			{Filename: "segment1.ts", DateTime: now, Duration: 10},
			{Filename: "segment2.ts", DateTime: nextMinute, Duration: 10},
			// Add a duplicate segment with same DateTime but different filename
			{Filename: "segment3.ts", DateTime: now, Duration: 10},
		},
	}

	// Create mocks
	gateway := &MockPlaylistGateway{
		playlist: recorderPlaylist,
		segments: make(map[string]string),
	}
	repository := &MockBackupRepository{}

	// Create app and run backup
	backupApp := app.NewBackupApp(gateway, repository)
	result := backupApp.Backup()

	// Verify results
	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	// Verify written playlist
	timeKey := now.Format("2006-01-02-15")
	backupPlaylist, exists := repository.playlists[timeKey]
	if !exists {
		t.Fatal("Expected playlist to be written, but it wasn't")
	}

	// Verify playlist metadata was initialized from the gateway playlist
	if backupPlaylist.Version != recorderPlaylist.Version {
		t.Errorf("Expected version %d, got %d", recorderPlaylist.Version, backupPlaylist.Version)
	}
	if backupPlaylist.TargetDuration != recorderPlaylist.TargetDuration {
		t.Errorf("Expected target duration %d, got %d", recorderPlaylist.TargetDuration, backupPlaylist.TargetDuration)
	}
	if backupPlaylist.MediaSequence != recorderPlaylist.MediaSequence {
		t.Errorf("Expected media sequence %d, got %d", recorderPlaylist.MediaSequence, backupPlaylist.MediaSequence)
	}

	// We should have 2 segments in the backup (the duplicate should be skipped)
	if len(backupPlaylist.Segments) != 2 {
		t.Errorf("Expected 2 segments in playlist, got %d", len(backupPlaylist.Segments))
	}

	// Verify segments were written with sequential filenames
	expectedFilenames := []string{"segment_00.ts", "segment_01.ts"}
	for i, segment := range backupPlaylist.Segments {
		if segment.Filename != expectedFilenames[i] {
			t.Errorf("Expected segment %d to have filename %s, got %s", i, expectedFilenames[i], segment.Filename)
		}

		// Verify the segment was written to the repository
		segmentKey := timeKey + "/" + segment.Filename
		if _, exists := repository.segments[segmentKey]; !exists {
			t.Errorf("Expected segment %s to be written, but it wasn't", segment.Filename)
		}
	}
}

// MockPlaylistGateway implements gateway.PlaylistGateway for testing
type MockPlaylistGateway struct {
	playlist *playlist.Playlist
	segments map[string]string
}

func (m *MockPlaylistGateway) GetPlaylist() (*playlist.Playlist, error) {
	return m.playlist, nil
}

func (m *MockPlaylistGateway) GetSegment(filename string) (io.ReadCloser, error) {
	content, ok := m.segments[filename]
	if !ok {
		content = "mock segment content"
	}
	return io.NopCloser(strings.NewReader(content)), nil
}

// MockBackupRepository implements repository.BackupRepository for testing
type MockBackupRepository struct {
	playlists map[string]*playlist.Playlist
	segments  map[string]string
}

func (m *MockBackupRepository) ReadBackupPlaylist(segmentTime time.Time) (*playlist.Playlist, error) {
	if m.playlists == nil {
		m.playlists = make(map[string]*playlist.Playlist)
	}

	timeKey := segmentTime.Format("2006-01-02-15")
	p, exists := m.playlists[timeKey]
	if !exists {
		// Return nil if the playlist doesn't exist
		return nil, nil
	}
	return p, nil
}

func (m *MockBackupRepository) WriteBackupPlaylist(segmentTime time.Time, p *playlist.Playlist) error {
	if m.playlists == nil {
		m.playlists = make(map[string]*playlist.Playlist)
	}

	timeKey := segmentTime.Format("2006-01-02-15")
	m.playlists[timeKey] = p
	return nil
}

func (m *MockBackupRepository) WriteSegment(segmentTime time.Time, filename string, content io.Reader) error {
	if m.segments == nil {
		m.segments = make(map[string]string)
	}

	timeKey := segmentTime.Format("2006-01-02-15")
	segmentKey := timeKey + "/" + filename

	contentBytes, err := io.ReadAll(content)
	if err != nil {
		return err
	}

	m.segments[segmentKey] = string(contentBytes)
	return nil
}
