# Integration Tests

This directory contains comprehensive integration tests for the Oreo.io API endpoints.

## Overview

The integration tests validate all API endpoints including:
- Health endpoints (`/health`, `/health/db`, `/health/redis`)
- Authentication endpoints (`/auth/*`)
- Project management (`/projects/*`)
- Dataset management (`/datasets/*`)
- Schema operations (`/schemas/*`)
- Data manipulation (`/data/*`)
- Sample data endpoints (`/sample-data/*`)

## Test Structure

```
tests/integration/
├── main_test.go        # Test setup, configuration, and utilities
├── auth_test.go        # Authentication endpoint tests
├── health_test.go      # Health check endpoint tests
├── project_test.go     # Project CRUD operation tests
├── dataset_test.go     # Dataset upload and management tests
├── schema_test.go      # Schema inference and CRUD tests
├── data_test.go        # Data querying and manipulation tests
├── sample_data_test.go # Sample data endpoint tests
└── .env.test          # Test environment configuration
```

## Prerequisites

1. **Go Environment**: Go 1.19 or higher
2. **Test Database**: PostgreSQL instance for testing
3. **Redis**: Redis instance for caching tests
4. **Backend Service**: The API server should be running

## Running Tests

### Using the Test Runner Script

The easiest way to run integration tests is using the provided script:

```bash
# From the project root
./run-integration-tests.sh
```

This script will:
1. Start the test environment using Docker Compose
2. Wait for services to be ready
3. Run all integration tests
4. Clean up the test environment

### Manual Test Execution

If you prefer to run tests manually:

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Wait for services to be ready
# Then run tests from the backend directory
cd backend
go test -v ./tests/integration/... -timeout 5m

# Cleanup
docker-compose -f docker-compose.test.yml down -v
```

### Running Specific Test Suites

```bash
# Run only authentication tests
go test -v ./tests/integration/ -run TestAuthEndpoints

# Run only project tests
go test -v ./tests/integration/ -run TestProjectEndpoints

# Run only health tests
go test -v ./tests/integration/ -run TestHealthEndpoints

# Run with verbose output and race detection
go test -v -race ./tests/integration/...
```

## Test Configuration

### Environment Variables

Tests use the following environment variables (see `.env.test`):

- `TEST_BACKEND_URL`: Backend API URL (default: http://localhost:8080)
- `TEST_ENV`: Test environment identifier
- `TEST_DATABASE_RESET`: Whether to reset database between tests
- `TEST_CLEANUP_ENABLED`: Whether to cleanup test data

### Database Configuration

Tests require a separate test database to avoid interfering with development data:

```env
DATABASE_NAME=oreo_test_db
DATABASE_USER=oreo_test
DATABASE_PASSWORD=testpassword123
```

### Authentication

Tests use a test user account:
- Email: `testuser@example.com`
- Password: `testpassword123`
- Name: `Test User`

## Test Data

### Sample CSV Files

Tests create temporary CSV files with the following structure:

```csv
name,age,city
John,25,New York
Jane,30,London
Bob,35,Paris
```

### Test Datasets

Large test datasets include:
- Employee data with ID, name, email, age, department, salary
- Mixed data types for schema inference testing
- Edge cases for data validation

## Test Coverage

### Health Endpoints
- ✅ `/health` - Basic health check
- ✅ `/health/db` - Database connectivity
- ✅ `/health/redis` - Redis connectivity

### Authentication
- ✅ User registration
- ✅ User login
- ✅ Token refresh
- ✅ User logout
- ✅ Invalid credentials handling
- ✅ Duplicate registration prevention

### Projects
- ✅ Create project
- ✅ List user projects
- ✅ Get project by ID
- ✅ Update project
- ✅ Delete project
- ✅ Access control validation

### Datasets
- ✅ Upload CSV datasets
- ✅ List user datasets
- ✅ List project datasets
- ✅ Get dataset by ID
- ✅ Delete dataset
- ✅ File validation
- ✅ Error handling

### Schemas
- ✅ Automatic schema inference
- ✅ Create custom schemas
- ✅ Get schema by dataset
- ✅ Update schema
- ✅ Delete schema
- ✅ Field validation

### Data Operations
- ✅ Get dataset data with pagination
- ✅ Query data with filters
- ✅ Sort data
- ✅ Update data rows
- ✅ Delete specific rows
- ✅ Preview mode (100 rows)
- ✅ Full mode (1000 rows)

### Sample Data
- ✅ List available samples
- ✅ Get sample information
- ✅ Preview sample data
- ✅ Download samples

## Error Handling

Tests validate proper error responses:
- `400 Bad Request` - Invalid input data
- `401 Unauthorized` - Missing or invalid authentication
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `409 Conflict` - Resource conflicts (e.g., duplicate email)
- `500 Internal Server Error` - Server errors

## Debugging Tests

### Enable Debug Logging

```bash
# Run tests with debug output
go test -v ./tests/integration/... -args -debug
```

### Check Service Logs

```bash
# View backend logs
docker-compose -f docker-compose.test.yml logs backend

# View database logs
docker-compose -f docker-compose.test.yml logs postgres

# View all logs
docker-compose -f docker-compose.test.yml logs
```

### Database Inspection

```bash
# Connect to test database
docker exec -it oreo-postgres-test psql -U oreo_test -d oreo_test_db

# Check tables
\dt

# Check data
SELECT * FROM users LIMIT 5;
SELECT * FROM projects LIMIT 5;
```

## Performance Tests

Tests include basic performance validation:
- Response times under 1 second for most endpoints
- File upload handling for files up to 50MB
- Pagination performance with large datasets
- Concurrent request handling

## Best Practices

1. **Isolation**: Each test is independent and can run alone
2. **Cleanup**: Tests clean up created resources
3. **Realistic Data**: Use representative test data
4. **Error Cases**: Test both success and failure scenarios
5. **Authentication**: Validate security controls
6. **Performance**: Monitor response times

## Continuous Integration

These tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow step
- name: Run Integration Tests
  run: |
    chmod +x ./run-integration-tests.sh
    ./run-integration-tests.sh --timeout 120
```

## Troubleshooting

### Common Issues

1. **Port Conflicts**: Ensure ports 8080, 5432, 6379 are available
2. **Database Connection**: Check database credentials and connectivity
3. **Service Readiness**: Wait for services to fully initialize
4. **Test Data**: Ensure test database is clean before running tests

### Test Failures

If tests fail:
1. Check service logs for errors
2. Verify database connectivity
3. Ensure proper test data setup
4. Check authentication token generation
5. Validate API endpoint responses

### Clean Reset

```bash
# Complete environment reset
docker-compose -f docker-compose.test.yml down -v --remove-orphans
docker system prune -f
./run-integration-tests.sh
```
