# ---------- Stage 1: Build ----------
    FROM golang:1.24-alpine AS builder
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
    
    # TLS certs + ffmpeg (ffprobe)
    RUN apk add --no-cache ca-certificates ffmpeg curl tar
    
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