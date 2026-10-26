# Stage 1: Build
FROM golang:1.27.3-alpine AS builder

WORKDIR /app

# Download dependencies first for better build caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build a static Linux binary
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath \
    -ldflags="-s -w" \
    -o /app/bin/server \
    ./cmd/server

# Stage 2: Runtime
FROM alpine:3.22

WORKDIR /app

# Install CA certificates for HTTPS requests
RUN apk add --no-cache ca-certificates

COPY --from=builder /app/bin/server ./server

EXPOSE 8080

CMD ["./server"]