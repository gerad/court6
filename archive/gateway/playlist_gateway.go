package gateway

import (
	"io"
	"os"
	"path/filepath"

	"archive/playlist"
)

// PlaylistGateway defines the interface for fetching playlists and segments
type PlaylistGateway interface {
	// GetPlaylist fetches the playlist from the recorder server
	GetPlaylist() (*playlist.Playlist, error)
	// GetSegment fetches a segment from the recorder server
	GetSegment(filename string) (io.ReadCloser, error)
}

// FileSystemPlaylistGateway implements PlaylistGateway using the filesystem
type FileSystemPlaylistGateway struct {
	basePath string
}

// NewFileSystemPlaylistGateway creates a new FileSystemPlaylistGateway
func NewFileSystemPlaylistGateway(basePath string) *FileSystemPlaylistGateway {
	return &FileSystemPlaylistGateway{
		basePath: basePath,
	}
}

// GetPlaylist reads the playlist from the filesystem
func (g *FileSystemPlaylistGateway) GetPlaylist() (*playlist.Playlist, error) {
	playlistPath := filepath.Join(g.basePath, "playlist.m3u8")
	file, err := os.Open(playlistPath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return playlist.Parse(file)
}

// GetSegment reads a segment from the filesystem
func (g *FileSystemPlaylistGateway) GetSegment(filename string) (io.ReadCloser, error) {
	segmentPath := filepath.Join(g.basePath, filename)
	file, err := os.Open(segmentPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
