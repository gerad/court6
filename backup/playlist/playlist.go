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

	var currentDateTime time.Time
	var currentProgramDateTime string
	var currentDuration float64

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.HasPrefix(line, "#EXT-X-VERSION:"):
			fmt.Sscanf(line, "#EXT-X-VERSION:%d", &playlist.Version)
		case strings.HasPrefix(line, "#EXT-X-TARGETDURATION:"):
			fmt.Sscanf(line, "#EXT-X-TARGETDURATION:%d", &playlist.TargetDuration)
		case strings.HasPrefix(line, "#EXT-X-MEDIA-SEQUENCE:"):
			fmt.Sscanf(line, "#EXT-X-MEDIA-SEQUENCE:%d", &playlist.MediaSequence)
		case dateTimeRegex.MatchString(line):
			matches := dateTimeRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentDateTime, _ = time.Parse(time.RFC3339, matches[1])
				currentProgramDateTime = matches[1]
			}
		case durationRegex.MatchString(line):
			fmt.Sscanf(line, "#EXTINF:%f,", &currentDuration)
		case strings.HasSuffix(line, ".ts"):
			playlist.Segments = append(playlist.Segments, Segment{
				Filename:        line,
				DateTime:        currentDateTime,
				ProgramDateTime: currentProgramDateTime,
				Duration:        currentDuration,
			})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error scanning playlist: %w", err)
	}

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

// FindNewSegments returns segments that exist in source but not in target
func FindNewSegments(source, target *Playlist) []Segment {
	existing := make(map[string]bool)
	for _, segment := range target.Segments {
		existing[segment.Filename] = true
	}

	var newSegments []Segment
	for _, segment := range source.Segments {
		if !existing[segment.Filename] {
			newSegments = append(newSegments, segment)
		}
	}

	return newSegments
}

// Concat returns a new playlist with the additional segment
func Concat(playlist *Playlist, segment Segment) *Playlist {
	newPlaylist := *playlist
	newPlaylist.Segments = append(newPlaylist.Segments, segment)
	return &newPlaylist
}
