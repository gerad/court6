#!/bin/bash

# Script to record m3u8 video stream using ffmpeg

# Check if RECORDING_URL is set
if [ -z "$RECORDING_URL" ]; then
    echo "Error: RECORDING_URL environment variable is not set"
    exit 1
fi

# Set default output directory and playlist name
OUTPUT_DIR="/recorder/recordings"
PLAYLIST_NAME="playlist.m3u8"

# Create output directory if it doesn't exist
mkdir -p "$OUTPUT_DIR"

echo "Starting recording from: $RECORDING_URL"
echo "Output will be saved to: $OUTPUT_DIR/$PLAYLIST_NAME"

# Record the m3u8 stream
# -y: Overwrite output file if it exists
# -i: Input URL
# -c copy: Copy streams without re-encoding
# -f hls: Use HLS format
# -hls_time 10: Create segments of 10 seconds
# -hls_list_size 0: Keep all segments
# -hls_segment_filename: Specify the pattern for segment files
ffmpeg -y \
    -i "$RECORDING_URL" \
    -c copy \
    -f hls \
    -hls_time 10 \
    -hls_list_size 0 \
    -hls_segment_filename "$OUTPUT_DIR/segment_%03d.ts" \
    "$OUTPUT_DIR/$PLAYLIST_NAME"

# Check if ffmpeg command was successful
if [ $? -eq 0 ]; then
    echo "Recording completed successfully."
    echo "Playlist saved to: $OUTPUT_DIR/$PLAYLIST_NAME"
    echo "Segments saved in: $OUTPUT_DIR/"
else
    echo "Error: Recording failed"
    exit 1
fi 
