package gateway

import (
	"io"
	"net/http"

	"backup/playlist"
)

// PlaylistGateway defines the interface for fetching playlists and segments
type PlaylistGateway interface {
	// GetPlaylist fetches the playlist from the recorder server
	GetPlaylist() (*playlist.Playlist, error)
	// GetSegment fetches a segment from the recorder server
	GetSegment(filename string) (io.ReadCloser, error)
}

// HTTPPlaylistGateway implements PlaylistGateway using HTTP
type HTTPPlaylistGateway struct {
	recorderURL string
	client      *http.Client
}

// NewHTTPPlaylistGateway creates a new HTTPPlaylistGateway
func NewHTTPPlaylistGateway(recorderURL string) *HTTPPlaylistGateway {
	return &HTTPPlaylistGateway{
		recorderURL: recorderURL + "/videos",
		client:      &http.Client{},
	}
}

// GetPlaylist fetches the playlist from the recorder server
func (g *HTTPPlaylistGateway) GetPlaylist() (*playlist.Playlist, error) {
	playlistURL := g.recorderURL + "/playlist.m3u8"
	resp, err := g.client.Get(playlistURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return playlist.Parse(resp.Body)
}

// GetSegment fetches a segment from the recorder server
func (g *HTTPPlaylistGateway) GetSegment(filename string) (io.ReadCloser, error) {
	segmentURL := g.recorderURL + "/" + filename
	resp, err := g.client.Get(segmentURL)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}
