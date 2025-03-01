# Catalog API

A RESTful API for a product catalog and review system, built with Go.

## Environment Setup

This application uses environment variables for configuration. You can set these variables in a `.env` file or directly in your environment.

### Development Setup

For local development:

1. Copy `.env.development` to `.env`:

```bash
cp .env.development .env
```

2. Update the variables in `.env` to match your local environment:

```
# Server configuration
APP_ENV=development
SERVER_PORT=8080

# Database configuration
DB_HOST=localhost  # Replace with your Docker host if not using localhost
DB_PORT=5432       # Replace with the exposed port of your PostgreSQL container
DB_USER=postgres   # Replace with your database username
DB_PASSWORD=postgres  # Replace with your database password
DB_NAME=postgres   # Replace with your database name
DB_SSLMODE=disable
```

## Running the Application

```bash
go run main.go
```

The server will start on the port specified in your environment variables (default: 8080).

## Docker PostgreSQL Setup

If you're using a dockerized PostgreSQL database, ensure it's running before starting the application.

Example Docker command to run PostgreSQL:

```bash
docker run --name postgres -e POSTGRES_PASSWORD=postgres -p 5432:5432 -d postgres
```

Or if using docker-compose:

```bash
docker-compose up -d
```

## Testing

```bash
go test ./...
```
