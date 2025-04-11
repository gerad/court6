# Build stage
FROM golang:1.24.2-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main .

# Final stage
FROM alpine:latest

# Install ffmpeg
RUN apk add --no-cache ffmpeg

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main .

# Switch to non-root user
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser

# Set default environment variable
ENV PORT=6001

# Expose the port (default 6001)
EXPOSE ${PORT}

# Run the binary
CMD ["./main"]
