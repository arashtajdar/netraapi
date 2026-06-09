# Stage 1: Build the Go binary
FROM golang:alpine AS builder

WORKDIR /app
COPY . .

# Generate go.sum if it's missing by scanning the source code
RUN go mod tidy
RUN go mod download

# Build a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -o sheedbox-api main.go

# Stage 2: Minimal runtime image
FROM alpine:latest

# Add certificates for TLS (required to talk to Cloudflare R2 / AWS S3)
RUN apk --no-cache add ca-certificates tzdata ffmpeg

WORKDIR /app

# Copy the binary and necessary view templates
COPY --from=builder /app/sheedbox-api .
COPY --from=builder /app/views ./views

# Expose the default port (Railway will override this with its own PORT env var)
EXPOSE 9876

CMD ["./sheedbox-api"]
