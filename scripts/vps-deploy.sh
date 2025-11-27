#!/bin/bash

# VPS Deployment Script - Compatible with Existing Docker
# This script works with existing Docker installation

set -e  # Exit on any error

echo "🚀 Starting VPS initialization (Docker Compatible)..."

# Configuration
APP_NAME="techtest-indico-be"
DEPLOY_DIR="/opt/techtest-indico-be"
SERVICE_NAME="techtest-indico"
DOMAIN="techtest-indico-be.alifdwt.com"
PROJECTS_DIR="$HOME/projects"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
log_info() {
    echo -e "${GREEN}ℹ️  $1${NC}"
}

log_warn() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    log_error "Please run as root (use sudo)"
    exit 1
fi

# Check existing Docker installation
check_existing_docker() {
    log_info "Checking existing Docker installation..."
    
    if command -v docker >/dev/null 2>&1; then
        log_info "✅ Docker is already installed"
        docker --version
        
        # Check if Docker daemon is running
        if systemctl is-active --quiet docker; then
            log_info "✅ Docker daemon is running"
        else
            log_info "Starting Docker daemon..."
            systemctl start docker
            systemctl enable docker
        fi
        
        # Check docker-compose
        if command -v docker-compose >/dev/null 2>&1; then
            log_info "✅ Docker Compose is available"
            docker-compose --version
        elif docker compose version >/dev/null 2>&1; then
            log_info "✅ Docker Compose plugin is available"
            docker compose version
        else
            log_warn "⚠️ Docker Compose not found, installing..."
            apt-get update
            apt-get install -y docker-compose-plugin
        fi
        
        return 0
    else
        log_error "❌ Docker not found. Please install Docker first."
        return 1
    fi
}

# Install additional dependencies
install_dependencies() {
    log_info "Installing additional dependencies..."
    
    # Update package index
    apt-get update
    
    # Install additional packages needed for this project
    apt-get install -y nginx curl git sqlite3 python3-certbot python3-certbot-nginx
    
    # Ensure current user is in docker group
    if ! groups $SUDO_USER | grep -q docker; then
        log_info "Adding user to docker group..."
        usermod -aG docker $SUDO_USER
        log_warn "⚠️ You may need to logout and login again for docker group to take effect"
    fi
    
    log_info "Dependencies installed successfully!"
}

# Create application directory
setup_app_directory() {
    log_info "Setting up application directory..."
    mkdir -p $DEPLOY_DIR
    cd $DEPLOY_DIR
    chown -R $SUDO_USER:$SUDO_USER $DEPLOY_DIR
}

# Setup Nginx reverse proxy
setup_nginx() {
    log_info "Setting up Nginx reverse proxy..."
    
    # Remove default Nginx configuration
    rm -f /etc/nginx/sites-enabled/default
    
    # Create new site configuration
    cat > /etc/nginx/sites-available/$APP_NAME << EOF
server {
    listen 80;
    server_name $DOMAIN;

    location / {
        proxy_pass http://localhost:2051;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
        
        # WebSocket support
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection "upgrade";
        
        # Timeout settings
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
        
        # Buffer settings
        proxy_buffering on;
        proxy_buffer_size 4k;
        proxy_buffers 8 4k;
    }
}
EOF

    # Enable site
    ln -sf /etc/nginx/sites-available/$APP_NAME /etc/nginx/sites-enabled/
    
    # Test and reload Nginx
    nginx -t && systemctl reload nginx
    
    log_info "Nginx configured successfully!"
}

# Setup SSL certificate with Let's Encrypt
setup_ssl() {
    log_info "Setting up SSL certificate..."
    
    # Check if domain is already resolving
    if nslookup $DOMAIN >/dev/null 2>&1; then
        certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@$DOMAIN --no-redirect || {
            log_warn "SSL setup failed, using HTTP only for now"
            log_warn "You can run SSL setup later with: sudo certbot --nginx -d $DOMAIN"
        }
    else
        log_warn "Domain $DOMAIN is not resolving yet. Skipping SSL setup."
        log_warn "You can run SSL setup later with: sudo certbot --nginx -d $DOMAIN"
    fi
}

# Create systemd service for auto-restart
create_systemd_service() {
    log_info "Creating systemd service..."
    
    # Determine docker-compose command
    DOCKER_COMPOSE_CMD="docker-compose"
    if ! command -v docker-compose >/dev/null 2>&1; then
        DOCKER_COMPOSE_CMD="docker compose"
    fi
    
    cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=Techtest Indico Backend
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$DEPLOY_DIR
ExecStart=$DOCKER_COMPOSE_CMD up -d
ExecStop=$DOCKER_COMPOSE_CMD down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

    # Enable and start service
    systemctl enable $SERVICE_NAME
    systemctl daemon-reload
    
    log_info "Systemd service created successfully!"
}

