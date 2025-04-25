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
RUN CGO_ENABLED=0 GOOS=linux go build -o videoserver .

# Final stage
FROM alpine:latest

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/videoserver .

COPY ./site /site

# Set default environment variable
ENV PORT=6001

# Expose the port
EXPOSE ${PORT}

# Run the binary
CMD ["./videoserver"]
