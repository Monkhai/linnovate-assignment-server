# EC2 Deployment Guide

This is a simplified guide for deploying the Docker image to AWS EC2.

## Step 1: Launch an EC2 Instance

1. Launch an EC2 instance (Amazon Linux 2 or Ubuntu)
2. Configure security group to allow:
   - SSH (port 22) from your IP
   - HTTP (port 80) from anywhere

## Step 2: Install Docker on EC2

For Amazon Linux 2:

```bash
sudo yum update -y
sudo amazon-linux-extras install docker -y
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -a -G docker ec2-user
# Log out and log back in
exit
```

After logging back in:

```bash
# Verify Docker is working
docker --version
```

## Step 3: Upload Your Configuration Files

You'll need to upload both your Firebase service account key and your environment configuration to the EC2 instance.

From your local machine:

```bash
# Upload the service account key
scp -i your-key.pem /path/to/local/serviceAccountKey.json ec2-user@your-ec2-public-ip:~/serviceAccountKey.json

# Upload the environment configuration
scp -i your-key.pem /path/to/local/.env.production ec2-user@your-ec2-public-ip:~/.env.production
```

## Step 4: Pull and Run the Docker Image

```bash
# Pull the image from Docker Hub
docker pull yourusername/catalogapi:latest

# Run the container with both configuration files mounted
docker run -d -p 8080:8080 \
  -v ~/serviceAccountKey.json:/app/serviceAccountKey.json \
  -v ~/.env.production:/app/.env.production \
  --name catalog-api \
  --restart unless-stopped \
  yourusername/catalogapi:latest

# Check if it's running
docker ps

# Check logs
docker logs catalog-api
```

## Step 5: Setup Nginx (if needed)

Install and configure Nginx as a reverse proxy:

```bash
# Install Nginx
# For Amazon Linux 2
sudo amazon-linux-extras install nginx1 -y

# For Ubuntu
sudo apt install -y nginx

# Start and enable Nginx
sudo systemctl start nginx
sudo systemctl enable nginx

# Create a config file
sudo nano /etc/nginx/conf.d/api.conf
```

Nginx configuration:

```
server {
    listen 80;
    server_name _;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;

        # CORS headers (if needed)
        add_header 'Access-Control-Allow-Origin' '*' always;
        add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS, PUT, DELETE' always;
        add_header 'Access-Control-Allow-Headers' 'Origin, X-Requested-With, Content-Type, Accept, Authorization' always;

        # Handle preflight requests
        if ($request_method = 'OPTIONS') {
            add_header 'Access-Control-Allow-Origin' '*';
            add_header 'Access-Control-Allow-Methods' 'GET, POST, OPTIONS, PUT, DELETE';
            add_header 'Access-Control-Allow-Headers' 'Origin, X-Requested-With, Content-Type, Accept, Authorization';
            add_header 'Access-Control-Max-Age' 1728000;
            add_header 'Content-Type' 'text/plain charset=UTF-8';
            add_header 'Content-Length' 0;
            return 204;
        }
    }
}
```

Then restart Nginx:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Step 6: Check Access

Access your API at `http://your-ec2-public-ip/`

## Troubleshooting

1. Check Docker logs:

   ```bash
   docker logs catalog-api
   ```

2. Check if the container is running:

   ```bash
   docker ps
   ```

3. Check Nginx logs:

   ```bash
   sudo tail -f /var/log/nginx/error.log
   ```

4. Make sure your RDS security group allows connections from your EC2 instance.

## Updating Your Deployment

When a new Docker image is available:

```bash
# Pull the latest image
docker pull yourusername/catalogapi:latest

# Stop and remove the current container
docker stop catalog-api
docker rm catalog-api

# Run the new container with both configuration files
docker run -d -p 8080:8080 \
  -v ~/serviceAccountKey.json:/app/serviceAccountKey.json \
  -v ~/.env.production:/app/.env.production \
  --name catalog-api \
  --restart unless-stopped \
  yourusername/catalogapi:latest
```
