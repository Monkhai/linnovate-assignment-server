FROM --platform=$BUILDPLATFORM golang:1.23.2-alpine AS builder

ARG TARGETPLATFORM
ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Create empty credential file if it doesn't exist (will be mounted in production)
RUN if [ ! -f serviceAccountKey.json ]; then echo "{}" > serviceAccountKey.json; fi

# Create empty .env.production file if it doesn't exist (will be overridden in production)
RUN if [ ! -f .env.production ]; then echo "APP_ENV=production\nSERVER_PORT=8080" > .env.production; fi

# Build the application for the target architecture from main.go
RUN GOOS=$TARGETOS GOARCH=$TARGETARCH go build -o app main.go
RUN ls -la
RUN chmod +x app

# Final stage
FROM --platform=$TARGETPLATFORM alpine:latest

WORKDIR /app

# Install CA certificates for HTTPS connections
RUN apk --no-cache add ca-certificates

# Copy the binary from the builder stage
COPY --from=builder /app/app /app/app

# Copy environment and configuration files
COPY --from=builder /app/.env.production ./
COPY --from=builder /app/serviceAccountKey.json ./

# Verify the app binary exists and is executable
RUN ls -la /app && chmod +x /app/app

# Expose the server port (default is 8080 as per .env.production)
EXPOSE 8080

# Run the app binary
CMD ["/app/app"] 