# Stage 1 - Build
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install git and other tools for go mod
RUN apk add --no-cache git

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the project
COPY . .

# Build the app binary from your main.go
RUN go build -o haerd-dating-api ./cmd

# Stage 2 — Minimal runtime image
FROM alpine:latest

# Install ffmpeg (includes ffprobe)
RUN apk add --no-cache ffmpeg

# Create a non-root user for security
RUN adduser -D appuser
USER appuser

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/haerd-dating-api .

# Copy migrations and .env file
COPY --from=builder /app/migrations ./migrations
#COPY --from=builder /app/.env .env
#COPY --from=builder /app/secrets ./secrets

# Expose the port the app listens on
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./dating-api"]