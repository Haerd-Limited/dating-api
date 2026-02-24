# ---------- Stage 1: Build ----------
    FROM golang:1.25-alpine AS builder
    WORKDIR /app
    
    # Tools needed during build
    RUN apk add --no-cache git
    
    # Go deps
    COPY go.mod go.sum ./
    RUN go mod download
    
    # Source
    COPY . .
    
    # Build API
    RUN go build -o /out/haerd-dating-api ./cmd
    
    # Install goose and stage it in /out
    RUN go install github.com/pressly/goose/v3/cmd/goose@latest
    RUN install -Dm755 /go/bin/goose /out/goose
    
    # Copy migrations to /out for convenience
    RUN mkdir -p /out/migrations && cp -r ./migrations/* /out/migrations/
    
    # ---------- Stage 2: Runtime ----------
    FROM alpine:latest
    
    # TLS certs + ffmpeg (ffprobe) + yt-dlp dependencies (Python3 needed for yt-dlp script)
    RUN apk add --no-cache ca-certificates ffmpeg curl tar python3
    
    # Install yt-dlp (Python script that requires Python3)
    RUN curl -L https://github.com/yt-dlp/yt-dlp/releases/latest/download/yt-dlp -o /usr/local/bin/yt-dlp && \
        chmod a+rx /usr/local/bin/yt-dlp
    
    WORKDIR /app
    
    # Copy artifacts
    COPY --from=builder /out/haerd-dating-api .
    COPY --from=builder /out/goose /usr/local/bin/goose
    COPY --from=builder /out/migrations ./migrations
    
    # Non-root user
    RUN adduser -D appuser
    USER appuser
    
    EXPOSE 8080
    ENTRYPOINT ["./haerd-dating-api"]