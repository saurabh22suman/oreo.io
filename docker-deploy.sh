#!/bin/bash

# Docker Management Script for Oreo.io
# Usage: ./docker-deploy.sh [command]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Function to print colored output
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Development environment functions
dev_up() {
    print_info "Starting development environment..."
    docker-compose -f docker-compose.dev.yml --env-file .env.docker.dev up -d
    print_success "Development environment started!"
    print_info "Frontend: http://localhost:3000"
    print_info "Backend API: http://localhost:8080"
    print_info "PostgreSQL: localhost:5432"
    print_info "Redis: localhost:6379"
}

dev_down() {
    print_info "Stopping development environment..."
    docker-compose -f docker-compose.dev.yml down
    print_success "Development environment stopped!"
}

dev_logs() {
    print_info "Showing development environment logs..."
    docker-compose -f docker-compose.dev.yml logs -f
}

dev_restart() {
    print_info "Restarting development environment..."
    dev_down
    dev_up
}

dev_clean() {
    print_warning "Cleaning development environment (this will remove volumes)..."
    read -p "Are you sure? This will delete all data! (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose -f docker-compose.dev.yml down -v
        docker system prune -f
        print_success "Development environment cleaned!"
    else
        print_info "Operation cancelled"
    fi
}

# Staging environment functions
staging_up() {
    print_info "Starting staging environment..."
    docker-compose -f docker-compose.staging.yml --env-file .env.docker.staging up -d
    print_success "Staging environment started!"
}

staging_down() {
    print_info "Stopping staging environment..."
    docker-compose -f docker-compose.staging.yml down
    print_success "Staging environment stopped!"
}

staging_logs() {
    print_info "Showing staging environment logs..."
    docker-compose -f docker-compose.staging.yml logs -f
}

# Production environment functions
prod_up() {
    print_warning "Starting production environment..."
    read -p "Are you sure you want to deploy to production? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose -f docker-compose.prod.yml --env-file .env.docker.prod up -d
        print_success "Production environment started!"
    else
        print_info "Production deployment cancelled"
    fi
}

prod_down() {
    print_warning "Stopping production environment..."
    read -p "Are you sure you want to stop production? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        docker-compose -f docker-compose.prod.yml down
        print_success "Production environment stopped!"
    else
        print_info "Operation cancelled"
    fi
}

prod_logs() {
    print_info "Showing production environment logs..."
    docker-compose -f docker-compose.prod.yml logs -f
}

# Database backup functions
backup_dev() {
    print_info "Creating development database backup..."
    timestamp=$(date +%Y%m%d_%H%M%S)
    mkdir -p backup
    docker exec oreo-postgres-dev pg_dump -U oreo_dev_user oreo_dev_db > backup/dev_backup_${timestamp}.sql
    print_success "Development backup created: backup/dev_backup_${timestamp}.sql"
}

backup_staging() {
    print_info "Creating staging database backup..."
    timestamp=$(date +%Y%m%d_%H%M%S)
    mkdir -p backup
    docker exec oreo-postgres-staging pg_dump -U oreo_staging_user oreo_staging_db > backup/staging_backup_${timestamp}.sql
    print_success "Staging backup created: backup/staging_backup_${timestamp}.sql"
}

backup_prod() {
    print_info "Creating production database backup..."
    timestamp=$(date +%Y%m%d_%H%M%S)
    mkdir -p backup
    docker exec oreo-postgres-prod pg_dump -U oreo_prod_user oreo_prod_db > backup/prod_backup_${timestamp}.sql
    print_success "Production backup created: backup/prod_backup_${timestamp}.sql"
}

# Health check functions
health_check() {
    environment=${1:-dev}
    
    case $environment in
        "dev")
            print_info "Checking development environment health..."
            docker-compose -f docker-compose.dev.yml ps
            ;;
        "staging")
            print_info "Checking staging environment health..."
            docker-compose -f docker-compose.staging.yml ps
            ;;
        "prod")
            print_info "Checking production environment health..."
            docker-compose -f docker-compose.prod.yml ps
            ;;
        *)
            print_error "Invalid environment. Use: dev, staging, or prod"
            exit 1
            ;;
    esac
}

# Show help
show_help() {
    echo "Docker Management Script for Oreo.io"
    echo ""
    echo "Usage: $0 [command]"
    echo ""
    echo "Development Commands:"
    echo "  dev-up        Start development environment"
    echo "  dev-down      Stop development environment"
    echo "  dev-restart   Restart development environment"
    echo "  dev-logs      Show development logs"
    echo "  dev-clean     Clean development environment (removes volumes)"
    echo ""
    echo "Staging Commands:"
    echo "  staging-up    Start staging environment"
    echo "  staging-down  Stop staging environment"
    echo "  staging-logs  Show staging logs"
    echo ""
    echo "Production Commands:"
    echo "  prod-up       Start production environment"
    echo "  prod-down     Stop production environment"
    echo "  prod-logs     Show production logs"
    echo ""
    echo "Backup Commands:"
    echo "  backup-dev    Backup development database"
    echo "  backup-staging Backup staging database"
    echo "  backup-prod   Backup production database"
    echo ""
    echo "Utility Commands:"
    echo "  health [env]  Check environment health (dev/staging/prod)"
    echo "  help          Show this help message"
    echo ""
}

# Main script logic
case "$1" in
    "dev-up") dev_up ;;
    "dev-down") dev_down ;;
    "dev-restart") dev_restart ;;
    "dev-logs") dev_logs ;;
    "dev-clean") dev_clean ;;
    "staging-up") staging_up ;;
    "staging-down") staging_down ;;
    "staging-logs") staging_logs ;;
    "prod-up") prod_up ;;
    "prod-down") prod_down ;;
    "prod-logs") prod_logs ;;
    "backup-dev") backup_dev ;;
    "backup-staging") backup_staging ;;
    "backup-prod") backup_prod ;;
    "health") health_check $2 ;;
    "help"|"") show_help ;;
    *) 
        print_error "Unknown command: $1"
        show_help
        exit 1
        ;;
esac
