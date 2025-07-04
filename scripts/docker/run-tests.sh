#!/bin/bash

# Docker Test Runner for Özgür Mutfak API
# This script runs tests inside a Docker container

echo "🐳 Running Tests with Docker for Özgür Mutfak API"
echo "=================================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    print_error "Docker is not installed. Please install Docker first."
    exit 1
fi

# Check if Docker is running
if ! docker info &> /dev/null; then
    print_error "Docker is not running. Please start Docker first."
    exit 1
fi

# Build test image
print_status "Building test Docker image..."
docker build -t ecommerce-test -f Dockerfile.test . 2>&1

if [ $? -ne 0 ]; then
    print_error "Failed to build test Docker image."
    exit 1
fi

print_status "✅ Test Docker image built successfully!"

# Create test results directory
mkdir -p test-results

# Run tests in Docker container
print_status "Running tests in Docker container..."
docker run --rm \
    -v "$(pwd)/test-results:/app/test-results" \
    -e GO_ENV=test \
    -e GIN_MODE=test \
    --name ecommerce-test-runner \
    ecommerce-test \
    /bin/bash -c "
        echo '🔍 Running Go Tests...'
        
        # Run model tests
        echo '📦 Testing Models...'
        go test -v ./internal/model/... -coverprofile=test-results/model-coverage.out 2>&1 | tee test-results/model-test.log
        MODEL_EXIT=\$?
        
        # Run service tests
        echo '⚙️  Testing Services...'
        go test -v ./internal/service/... -coverprofile=test-results/service-coverage.out 2>&1 | tee test-results/service-test.log
        SERVICE_EXIT=\$?
        
        # Run handler tests
        echo '🌐 Testing Handlers...'
        go test -v ./internal/api/handler/... -coverprofile=test-results/handler-coverage.out 2>&1 | tee test-results/handler-test.log
        HANDLER_EXIT=\$?
        
        # Run integration tests
        echo '🔗 Testing Integration...'
        go test -v ./tests/... -coverprofile=test-results/integration-coverage.out 2>&1 | tee test-results/integration-test.log
        INTEGRATION_EXIT=\$?
        
        # Run all tests for overall coverage
        echo '📊 Generating Overall Coverage...'
        go test -v ./... -coverprofile=test-results/overall-coverage.out 2>&1 | tee test-results/overall-test.log
        OVERALL_EXIT=\$?
        
        # Generate coverage report
        go tool cover -html=test-results/overall-coverage.out -o test-results/coverage.html
        
        # Calculate coverage percentage
        COVERAGE=\$(go tool cover -func=test-results/overall-coverage.out | grep total | awk '{print \$3}')
        
        # Run race condition tests
        echo '🏁 Testing Race Conditions...'
        go test -race -v ./internal/service/... 2>&1 | tee test-results/race-test.log
        RACE_EXIT=\$?
        
        # Run benchmarks
        echo '⚡ Running Benchmarks...'
        go test -bench=. -benchmem ./internal/api/handler/... 2>&1 | tee test-results/benchmark.log
        
        # Generate test summary
        cat > test-results/test-summary.txt << EOF
Docker Test Summary for Özgür Mutfak API
=========================================

Test Results:
- Model Tests: \$([ \$MODEL_EXIT -eq 0 ] && echo 'PASSED' || echo 'FAILED')
- Service Tests: \$([ \$SERVICE_EXIT -eq 0 ] && echo 'PASSED' || echo 'FAILED')
- Handler Tests: \$([ \$HANDLER_EXIT -eq 0 ] && echo 'PASSED' || echo 'FAILED')
- Integration Tests: \$([ \$INTEGRATION_EXIT -eq 0 ] && echo 'PASSED' || echo 'FAILED')

Coverage: \$COVERAGE

Test Environment:
- Container OS: \$(cat /etc/os-release | grep PRETTY_NAME | cut -d'=' -f2 | tr -d '\"')
- Go Version: \$(go version)
- Docker: Yes

Test Files Generated:
- Model Test Log: test-results/model-test.log
- Service Test Log: test-results/service-test.log
- Handler Test Log: test-results/handler-test.log
- Integration Test Log: test-results/integration-test.log
- Overall Test Log: test-results/overall-test.log
- Race Test Log: test-results/race-test.log
- Benchmark Log: test-results/benchmark.log
- Coverage Report: test-results/coverage.html

Generated on: \$(date)
EOF

        echo ''
        echo '📋 Test Summary:'
        cat test-results/test-summary.txt
        
        # Exit with appropriate code
        if [ \$MODEL_EXIT -ne 0 ] || [ \$SERVICE_EXIT -ne 0 ] || [ \$HANDLER_EXIT -ne 0 ] || [ \$INTEGRATION_EXIT -ne 0 ]; then
            echo '❌ Some tests failed!'
            exit 1
        else
            echo '✅ All tests passed!'
            exit 0
        fi
    "

TEST_EXIT_CODE=$?

# Clean up
print_status "Cleaning up Docker resources..."
docker rmi ecommerce-test 2>/dev/null || true

# Display results
if [ $TEST_EXIT_CODE -eq 0 ]; then
    print_status "🎉 Docker tests completed successfully!"
    print_status "📊 Coverage Report: test-results/coverage.html"
    print_status "📋 Test Summary: test-results/test-summary.txt"
else
    print_error "❌ Docker tests failed! Check the logs for details."
fi

exit $TEST_EXIT_CODE
