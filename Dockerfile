# Stage 1: Build
FROM golang:1.26.2-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the application
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main cmd/api/main.go

# Stage 2: Test
FROM builder AS tester
RUN go test -v ./...

# Stage 3: Runner
FROM alpine:3.19 AS runner

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /app/main /app/main

# Copy config if needed (e.g., .env)
# COPY --from=builder /app/.env .env

EXPOSE 8080

ENTRYPOINT ["/app/main"]
