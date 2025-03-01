# Catalog API

## Docker Deployment

This application uses Docker for deployment. The image is automatically built and pushed to Docker Hub when changes are pushed to the main branch.

### Configuration Files

The application requires two important configuration files:

1. **Firebase Service Account Key** (`serviceAccountKey.json`)
2. **Environment Configuration** (`.env.production`)

For security reasons, these files are not included in the repository. When the Docker image is built, placeholder files are created with minimal default values.

When deploying the application, you need to provide the actual files:

#### Option 1: Mount both files when running the container

```bash
docker run -d -p 8080:8080 \
  -v /path/to/your/serviceAccountKey.json:/app/serviceAccountKey.json \
  -v /path/to/your/.env.production:/app/.env.production \
  --name catalog-api \
  yourusername/catalogapi:latest
```

#### Option 2: Copy the files into a running container

```bash
# Start the container
docker run -d -p 8080:8080 --name catalog-api yourusername/catalogapi:latest

# Copy the files into the container
docker cp /path/to/your/serviceAccountKey.json catalog-api:/app/serviceAccountKey.json
docker cp /path/to/your/.env.production catalog-api:/app/.env.production

# Restart the container
docker restart catalog-api
```

#### Option 3: Use environment variables instead of .env.production

```bash
docker run -d -p 8080:8080 \
  -v /path/to/your/serviceAccountKey.json:/app/serviceAccountKey.json \
  -e DB_HOST=your-db-host \
  -e DB_PORT=5432 \
  -e DB_USER=your-user \
  -e DB_PASSWORD=your-password \
  -e DB_NAME=your-db-name \
  -e SERVER_PORT=8080 \
  -e APP_ENV=production \
  -e AWS_REGION=your-region \
  -e AWS_DB_SECRET_NAME=your-secret-name \
  --name catalog-api \
  yourusername/catalogapi:latest
```

## Local Development

For local development:

1. Clone the repository
2. Create a `.env.development` file with your local configuration
3. Place your `serviceAccountKey.json` file in the root directory
4. Run `go run main.go`
