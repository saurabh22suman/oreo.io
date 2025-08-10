#!/bin/bash

# Integration Test Runner for Oreo.io
# This script runs comprehensive integration tests against the Docker development environment

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# Function to check if Docker services are running
check_services() {
    print_status "Checking Docker services..."
    
    if ! docker ps | grep -q "oreo-backend-dev"; then
        print_error "Backend service is not running. Please run './docker-deploy.sh dev-up' first."
        exit 1
    fi
    
    if ! docker ps | grep -q "oreo-frontend-dev"; then
        print_error "Frontend service is not running. Please run './docker-deploy.sh dev-up' first."
        exit 1
    fi
    
    if ! docker ps | grep -q "oreo-postgres-dev"; then
        print_error "PostgreSQL service is not running. Please run './docker-deploy.sh dev-up' first."
        exit 1
    fi
    
    if ! docker ps | grep -q "oreo-redis-dev"; then
        print_error "Redis service is not running. Please run './docker-deploy.sh dev-up' first."
        exit 1
    fi
    
    print_success "All Docker services are running"
}

# Function to wait for services to be healthy
wait_for_services() {
    print_status "Waiting for services to be healthy..."
    
    local max_attempts=30
    local attempt=1
    
    while [ $attempt -le $max_attempts ]; do
        if curl -s http://localhost:8080/health > /dev/null 2>&1; then
            print_success "Backend service is healthy"
            break
        fi
        
        print_status "Attempt $attempt/$max_attempts: Waiting for backend service..."
        sleep 2
        ((attempt++))
    done
    
    if [ $attempt -gt $max_attempts ]; then
        print_error "Backend service did not become healthy within timeout"
        exit 1
    fi
    
    # Check database health
    if curl -s http://localhost:8080/health/db | grep -q "healthy"; then
        print_success "Database connection is healthy"
    else
        print_error "Database connection is not healthy"
        exit 1
    fi
    
    # Check Redis health
    if curl -s http://localhost:8080/health/redis | grep -q "healthy"; then
        print_success "Redis connection is healthy"
    else
        print_error "Redis connection is not healthy"
        exit 1
    fi
}

# Function to run integration tests
run_tests() {
    print_status "Running integration tests..."
    
    cd backend
    
    # Install test dependencies if not already installed
    if ! go list -m github.com/stretchr/testify > /dev/null 2>&1; then
        print_status "Installing test dependencies..."
        go mod tidy
    fi
    
    # Run integration tests with verbose output
    print_status "Executing integration tests..."
    
    if go test -v ./tests/integration/... -timeout=5m; then
        print_success "All integration tests passed!"
        return 0
    else
        print_error "Some integration tests failed!"
        return 1
    fi
}

# Function to generate test report
generate_report() {
    print_status "Generating test report..."
    
    cd backend
    
    # Create test results directory
    mkdir -p ../test-results
    
    # Run tests with JSON output for reporting
    go test -v ./tests/integration/... -timeout=5m -json > ../test-results/integration-test-results.json 2>&1 || true
    
    # Generate HTML report (if go-junit-report is available)
    if command -v go-junit-report > /dev/null 2>&1; then
        cat ../test-results/integration-test-results.json | go-junit-report > ../test-results/integration-tests.xml
        print_success "JUnit XML report generated: test-results/integration-tests.xml"
    fi
    
    print_success "Test results saved to: test-results/integration-test-results.json"
}

# Function to check service logs if tests fail
check_logs() {
    print_status "Checking service logs for debugging..."
    
    echo ""
    print_status "Backend logs (last 20 lines):"
    docker logs --tail 20 oreo-backend-dev
    
    echo ""
    print_status "PostgreSQL logs (last 10 lines):"
    docker logs --tail 10 oreo-postgres-dev
    
    echo ""
    print_status "Redis logs (last 10 lines):"
    docker logs --tail 10 oreo-redis-dev
}

# Function to run specific test
run_specific_test() {
    local test_name=$1
    print_status "Running specific test: $test_name"
    
    cd backend
    go test -v ./tests/integration/... -run "$test_name" -timeout=2m
}

# Function to show help
show_help() {
    echo "Integration Test Runner for Oreo.io"
    echo ""
    echo "Usage: $0 [command] [options]"
    echo ""
    echo "Commands:"
    echo "  run              Run all integration tests (default)"
    echo "  check            Check if services are ready for testing"
    echo "  logs             Show service logs for debugging"
    echo "  report           Generate test report"
    echo "  test <name>      Run specific test by name"
    echo "  help             Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0                           # Run all tests"
    echo "  $0 run                       # Run all tests"
    echo "  $0 check                     # Check service health"
    echo "  $0 test TestUserRegistration # Run specific test"
    echo "  $0 logs                      # Show service logs"
    echo ""
}

# Main execution
main() {
    local command=${1:-run}
    
    case $command in
        "run"|"")
            check_services
            wait_for_services
            if run_tests; then
                print_success "Integration tests completed successfully!"
                exit 0
            else
                print_error "Integration tests failed!"
                check_logs
                exit 1
            fi
            ;;
        "check")
            check_services
            wait_for_services
            print_success "All services are ready for testing!"
            ;;
        "logs")
            check_logs
            ;;
        "report")
            check_services
            wait_for_services
            generate_report
            ;;
        "test")
            if [ -z "$2" ]; then
                print_error "Please specify test name"
                show_help
                exit 1
            fi
            check_services
            wait_for_services
            run_specific_test "$2"
            ;;
        "help"|"-h"|"--help")
            show_help
            ;;
        *)
            print_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
}

# Run main function with all arguments
main "$@"
