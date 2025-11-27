#!/bin/bash

# VPS Deployment Script for GitHub Actions
# This script is designed to work with GitHub Actions deployment

set -e  # Exit on any error

echo "🚀 Starting deployment process..."

# Configuration
APP_NAME="techtest-indico-be"
DEPLOY_DIR="/opt/techtest-indico-be"
SERVICE_NAME="techtest-indico"
DOMAIN="techtest-indico-be.alifdwt.com"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
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

# Install dependencies if not already installed
install_dependencies() {
    log_info "Installing dependencies..."
    apt update
    apt install -y docker.io docker-compose nginx curl git sqlite3
    
    # Start and enable Docker
    systemctl start docker
    systemctl enable docker
    
    # Add current user to docker group
    usermod -aG docker $SUDO_USER
    
    log_info "Dependencies installed successfully!"
}

# Create application directory
setup_app_directory() {
    log_info "Setting up application directory..."
    mkdir -p $DEPLOY_DIR
    cd $DEPLOY_DIR
}

# Setup Nginx reverse proxy
setup_nginx() {
    log_info "Setting up Nginx reverse proxy..."
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
    certbot --nginx -d $DOMAIN --non-interactive --agree-tos --email admin@$DOMAIN || {
        log_warn "SSL setup failed, using HTTP only for now"
    }
}

# Create systemd service for auto-restart
create_systemd_service() {
    log_info "Creating systemd service..."
    cat > /etc/systemd/system/$SERVICE_NAME.service << EOF
[Unit]
Description=Techtest Indico Backend
Requires=docker.service
After=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=$DEPLOY_DIR
ExecStart=/usr/bin/docker-compose up -d
ExecStop=/usr/bin/docker-compose down
TimeoutStartSec=0

[Install]
WantedBy=multi-user.target
EOF

    # Enable and start service
    systemctl enable $SERVICE_NAME
    systemctl start $SERVICE_NAME
    
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
        docker-compose restart backend
    endscript
}
EOF

    log_info "Log rotation configured successfully!"
}

# Health check function
health_check() {
    log_info "Performing health check..."
    
    # Check if application is responding
    if curl -f http://localhost:2051/health > /dev/null 2>&1; then
        log_info "✅ Application is healthy locally!"
    else
        log_error "❌ Application health check failed!"
        docker-compose logs backend
        return 1
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
    
    # Build and start services
    log_info "Building and starting services..."
    docker-compose down
    docker-compose build
    docker-compose up -d
    
    # Wait for services to be ready
    log_info "Waiting for services to start..."
    sleep 30
    
    # Run database migrations
    log_info "Running database migrations..."
    docker-compose exec -T postgres psql -U postgres -d techtest_indico -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" || true
    docker-compose exec -T postgres psql -U postgres -d techtest_indico -f db/migration/000001_create_vouchers_table.up.sql || true
    
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
    echo ""
}

# Main deployment flow
main() {
    case "$1" in
        "init")
            install_dependencies
            setup_app_directory
            setup_nginx
            setup_ssl
            create_systemd_service
            setup_log_rotation
            log_info "VPS initialization completed!"
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
        *)
            echo "Usage: $0 {init|deploy|status|health}"
            echo "  init   - Initialize VPS for deployment"
            echo "  deploy - Manual deployment (for testing)"
            echo "  status - Show deployment status"
            echo "  health - Perform health check"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"