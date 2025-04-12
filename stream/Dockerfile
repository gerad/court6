FROM alpine:latest

# Install ffmpeg
RUN apk add --no-cache ffmpeg

WORKDIR /app

# Copy the stream script
COPY stream.sh .

# Make the script executable
RUN chmod +x stream.sh

# Switch to non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Set default environment variable
ENV PORT=6001

# Run the script
CMD ["/bin/sh", "./stream.sh"]
