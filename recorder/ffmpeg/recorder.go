package ffmpeg

import (
	"fmt"
	"log"
	"os"
	"os/exec"
)

const (
	// DefaultHLS segment duration in seconds
	DefaultHLSTime = 6
	// Default number of segments to keep in the playlist
	DefaultHLSListSize = 10
	// Default playlist filename
	DefaultPlaylistName = "playlist.m3u8"
	// Default segment filename pattern
	DefaultSegmentPattern = "segment_%03d.ts"
)

// Recorder represents an FFmpeg HLS stream recorder
type Recorder struct {
	recordingURL string
	outputDir    string
	cmd          *exec.Cmd
}

// New creates a new FFmpeg recorder
func New(recordingURL, outputDir string) *Recorder {
	return &Recorder{
		recordingURL: recordingURL,
		outputDir:    outputDir,
	}
}

// Start begins recording the stream
func (r *Recorder) Start() error {
	// Create output directory if it doesn't exist
	if err := os.MkdirAll(r.outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Build FFmpeg command
	r.cmd = exec.Command("ffmpeg",
		"-y",
		"-i", r.recordingURL,
		"-c", "copy",
		"-f", "hls",
		"-hls_time", fmt.Sprintf("%d", DefaultHLSTime),
		"-hls_list_size", fmt.Sprintf("%d", DefaultHLSListSize),
		"-hls_flags", "delete_segments+program_date_time",
		"-hls_segment_filename", fmt.Sprintf("%s/%s", r.outputDir, DefaultSegmentPattern),
		fmt.Sprintf("%s/%s", r.outputDir, DefaultPlaylistName),
	)

	// Connect stdout and stderr to the process
	r.cmd.Stdout = os.Stdout
	r.cmd.Stderr = os.Stderr

	// Start the process
	if err := r.cmd.Start(); err != nil {
		return fmt.Errorf("failed to start FFmpeg: %w", err)
	}

	log.Printf("Started FFmpeg recording from %s to %s", r.recordingURL, r.outputDir)
	return nil
}

// Stop gracefully stops the recording
func (r *Recorder) Stop() error {
	if r.cmd == nil || r.cmd.Process == nil {
		return nil
	}

	log.Println("Stopping FFmpeg recording...")

	// Try SIGTERM first
	if err := r.cmd.Process.Signal(os.Interrupt); err != nil {
		log.Printf("Error sending SIGTERM to FFmpeg: %v", err)
		// If SIGTERM fails, try SIGKILL
		if err := r.cmd.Process.Kill(); err != nil {
			return fmt.Errorf("failed to kill FFmpeg process: %w", err)
		}
	}

	// Wait for the process to finish
	if err := r.cmd.Wait(); err != nil {
		return fmt.Errorf("FFmpeg process exited with error: %w", err)
	}

	log.Println("FFmpeg recording stopped")
	return nil
}
