# Catalog API

## Docker Deployment

This application uses Docker for deployment. The image is automatically built and pushed to Docker Hub when changes are pushed to the main branch.

### Credentials Management

The application requires a Firebase service account key (`serviceAccountKey.json`) to work properly. For security reasons, this file is not included in the repository. When the Docker image is built, a placeholder empty JSON file is created.

When deploying the application, you need to provide the actual service account key in one of the following ways:

#### Option 1: Mount the file when running the container

```bash
docker run -d -p 8080:8080 \
  -v /path/to/your/serviceAccountKey.json:/app/serviceAccountKey.json \
  --name catalog-api \
  yourusername/catalogapi:latest
```

#### Option 2: Copy the file into a running container

```bash
# Start the container
docker run -d -p 8080:8080 --name catalog-api yourusername/catalogapi:latest

# Copy the file into the container
docker cp /path/to/your/serviceAccountKey.json catalog-api:/app/serviceAccountKey.json

# Restart the container
docker restart catalog-api
```

### Environment Variables

The application uses environment variables from `.env.production`. If you need to override any of these variables, you can pass them when running the container:

```bash
docker run -d -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=your-user \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=your-db-name \
  --name catalog-api \
  yourusername/catalogapi:latest
```

## Local Development

For local development:

1. Clone the repository
2. Create a `.env.development` file with your local configuration
3. Place your `serviceAccountKey.json` file in the root directory
4. Run `go run main.go`
