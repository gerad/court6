package app_test

import (
	"archive/app"
	"archive/playlist"
	"bytes"
	"errors"
	"fmt"
	"io"
	"testing"
	"time"
)

func TestArchiveApp_Archive_Successful(t *testing.T) {
	// Setup
	now := time.Now()
	streamRepo := &mockStreamRepo{
		playlist: &playlist.Playlist{
			Version:        3,
			TargetDuration: 10,
			MediaSequence:  0,
			Segments: []playlist.Segment{
				{
					Filename:        "segment_00.ts",
					Duration:        10,
					DateTime:        now,
					ProgramDateTime: "",
				},
			},
		},
		segment: []byte("test segment"),
	}
	archiveRepo := &mockArchiveRepo{}

	// Execute
	app := app.NewArchiveApp(streamRepo, archiveRepo)
	result := app.Archive()

	// Assert
	expectedSegments := 1
	if result.ArchivedSegments != expectedSegments {
		t.Errorf("ArchivedSegments = %v, want %v", result.ArchivedSegments, expectedSegments)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	// Verify the playlist was written to the archive
	if archiveRepo.playlist == nil {
		t.Error("Expected playlist to be written to archive, got nil")
	}
}

func TestArchiveApp_Archive_MultipleSegments(t *testing.T) {
	// Setup
	now := time.Now()
	segment1Time := now
	segment2Time := now.Add(10 * time.Second)
	segment3Time := now.Add(20 * time.Second)

	streamRepo := &mockStreamRepo{
		playlist: &playlist.Playlist{
			Version:        3,
			TargetDuration: 10,
			MediaSequence:  0,
			Segments: []playlist.Segment{
				{
					Filename:        "segment_00.ts",
					Duration:        10,
					DateTime:        segment1Time,
					ProgramDateTime: segment1Time.Format("2006-01-02T15:04:05Z"),
				},
				{
					Filename:        "segment_01.ts",
					Duration:        10,
					DateTime:        segment2Time,
					ProgramDateTime: segment2Time.Format("2006-01-02T15:04:05Z"),
				},
				{
					Filename:        "segment_02.ts",
					Duration:        10,
					DateTime:        segment3Time,
					ProgramDateTime: segment3Time.Format("2006-01-02T15:04:05Z"),
				},
			},
		},
		segment: []byte("test segment"),
	}
	archiveRepo := &mockArchiveRepo{}

	// Execute
	app := app.NewArchiveApp(streamRepo, archiveRepo)
	result := app.Archive()

	// Assert
	expectedSegments := 3
	if result.ArchivedSegments != expectedSegments {
		t.Errorf("ArchivedSegments = %v, want %v", result.ArchivedSegments, expectedSegments)
	}

	if result.Error != nil {
		t.Errorf("Expected no error, got %v", result.Error)
	}

	// Verify the playlist was written to the archive
	if archiveRepo.playlist == nil {
		t.Error("Expected playlist to be written to archive, got nil")
	}

	// Verify all segments were archived
	if len(archiveRepo.playlist.Segments) != expectedSegments {
		t.Errorf("Expected %d segments in archive playlist, got %d", expectedSegments, len(archiveRepo.playlist.Segments))
	}

	// Verify program date time tags are preserved
	for i, segment := range archiveRepo.playlist.Segments {
		if segment.ProgramDateTime == "" {
			t.Errorf("Segment %d missing ProgramDateTime", i)
		}
	}

	// Verify segment filenames are sequential
	for i, segment := range archiveRepo.playlist.Segments {
		expectedFilename := fmt.Sprintf("segment_%03d.ts", i)
		if segment.Filename != expectedFilename {
			t.Errorf("Segment %d filename = %s, want %s", i, segment.Filename, expectedFilename)
		}
	}
}

func TestArchiveApp_Archive_StreamRepoError(t *testing.T) {
	// Setup
	streamRepo := &mockStreamRepo{
		err: errors.New("stream repo error"),
	}
	archiveRepo := &mockArchiveRepo{}

	// Execute
	app := app.NewArchiveApp(streamRepo, archiveRepo)
	result := app.Archive()

	// Assert
	expectedError := errors.New("failed to get recorder playlist: stream repo error")
	if result.Error == nil {
		t.Error("Expected error, got nil")
	} else if result.Error.Error() != expectedError.Error() {
		t.Errorf("Error = %v, want %v", result.Error, expectedError)
	}

	if result.ArchivedSegments != 0 {
		t.Errorf("ArchivedSegments = %v, want 0", result.ArchivedSegments)
	}

	// Verify no playlist was written to the archive
	if archiveRepo.playlist != nil {
		t.Error("Expected no playlist to be written to archive, got non-nil")
	}
}

func TestArchiveApp_Archive_ArchiveRepoError(t *testing.T) {
	// Setup
	now := time.Now()
	streamRepo := &mockStreamRepo{
		playlist: &playlist.Playlist{
			Version:        3,
			TargetDuration: 10,
			MediaSequence:  0,
			Segments: []playlist.Segment{
				{
					Filename:        "segment_00.ts",
					Duration:        10,
					DateTime:        now,
					ProgramDateTime: "",
				},
			},
		},
		segment: []byte("test segment"),
	}
	archiveRepo := &mockArchiveRepo{
		err: errors.New("repository error"),
	}

	// Execute
	app := app.NewArchiveApp(streamRepo, archiveRepo)
	result := app.Archive()

	// Assert
	if result.ArchivedSegments != 0 {
		t.Errorf("ArchivedSegments = %v, want 0", result.ArchivedSegments)
	}

	// Check for error in result
	if result.Error == nil {
		t.Error("Expected error in result, got nil")
	} else if !errors.Is(result.Error, archiveRepo.err) {
		t.Errorf("Expected error to be %v, got %v", archiveRepo.err, result.Error)
	}

	// Verify no playlist was written to the archive
	if archiveRepo.playlist != nil {
		t.Error("Expected no playlist to be written to archive, got non-nil")
	}
}

// Mock implementations

type mockStreamRepo struct {
	playlist *playlist.Playlist
	segment  []byte
	err      error
}

func (m *mockStreamRepo) GetPlaylist() (*playlist.Playlist, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.playlist, nil
}

func (m *mockStreamRepo) GetSegment(filename string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(bytes.NewReader(m.segment)), nil
}

type mockArchiveRepo struct {
	playlist *playlist.Playlist
	err      error
}

func (m *mockArchiveRepo) ReadPlaylist(time time.Time) (*playlist.Playlist, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.playlist, nil
}

func (m *mockArchiveRepo) WritePlaylist(time time.Time, playlist *playlist.Playlist) error {
	if m.err != nil {
		return m.err
	}
	m.playlist = playlist
	return nil
}

func (m *mockArchiveRepo) WriteSegment(time time.Time, filename string, content io.ReadCloser) error {
	if m.err != nil {
		return m.err
	}
	return nil
}
