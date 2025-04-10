# Use Alpine Linux as our base image
FROM alpine:latest

# Install ffmpeg, bash and its dependencies
RUN apk add --no-cache ffmpeg bash

RUN mkdir -p /app
COPY record.sh /app/

# Switch to non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Set working directory
WORKDIR /app

# Default command to run the script
CMD ["bash", "/app/record.sh"]
