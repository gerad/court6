package playlist

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"
)

// Segment represents a single segment in an HLS playlist
type Segment struct {
	Filename        string
	DateTime        time.Time
	ProgramDateTime string
	Duration        float64
}

// Playlist represents a complete HLS playlist
type Playlist struct {
	Version        int
	TargetDuration int
	MediaSequence  int
	Segments       []Segment
}

// Parse reads an HLS playlist from a reader and returns a Playlist struct
func Parse(reader io.Reader) (*Playlist, error) {
	playlist := &Playlist{
		Version:        3,
		TargetDuration: 10,
		MediaSequence:  0,
		Segments:       make([]Segment, 0),
	}

	scanner := bufio.NewScanner(reader)
	dateTimeRegex := regexp.MustCompile(`#EXT-X-PROGRAM-DATE-TIME:(.+)`)
	durationRegex := regexp.MustCompile(`#EXTINF:([\d.]+),`)

	// Read all lines into memory first
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning playlist: %w", err)
	}

	fmt.Printf("Parsing playlist with %d lines\n", len(lines))
	for i, line := range lines {
		fmt.Printf("Line %d: %s\n", i, line)
	}

	// Process the lines
	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// Handle playlist headers
		switch {
		case strings.HasPrefix(line, "#EXT-X-VERSION:"):
			fmt.Sscanf(line, "#EXT-X-VERSION:%d", &playlist.Version)
			fmt.Printf("Found version: %d\n", playlist.Version)
		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			fmt.Sscanf(line, "#EXT-X-TARGETDURATION:%d", &playlist.TargetDuration)
			fmt.Printf("Found target duration: %d\n", playlist.TargetDuration)
		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &playlist.MediaSequence)
			fmt.Printf("Found media sequence: %d\n", playlist.MediaSequence)
		}

		// Look for segment information
		if strings.HasSuffix(line, ".ts") {
			fmt.Printf("Found segment: %s\n", line)
			// This is a segment filename, look for its metadata in previous lines
			var segment Segment
			segment.Filename = line

			// Look for duration in previous lines
			for j := i - 1; j >= 0; j-- {
				if durationRegex.MatchString(lines[j]) {
					fmt.Sscanf(lines[j], "#EXTINF:%f,", &segment.Duration)
					fmt.Printf("Found duration for %s: %f\n", line, segment.Duration)
					break
				}
			}

			// Look for program date time in previous lines
			for j := i - 1; j >= 0; j-- {
				if dateTimeRegex.MatchString(lines[j]) {
					matches := dateTimeRegex.FindStringSubmatch(lines[j])
					if len(matches) > 1 {
						var err error
						// Try parsing with RFC3339 first
						segment.DateTime, err = time.Parse(time.RFC3339, matches[1])
						if err != nil {
							// If that fails, try parsing with a custom format that handles +0000 timezone
							segment.DateTime, err = time.Parse("2006-01-02T15:04:05.999-0700", matches[1])
							if err != nil {
								fmt.Printf("Failed to parse date time %s: %v\n", matches[1], err)
								continue
							}
						}
						segment.ProgramDateTime = matches[1]
						fmt.Printf("Found date time for %s: %s\n", line, segment.ProgramDateTime)
					}
					break
				}
			}

			// Only add the segment if we have all the required information
			if !segment.DateTime.IsZero() && segment.Duration > 0 {
				fmt.Printf("Adding segment to playlist: %s\n", line)
				playlist.Segments = append(playlist.Segments, segment)
			} else {
				fmt.Printf("Skipping segment %s: DateTime=%v, Duration=%f\n", line, segment.DateTime, segment.Duration)
			}
		}
	}

	fmt.Printf("Parsed playlist with %d segments\n", len(playlist.Segments))
	return playlist, nil
}

// String returns the HLS playlist as a string
func (p *Playlist) String() string {
	var sb strings.Builder

	sb.WriteString("#EXTM3U\n")
	sb.WriteString(fmt.Sprintf("#EXT-X-VERSION:%d\n", p.Version))
	sb.WriteString(fmt.Sprintf("#EXT-X-TARGETDURATION:%d\n", p.TargetDuration))
	sb.WriteString(fmt.Sprintf("#EXT-X-MEDIA-SEQUENCE:%d\n", p.MediaSequence))

	for _, segment := range p.Segments {
		sb.WriteString(fmt.Sprintf("#EXTINF:%.6f,\n", segment.Duration))
		if segment.ProgramDateTime != "" {
			sb.WriteString(fmt.Sprintf("#EXT-X-PROGRAM-DATE-TIME:%s\n", segment.ProgramDateTime))
		}
		sb.WriteString(segment.Filename + "\n")
	}

	return sb.String()
}

// Concat returns a new playlist with the additional segment
func Concat(playlist *Playlist, segment Segment) *Playlist {
	newPlaylist := *playlist
	newPlaylist.Segments = append(newPlaylist.Segments, segment)
	return &newPlaylist
}
