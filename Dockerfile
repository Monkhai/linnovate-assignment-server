FROM golang:1.23.2-alpine AS builder

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Create empty credential file if it doesn't exist (will be mounted in production)
RUN if [ ! -f serviceAccountKey.json ]; then echo "{}" > serviceAccountKey.json; fi

# Build the application
RUN go build -o server .

# Final stage
FROM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/server .

# Copy environment and configuration files
COPY .env.production ./
COPY --from=builder /app/serviceAccountKey.json ./

# Expose the server port (default is 8080 as per .env.production)
EXPOSE 8080

# Run the server
CMD ["./server"] 