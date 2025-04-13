#!/bin/sh

# Script to record m3u8 video stream using ffmpeg


PLAYLIST_NAME="playlist.m3u8"
SEGMENT_TEMPLATE="segment_%03d.ts"

# Check if required environment variables are set
if [ -z "$STREAM_URL" ]; then
    echo "Error: STREAM_URL environment variable is not set"
    exit 1
fi

if [ -z "$OUTPUT_DIR" ]; then
    echo "Error: OUTPUT_DIR environment variable is not set"
    exit 1
fi

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

# Start FFmpeg to capture the stream
#
# Arguments:
# -y: Overwrite output files without asking
# -i: Input URL (HLS stream)
# -c copy: Copy streams without re-encoding
# -f hls: Force HLS format output
# -hls_time 10: Duration of each segment in seconds
# -hls_list_size 360: Keep 360 segments (1 hour at 10s segments)
# -hls_flags:
#   - delete_segments: Remove old segments when they exceed list_size
#   - program_date_time: Add EXT-X-PROGRAM-DATE-TIME tags to segments
# -hls_segment_filename: Pattern for segment filenames
echo "Starting recording from: $STREAM_URL"
exec ffmpeg -y \
    -i "$STREAM_URL" \
    -c copy \
    -f hls \
    -hls_time 10 \
    -hls_list_size 360 \
    -hls_flags delete_segments+program_date_time \
    -hls_segment_filename "$OUTPUT_DIR/$SEGMENT_TEMPLATE" \
    "$OUTPUT_DIR/$PLAYLIST_NAME"
