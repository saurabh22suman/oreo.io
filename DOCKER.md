# Docker Deployment Guide

This repository includes comprehensive Docker setup for **development**, **staging**, and **production** environments.

## 🚀 Quick Start

### Prerequisites
- Docker 20.10+ and Docker Compose V2
- Git
- Make sure ports 3000, 5432, 6379, 8080 are available

### Development Environment

```bash
# Make the script executable
chmod +x docker-deploy.sh

# Start development environment
./docker-deploy.sh dev-up

# View logs
./docker-deploy.sh dev-logs

# Stop environment
./docker-deploy.sh dev-down
```

**Access URLs:**
- Frontend: http://localhost:3000
- Backend API: http://localhost:8080
- PostgreSQL: localhost:5432
- Redis: localhost:6379

## 📁 File Structure

```
├── docker-compose.dev.yml      # Development environment
├── docker-compose.staging.yml  # Staging environment  
├── docker-compose.prod.yml     # Production environment
├── .env.docker.dev             # Development environment variables
├── .env.docker.staging         # Staging environment variables
├── .env.docker.prod            # Production environment variables
├── docker-deploy.sh            # Management script
├── backend/
│   ├── Dockerfile              # Backend container definition
│   └── .dockerignore           # Backend ignore file
└── frontend/
    ├── Dockerfile              # Frontend container definition
    ├── nginx.conf              # Nginx configuration
    └── .dockerignore           # Frontend ignore file
```

## 🐳 Environments

### Development
- **Database**: PostgreSQL with persistent volume
- **Cache**: Redis with persistent volume  
- **Backend**: Hot reload, debug mode
- **Frontend**: Development build with proxy to backend
- **Ports**: Frontend (3000), Backend (8080), DB (5432), Redis (6379)

### Staging
- **Database**: PostgreSQL with backups
- **Cache**: Redis with password authentication
- **Backend**: Staging build with performance optimizations
- **Frontend**: Production build with staging API
- **Reverse Proxy**: Nginx with SSL termination
- **Monitoring**: Health checks and logging

### Production
- **Database**: PostgreSQL with automated backups
- **Cache**: Redis with authentication and persistence
- **Backend**: Multiple replicas with load balancing
- **Frontend**: Optimized production build
- **Reverse Proxy**: Nginx with SSL, security headers, rate limiting
- **Monitoring**: Comprehensive health checks

## 🛠️ Management Commands

### Development
```bash
./docker-deploy.sh dev-up        # Start development
./docker-deploy.sh dev-down      # Stop development
./docker-deploy.sh dev-restart   # Restart development
./docker-deploy.sh dev-logs      # View logs
./docker-deploy.sh dev-clean     # Clean + remove volumes
```

### Staging
```bash
./docker-deploy.sh staging-up    # Deploy to staging
./docker-deploy.sh staging-down  # Stop staging
./docker-deploy.sh staging-logs  # View staging logs
```

### Production
```bash
./docker-deploy.sh prod-up       # Deploy to production
./docker-deploy.sh prod-down     # Stop production
./docker-deploy.sh prod-logs     # View production logs
```

### Database Backups
```bash
./docker-deploy.sh backup-dev      # Backup development DB
./docker-deploy.sh backup-staging  # Backup staging DB  
./docker-deploy.sh backup-prod     # Backup production DB
```

### Health Checks
```bash
./docker-deploy.sh health dev      # Check development health
./docker-deploy.sh health staging  # Check staging health
./docker-deploy.sh health prod     # Check production health
```

## 🔧 Configuration

### Environment Variables

Environment-specific configurations are stored in:
- `.env.docker.dev` - Development settings
- `.env.docker.staging` - Staging settings  
- `.env.docker.prod` - Production settings

**Key Variables:**
```bash
# Database
POSTGRES_DB=oreo_dev_db
POSTGRES_USER=oreo_dev_user
POSTGRES_PASSWORD=secure_password

# Backend
ENVIRONMENT=development
JWT_SECRET=your-secure-jwt-secret
USE_MOCK_DB=false

# Frontend  
VITE_API_URL=http://localhost:8080
VITE_ENVIRONMENT=development
```

### Security Notes

🔒 **Important Security Considerations:**

1. **Change Default Passwords**: Update all passwords in production environment files
2. **JWT Secrets**: Use strong, unique JWT secrets for each environment
3. **Database Credentials**: Use different credentials for each environment
4. **SSL Certificates**: Configure proper SSL certificates for staging/production
5. **Environment Files**: Never commit real production credentials to version control

## 🔍 Troubleshooting

### Common Issues

**Port Conflicts:**
```bash
# Check what's using ports
netstat -tulpn | grep :3000
netstat -tulpn | grep :8080

# Kill processes if needed
sudo lsof -ti:3000 | xargs kill -9
```

**Container Issues:**
```bash
# View container logs
docker logs oreo-backend-dev
docker logs oreo-frontend-dev

# Restart specific service
docker-compose -f docker-compose.dev.yml restart backend

# Rebuild containers
docker-compose -f docker-compose.dev.yml build --no-cache
```

**Database Connection Issues:**
```bash
# Check database status
docker exec oreo-postgres-dev pg_isready -U oreo_dev_user

# Connect to database
docker exec -it oreo-postgres-dev psql -U oreo_dev_user -d oreo_dev_db
```

**Clean Start:**
```bash
# Complete cleanup and restart
./docker-deploy.sh dev-clean
./docker-deploy.sh dev-up
```

## 📊 Monitoring

### Health Checks
All services include health checks:
- **Backend**: HTTP health endpoint (`/health`)
- **Frontend**: HTTP availability check
- **Database**: PostgreSQL connection check
- **Redis**: Redis ping check

### Logs
```bash
# View all service logs
./docker-deploy.sh dev-logs

# View specific service logs
docker-compose -f docker-compose.dev.yml logs -f backend
docker-compose -f docker-compose.dev.yml logs -f frontend
```

## 🚀 Deployment Workflow

### Development to Staging
1. Test changes in development environment
2. Commit and push to feature branch
3. Deploy to staging: `./docker-deploy.sh staging-up`
4. Run staging tests
5. Create pull request to main

### Staging to Production  
1. Merge to main branch
2. Create production backup: `./docker-deploy.sh backup-prod`
3. Deploy to production: `./docker-deploy.sh prod-up`
4. Monitor health checks
5. Verify deployment success

## 📈 Performance Optimization

### Production Optimizations
- Multi-stage Docker builds for smaller images
- Nginx gzip compression and caching
- Backend horizontal scaling (multiple replicas)
- Database connection pooling
- Redis caching layer
- Security headers and rate limiting

### Resource Limits
```yaml
# Example resource limits (add to docker-compose files)
deploy:
  resources:
    limits:
      cpus: '0.5'
      memory: 512M
    reservations:
      cpus: '0.25'
      memory: 256M
```

## 🆘 Support

For issues with Docker deployment:
1. Check the troubleshooting section above
2. View container logs for error details
3. Ensure all prerequisites are installed
4. Verify environment variables are correctly set

---

**Happy Dockerizing! 🐳**
