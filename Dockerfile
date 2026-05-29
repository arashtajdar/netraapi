# ==========================================
# STAGE 1: BUILD
# ==========================================
FROM golang:1.22-alpine AS builder

# Install git for fetching dependencies
RUN apk update && apk add --no-cache git

WORKDIR /app

# Copy dependency graphs
COPY go.mod ./

# Fetch dependencies securely
RUN go mod tidy
RUN go mod download

# Copy the entire backend source code
COPY . .

# Compile the Go application as a static standalone binary stripping debug symbols for size
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -a -installsuffix cgo -o netra-api .

# ==========================================
# STAGE 2: PRODUCTION (Ultra-lightweight)
# ==========================================
FROM scratch

WORKDIR /app/

# Copy the compiled binary from the builder stage
COPY --from=builder /app/netra-api .

# Ensure the .env file is copied (Alternatively inject via Docker runtime flags)
COPY --from=builder /app/.env .

# Required for SSL/TLS connections if the Go backend hits external HTTPS APIs
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

EXPOSE 9876

CMD ["./netra-api"]
