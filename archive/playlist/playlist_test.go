package playlist

import (
	"strings"
	"testing"
	"time"
)

func TestParse(t *testing.T) {
	input := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXT-X-PROGRAM-DATE-TIME:2024-04-10T23:58:00Z
#EXTINF:10.0,
segment_1.ts
#EXT-X-PROGRAM-DATE-TIME:2024-04-10T23:58:10Z
#EXTINF:10.0,
segment_2.ts`

	playlist, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if playlist.Version != 3 {
		t.Errorf("Expected version 3, got %d", playlist.Version)
	}

	if playlist.TargetDuration != 10 {
		t.Errorf("Expected target duration 10, got %d", playlist.TargetDuration)
	}

	if len(playlist.Segments) != 2 {
		t.Fatalf("Expected 2 segments, got %d", len(playlist.Segments))
	}

	expectedTimes := []string{
		"2024-04-10T23:58:00Z",
		"2024-04-10T23:58:10Z",
	}

	for i, segment := range playlist.Segments {
		if segment.ProgramDateTime != expectedTimes[i] {
			t.Errorf("Segment %d: Expected time %s, got %s", i, expectedTimes[i], segment.ProgramDateTime)
		}
		if segment.Duration != 10.0 {
			t.Errorf("Segment %d: Expected duration 10.0, got %f", i, segment.Duration)
		}
	}
}

func TestString(t *testing.T) {
	playlist := &Playlist{
		Version:        3,
		TargetDuration: 10,
		MediaSequence:  0,
		Segments: []Segment{
			{
				Duration:        10.0,
				Filename:        "segment_1.ts",
				ProgramDateTime: "2024-04-10T23:58:00Z",
				DateTime:        time.Date(2024, 4, 10, 23, 58, 0, 0, time.UTC),
			},
			{
				Duration:        10.0,
				Filename:        "segment_2.ts",
				ProgramDateTime: "2024-04-10T23:58:10Z",
				DateTime:        time.Date(2024, 4, 10, 23, 58, 10, 0, time.UTC),
			},
		},
	}

	expected := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:10
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:10.000000,
#EXT-X-PROGRAM-DATE-TIME:2024-04-10T23:58:00Z
segment_1.ts
#EXTINF:10.000000,
#EXT-X-PROGRAM-DATE-TIME:2024-04-10T23:58:10Z
segment_2.ts
`

	if playlist.String() != expected {
		t.Errorf("Expected:\n%s\nGot:\n%s", expected, playlist.String())
	}
}

func TestConcat(t *testing.T) {
	original := &Playlist{
		Segments: []Segment{
			{Filename: "segment_1.ts"},
		},
	}

	newSegment := Segment{Filename: "segment_2.ts"}
	updated := Concat(original, newSegment)

	if len(updated.Segments) != 2 {
		t.Fatalf("Expected 2 segments, got %d", len(updated.Segments))
	}

	if updated.Segments[1].Filename != "segment_2.ts" {
		t.Errorf("Expected segment_2.ts, got %s", updated.Segments[1].Filename)
	}

	// Verify original wasn't modified
	if len(original.Segments) != 1 {
		t.Errorf("Original playlist was modified")
	}
}

func TestParseRealPlaylist(t *testing.T) {
	// Use the actual content from the playlist.m3u8 file
	input := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:61
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:60.802000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:27:48.996+0000
segment_000.ts
#EXTINF:59.202000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:28:49.798+0000
segment_001.ts
#EXTINF:60.801000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:29:48.999+0000
segment_002.ts
#EXTINF:59.202000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:30:49.801+0000
segment_003.ts
#EXTINF:60.801000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:31:49.003+0000
segment_004.ts
#EXTINF:59.202000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:32:49.804+0000
segment_005.ts
#EXTINF:60.802000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:33:49.006+0000
segment_006.ts
#EXTINF:59.201000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:34:49.808+0000
segment_007.ts
#EXTINF:60.802000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:35:49.009+0000
segment_008.ts
#EXTINF:59.201000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T00:36:49.811+0000
segment_009.ts`

	playlist, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	// Test roundtrip - parse and convert back to string
	output := playlist.String()

	// Normalize both strings by trimming whitespace and normalizing line endings
	normalizedInput := strings.TrimSpace(strings.ReplaceAll(input, "\r\n", "\n"))
	normalizedOutput := strings.TrimSpace(strings.ReplaceAll(output, "\r\n", "\n"))

	if normalizedInput != normalizedOutput {
		t.Errorf("Roundtrip failed.\nExpected:\n%s\n\nGot:\n%s", normalizedInput, normalizedOutput)
	}
}

func TestParseWithTimezoneOffset(t *testing.T) {
	input := `#EXTM3U
#EXT-X-VERSION:3
#EXT-X-TARGETDURATION:61
#EXT-X-MEDIA-SEQUENCE:0
#EXTINF:60.800000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T04:53:28.029+0000
segment_000.ts
#EXTINF:59.201000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T04:54:28.829+0000
segment_001.ts
#EXTINF:60.801000,
#EXT-X-PROGRAM-DATE-TIME:2025-04-11T04:55:28.030+0000
segment_002.ts`

	playlist, err := Parse(strings.NewReader(input))
	if err != nil {
		t.Fatalf("Parse failed: %v", err)
	}

	if len(playlist.Segments) != 3 {
		t.Fatalf("Expected 3 segments, got %d", len(playlist.Segments))
	}

	expectedTimes := []string{
		"2025-04-11T04:53:28.029+0000",
		"2025-04-11T04:54:28.829+0000",
		"2025-04-11T04:55:28.030+0000",
	}

	for i, segment := range playlist.Segments {
		if segment.ProgramDateTime != expectedTimes[i] {
			t.Errorf("Segment %d: Expected time %s, got %s", i, expectedTimes[i], segment.ProgramDateTime)
		}
		if segment.DateTime.IsZero() {
			t.Errorf("Segment %d: DateTime is zero", i)
		}
		if segment.Duration != 60.8 && segment.Duration != 59.201 && segment.Duration != 60.801 {
			t.Errorf("Segment %d: Unexpected duration %f", i, segment.Duration)
		}
	}
}
