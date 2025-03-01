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
```

For Ubuntu:

```bash
sudo apt update
sudo apt install -y docker.io
sudo systemctl start docker
sudo systemctl enable docker
sudo usermod -a -G docker ubuntu
# Log out and log back in
```

## Step 3: Pull and Run the Docker Image

```bash
# Pull the image from Docker Hub
docker pull yourusername/catalogapi:latest

# Run the container
docker run -d -p 8080:8080 --name catalog-api yourusername/catalogapi:latest

# Check if it's running
docker ps
```

## Step 4: Setup Nginx (if needed)

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
    }
}
```

Then restart Nginx:

```bash
sudo nginx -t
sudo systemctl reload nginx
```

## Step 5: Check Access

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
