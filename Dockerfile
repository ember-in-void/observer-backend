# Build stage
FROM golang:1.24-alpine AS builder

# Install required system dependencies (if any)
RUN apk add --no-cache git

WORKDIR /app

# specific cache for go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
# -ldflags="-w -s" reduces binary size by stripping debug information
# CGO_ENABLED=0 ensures a statically linked binary for scratch/distroless compatibility
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/server ./cmd/api/main.go

# Final stage
# strict minimal image for security and size
FROM gcr.io/distroless/static-debian12:nonroot

WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Expose the application port (defaults to 8080 in config)
EXPOSE 8080

# Run as non-root user (provided by distroless)
USER nonroot:nonroot

ENTRYPOINT ["/app/server"]
