# Stage 1: Build the Go binary
FROM golang:1.26.3-alpine AS builder

WORKDIR /app

# Copy dependency manifests
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application as a statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api

# Stage 2: Create a minimal production image
FROM alpine:3.20

# Install CA certificates (critical for outgoing HTTPS requests like Firebase Auth)
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy the compiled binary from the builder stage
COPY --from=builder /app/api .

# Expose port (Cloud Run will inject the PORT environment variable)
EXPOSE 8080

# Run the binary
CMD ["./api"]
