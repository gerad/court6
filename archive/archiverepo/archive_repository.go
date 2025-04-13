package archiverepo

import (
	"archive/playlist"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

// ArchiveRepository stores video playlists and segments on the filesystem
type ArchiveRepository struct {
	basePath string
}

// NewArchiveRepository creates a new ArchiveRepository
func New(basePath string) *ArchiveRepository {
	return &ArchiveRepository{
		basePath: basePath,
	}
}

// getBackupPath returns the path for a specific time
func (r *ArchiveRepository) getBackupPath(segmentTime time.Time) (string, error) {
	path := filepath.Join(r.basePath,
		fmt.Sprintf("%d", segmentTime.Year()),
		fmt.Sprintf("%02d", segmentTime.Month()),
		fmt.Sprintf("%02d", segmentTime.Day()),
		fmt.Sprintf("%02d", segmentTime.Hour()))
	return path, nil
}

// ReadPlaylist reads the archive playlist from the filesystem for a specific time
func (r *ArchiveRepository) ReadPlaylist(segmentTime time.Time) (*playlist.Playlist, error) {
	// Ensure backup directory exists before reading
	if err := r.ensureBackupDirectory(segmentTime); err != nil {
		return nil, err
	}

	path, err := r.getBackupPath(segmentTime)
	if err != nil {
		return nil, err
	}

	playlistPath := filepath.Join(path, "playlist.m3u8")
	file, err := os.Open(playlistPath)
	if err != nil {
		if os.IsNotExist(err) {
			// If the playlist doesn't exist, return nil
			return nil, nil
		}
		return nil, err
	}
	defer file.Close()

	return playlist.Parse(file)
}

// WritePlaylist writes the playlist to the filesystem for a specific time
func (r *ArchiveRepository) WritePlaylist(segmentTime time.Time, playlist *playlist.Playlist) error {
	// Ensure backup directory exists before writing
	if err := r.ensureBackupDirectory(segmentTime); err != nil {
		return err
	}

	path, err := r.getBackupPath(segmentTime)
	if err != nil {
		return err
	}

	playlistPath := filepath.Join(path, "playlist.m3u8")
	file, err := os.Create(playlistPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(playlist.String())
	return err
}

// WriteSegment writes a segment to the filesystem for a specific time
func (r *ArchiveRepository) WriteSegment(segmentTime time.Time, filename string, content io.ReadCloser) error {
	// Ensure backup directory exists before writing
	if err := r.ensureBackupDirectory(segmentTime); err != nil {
		return err
	}

	path, err := r.getBackupPath(segmentTime)
	if err != nil {
		return err
	}

	segmentPath := filepath.Join(path, filename)
	file, err := os.Create(segmentPath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.Copy(file, content)
	return err
}

// ensureBackupDirectory creates the backup directory if it doesn't exist
// This is now a private method used internally by the repository
func (r *ArchiveRepository) ensureBackupDirectory(segmentTime time.Time) error {
	path, err := r.getBackupPath(segmentTime)
	if err != nil {
		return err
	}
	fmt.Printf("Ensuring backup directory exists: %s\n", path)
	err = os.MkdirAll(path, 0755)
	if err != nil {
		fmt.Printf("Error creating directory %s: %v\n", path, err)
		return err
	}
	fmt.Printf("Successfully created/verified directory: %s\n", path)
	return nil
}