# Setup log rotation
setup_log_rotation() {
    log_info "Setting up log rotation..."
    cat > /etc/logrotate.d/$APP_NAME << EOF
$DEPLOY_DIR/logs/*.log {
    daily
    missingok
    rotate 7
    compress
    delaycompress
    notifempty
    create 644 root root
    postrotate
        $DOCKER_COMPOSE_CMD restart backend
    endscript
}
EOF

    log_info "Log rotation configured successfully!"
}

# Setup firewall rules
setup_firewall() {
    log_info "Configuring firewall..."
    
    # Install UFW if not present
    apt-get install -y ufw
    
    # Reset firewall rules
    ufw --force reset
    
    # Default policies
    ufw default deny incoming
    ufw default allow outgoing
    
    # Allow SSH
    ufw allow ssh
    
    # Allow HTTP and HTTPS
    ufw allow 80/tcp
    ufw allow 443/tcp
    
    # Allow existing Docker containers (3000, 3010 for portfolio)
    ufw allow 3000/tcp
    ufw allow 3010/tcp
    ufw allow 2051/tcp  # For this application
    
    # Enable firewall
    ufw --force enable
    
    log_info "Firewall configured successfully!"
}

# Health check function
health_check() {
    log_info "Performing health check..."
    
    # Check if application is responding
    if curl -f http://localhost:2051/health > /dev/null 2>&1; then
        log_info "✅ Application is healthy locally!"
    else
        log_warn "⚠️ Application is not running locally (this is normal before first deployment)"
    fi
    
    # Check external domain
    if curl -f http://$DOMAIN/health > /dev/null 2>&1; then
        log_info "✅ External domain is accessible!"
    else
        log_warn "⚠️ External domain not yet accessible (DNS propagation might take time)"
    fi
}

# Manual deployment function
manual_deploy() {
    log_info "Starting manual deployment..."
    
    # Clone or update repository
    if [ -d ".git" ]; then
        log_info "Pulling latest changes..."
        git pull origin main
    else
        log_info "Cloning repository..."
        git clone https://github.com/alifdwt/techtest-indico-be.git .
    fi
    
    # Create environment file
    cat > .env << EOF
# Server Configuration
PORT=2051
GIN_MODE=release

# Database Configuration
DB_HOST=postgres
DB_PORT=2050
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=techtest_indico

# Docker Configuration
COMPOSE_PROJECT_NAME=techtest-indico
EOF
    
    # Determine docker-compose command
    DOCKER_COMPOSE_CMD="docker-compose"
    if ! command -v docker-compose >/dev/null 2>&1; then
        DOCKER_COMPOSE_CMD="docker compose"
    fi
    
    # Build and start services
    log_info "Building and starting services..."
    $DOCKER_COMPOSE_CMD down
    $DOCKER_COMPOSE_CMD build
    $DOCKER_COMPOSE_CMD up -d
    
    # Wait for services to be ready
    log_info "Waiting for services to start..."
    sleep 30
    
    # Run database migrations
    log_info "Running database migrations..."
    $DOCKER_COMPOSE_CMD exec -T postgres psql -U postgres -d techtest_indico -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" || true
    $DOCKER_COMPOSE_CMD exec -T postgres psql -U postgres -d techtest_indico -f db/migration/000001_create_vouchers_table.up.sql || true
    
    health_check
}

# Show deployment status
show_status() {
    log_info "Deployment Status:"
    echo ""
    echo "📍 Application Details:"
    echo "   - API URL: https://$DOMAIN"
    echo "   - Swagger: https://$DOMAIN/swagger/index.html"
    echo "   - Health: https://$DOMAIN/health"
    echo ""
    echo "🔧 Management Commands:"
    echo "   - View logs: docker-compose logs -f"
    echo "   - Restart: systemctl restart $SERVICE_NAME"
    echo "   - Update: cd $DEPLOY_DIR && git pull && docker-compose up -d --build"
    echo ""
    echo "📊 Monitoring:"
    echo "   - Service status: systemctl status $SERVICE_NAME"
    echo "   - Docker status: docker ps"
    echo "   - All containers: docker container ls"
    echo ""
    echo "🐳 Existing Containers:"
    docker container ls
    echo ""
}

# Clean up Docker resources (safe cleanup)
cleanup_docker() {
    log_info "Cleaning up unused Docker resources..."
    
    # Remove unused images only (keep running containers)
    docker image prune -f
    
    # Remove unused volumes
    docker volume prune -f
    
    log_info "Docker cleanup completed!"
}

# Main deployment flow
main() {
    case "$1" in
        "init")
            log_info "Initializing VPS for deployment (Docker Compatible)..."
            check_existing_docker
            install_dependencies
            setup_app_directory
            setup_nginx
            setup_firewall
            create_systemd_service
            setup_log_rotation
            cleanup_docker
            log_info "✅ VPS initialization completed!"
            log_info "Next steps:"
            echo "1. Setup DNS for $DOMAIN"
            echo "2. Run: sudo ./scripts/vps-deploy-compatible.sh deploy"
            echo "3. Configure GitHub Actions for automatic deployment"
            echo ""
            echo "🐳 Your existing Docker containers are preserved!"
            ;;
        "deploy")
            setup_app_directory
            manual_deploy
            show_status
            ;;
        "status")
            show_status
            ;;
        "health")
            health_check
            ;;
        "ssl")
            setup_ssl
            ;;
        "cleanup")
            cleanup_docker
            ;;
        *)
            echo "Usage: $0 {init|deploy|status|health|ssl|cleanup}"
            echo "  init    - Initialize VPS for deployment (Docker compatible)"
            echo "  deploy  - Manual deployment (for testing)"
            echo "  status  - Show deployment status"
            echo "  health  - Perform health check"
            echo "  ssl     - Setup SSL certificate"
            echo "  cleanup - Clean up unused Docker resources"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"