#!/bin/bash

# Get the absolute path for the current user
USER_HOME=$(eval echo ~$USER)
SA_KEY_PATH="$USER_HOME/serviceAccountKey.json"
ENV_PATH="$USER_HOME/.env.production"

# Check if files exist
if [ ! -f "$SA_KEY_PATH" ]; then
  echo "Error: $SA_KEY_PATH does not exist!"
  echo "Please upload the serviceAccountKey.json file first."
  exit 1
fi

if [ ! -f "$ENV_PATH" ]; then
  echo "Error: $ENV_PATH does not exist!"
  echo "Please upload the .env.production file first."
  exit 1
fi

# Pull the latest image
echo "Pulling the latest image..."
docker pull yohaiwiener/catalogapi:latest

# Check if the container already exists and remove it
if [ "$(docker ps -a -q -f name=catalog-api)" ]; then
  echo "Container catalog-api already exists. Stopping and removing it..."
  docker stop catalog-api
  docker rm catalog-api
fi

# Run the container with configuration files mounted
echo "Starting new container..."
docker run -d -p 8080:8080 \
  -v "$SA_KEY_PATH:/app/serviceAccountKey.json" \
  -v "$ENV_PATH:/app/.env.production" \
  --name catalog-api \
  --restart unless-stopped \
  yohaiwiener/catalogapi:latest

# Check if it's running
echo "Checking if container is running..."
docker ps | grep catalog-api
docker logs catalog-api

# Get the public IP and display it
PUBLIC_IP=$(curl -s http://169.254.169.254/latest/meta-data/public-ipv4)
if [ -n "$PUBLIC_IP" ]; then
  echo "Done! Your container should be running at http://$PUBLIC_IP:8080"
else
  echo "Done! Your container should be running. Couldn't determine public IP."
fi 