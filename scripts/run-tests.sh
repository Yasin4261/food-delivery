#!/bin/bash

# Test runner script for the ecommerce API

echo "🚀 Starting Test Suite for Özgür Mutfak API"
echo "============================================="

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to print colored output
print_status() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
if ! command -v go &> /dev/null; then
    print_error "Go is not installed. Please install Go first."
    exit 1
fi

# Set test environment
export GO_ENV=test
export GIN_MODE=test

print_status "Setting up test environment..."

# Create test results directory
mkdir -p test-results

# Run model tests
print_status "Running Model Tests..."
go test -v ./internal/model/... -coverprofile=test-results/model-coverage.out 2>&1 | tee test-results/model-test.log

if [ $? -eq 0 ]; then
    print_status "✅ Model tests passed!"
else
    print_error "❌ Model tests failed!"
    MODEL_FAILED=true
fi

# Run service tests
print_status "Running Service Tests..."
go test -v ./internal/service/... -coverprofile=test-results/service-coverage.out 2>&1 | tee test-results/service-test.log

if [ $? -eq 0 ]; then
    print_status "✅ Service tests passed!"
else
    print_error "❌ Service tests failed!"
    SERVICE_FAILED=true
fi

# Run handler tests
print_status "Running Handler Tests..."
go test -v ./internal/api/handler/... -coverprofile=test-results/handler-coverage.out 2>&1 | tee test-results/handler-test.log

if [ $? -eq 0 ]; then
    print_status "✅ Handler tests passed!"
else
    print_error "❌ Handler tests failed!"
    HANDLER_FAILED=true
fi

# Run integration tests
print_status "Running Integration Tests..."
go test -v ./tests/... -coverprofile=test-results/integration-coverage.out 2>&1 | tee test-results/integration-test.log

if [ $? -eq 0 ]; then
    print_status "✅ Integration tests passed!"
else
    print_error "❌ Integration tests failed!"
    INTEGRATION_FAILED=true
fi

# Run all tests together for overall coverage
print_status "Running All Tests for Overall Coverage..."
go test -v ./... -coverprofile=test-results/overall-coverage.out 2>&1 | tee test-results/overall-test.log

# Generate coverage report
print_status "Generating Coverage Report..."
go tool cover -html=test-results/overall-coverage.out -o test-results/coverage.html

# Calculate coverage percentage
COVERAGE=$(go tool cover -func=test-results/overall-coverage.out | grep total | awk '{print $3}')
print_status "Overall Test Coverage: $COVERAGE"

# Run race condition tests
print_status "Running Race Condition Tests..."
go test -race -v ./internal/service/... 2>&1 | tee test-results/race-test.log

if [ $? -eq 0 ]; then
    print_status "✅ Race condition tests passed!"
else
    print_warning "⚠️  Race condition tests found issues!"
fi

# Run benchmarks
print_status "Running Benchmarks..."
go test -bench=. -benchmem ./internal/api/handler/... 2>&1 | tee test-results/benchmark.log

# Generate test summary
print_status "Generating Test Summary..."
cat > test-results/test-summary.txt << EOF
Test Summary for Özgür Mutfak API
==================================

Test Results:
- Model Tests: $([ "$MODEL_FAILED" = true ] && echo "FAILED" || echo "PASSED")
- Service Tests: $([ "$SERVICE_FAILED" = true ] && echo "FAILED" || echo "PASSED")
- Handler Tests: $([ "$HANDLER_FAILED" = true ] && echo "FAILED" || echo "PASSED")
- Integration Tests: $([ "$INTEGRATION_FAILED" = true ] && echo "FAILED" || echo "PASSED")

Coverage: $COVERAGE

Test Files Generated:
- Model Test Log: test-results/model-test.log
- Service Test Log: test-results/service-test.log
- Handler Test Log: test-results/handler-test.log
- Integration Test Log: test-results/integration-test.log
- Overall Test Log: test-results/overall-test.log
- Race Test Log: test-results/race-test.log
- Benchmark Log: test-results/benchmark.log
- Coverage Report: test-results/coverage.html

Generated on: $(date)
EOF

# Display summary
print_status "Test Summary:"
cat test-results/test-summary.txt

# Final status
if [ "$MODEL_FAILED" = true ] || [ "$SERVICE_FAILED" = true ] || [ "$HANDLER_FAILED" = true ] || [ "$INTEGRATION_FAILED" = true ]; then
    print_error "❌ Some tests failed! Check the logs for details."
    exit 1
else
    print_status "🎉 All tests passed successfully!"
    echo ""
    print_status "📊 Coverage Report: test-results/coverage.html"
    print_status "📋 Test Summary: test-results/test-summary.txt"
    echo ""
    print_status "Test suite completed successfully! 🚀"
    exit 0
fi
