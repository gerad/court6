package app_test

import (
	"archive/app"
	"archive/playlist"
	"bytes"
	"errors"
	"io"
	"testing"
	"time"
)

type mockGateway struct {
	playlist *playlist.Playlist
	segment  []byte
	err      error
}

func (m *mockGateway) GetPlaylist() (*playlist.Playlist, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.playlist, nil
}

func (m *mockGateway) GetSegment(filename string) (io.ReadCloser, error) {
	if m.err != nil {
		return nil, m.err
	}
	return io.NopCloser(bytes.NewReader(m.segment)), nil
}

type mockRepository struct {
	playlist *playlist.Playlist
	err      error
}

func (m *mockRepository) ReadPlaylist(time time.Time) (*playlist.Playlist, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.playlist, nil
}

func (m *mockRepository) WritePlaylist(time time.Time, playlist *playlist.Playlist) error {
	if m.err != nil {
		return m.err
	}
	m.playlist = playlist
	return nil
}

func (m *mockRepository) WriteSegment(time time.Time, filename string, content io.ReadCloser) error {
	if m.err != nil {
		return m.err
	}
	return nil
}

func TestArchiveApp_Archive(t *testing.T) {
	tests := []struct {
		name           string
		gateway        *mockGateway
		repository     *mockRepository
		expectedResult app.ArchiveResult
	}{
		{
			name: "successful archive",
			gateway: &mockGateway{
				playlist: &playlist.Playlist{
					Version:        3,
					TargetDuration: 10,
					MediaSequence:  0,
					Segments: []playlist.Segment{
						{
							Filename:        "segment_00.ts",
							Duration:        10,
							DateTime:        time.Now(),
							ProgramDateTime: "",
						},
					},
				},
				segment: []byte("test segment"),
			},
			repository: &mockRepository{},
			expectedResult: app.ArchiveResult{
				ArchivedSegments: 1,
			},
		},
		{
			name: "gateway error",
			gateway: &mockGateway{
				err: errors.New("gateway error"),
			},
			repository: &mockRepository{},
			expectedResult: app.ArchiveResult{
				Error: errors.New("failed to get recorder playlist: gateway error"),
			},
		},
		{
			name: "repository error",
			gateway: &mockGateway{
				playlist: &playlist.Playlist{
					Version:        3,
					TargetDuration: 10,
					MediaSequence:  0,
					Segments: []playlist.Segment{
						{
							Filename:        "segment_00.ts",
							Duration:        10,
							DateTime:        time.Now(),
							ProgramDateTime: "",
						},
					},
				},
				segment: []byte("test segment"),
			},
			repository: &mockRepository{
				err: errors.New("repository error"),
			},
			expectedResult: app.ArchiveResult{
				ArchivedSegments: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := app.NewArchiveApp(tt.gateway, tt.repository)
			result := app.Archive()

			if result.ArchivedSegments != tt.expectedResult.ArchivedSegments {
				t.Errorf("BackedUpSegments = %v, want %v", result.ArchivedSegments, tt.expectedResult.ArchivedSegments)
			}

			if result.Error != nil && tt.expectedResult.Error != nil {
				if result.Error.Error() != tt.expectedResult.Error.Error() {
					t.Errorf("Error = %v, want %v", result.Error, tt.expectedResult.Error)
				}
			} else if result.Error != tt.expectedResult.Error {
				t.Errorf("Error = %v, want %v", result.Error, tt.expectedResult.Error)
			}
		})
	}
}
