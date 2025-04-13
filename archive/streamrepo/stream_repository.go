package streamrepo

import (
	"io"
	"os"
	"path/filepath"

	"archive/playlist"
)

// StreamRepository reads a video stream from the file system
type StreamRepository struct {
	basePath string
}

// New creates a new StreamRepository
func New(basePath string) *StreamRepository {
	return &StreamRepository{
		basePath: basePath,
	}
}

// GetPlaylist reads the playlist from the filesystem
func (g *StreamRepository) GetPlaylist() (*playlist.Playlist, error) {
	playlistPath := filepath.Join(g.basePath, "playlist.m3u8")
	file, err := os.Open(playlistPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return playlist.Parse(file)
}

// GetSegment reads a segment from the filesystem
func (g *StreamRepository) GetSegment(filename string) (io.ReadCloser, error) {
	segmentPath := filepath.Join(g.basePath, filename)
	file, err := os.Open(segmentPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
