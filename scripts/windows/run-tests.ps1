# PowerShell Test Runner Script for Özgür Mutfak API

param(
    [switch]$Verbose,
    [switch]$Coverage,
    [switch]$Race,
    [switch]$Bench,
    [string]$TestPattern = "."
)

Write-Host "🚀 Starting Test Suite for Özgür Mutfak API" -ForegroundColor Green
Write-Host "=============================================" -ForegroundColor Green

# Function to print colored output
function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor Yellow
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor Red
}

# Check if Go is installed
if (!(Get-Command go -ErrorAction SilentlyContinue)) {
    Write-Error "Go is not installed. Please install Go first."
    exit 1
}

# Set test environment
$env:GO_ENV = "test"
$env:GIN_MODE = "test"

Write-Status "Setting up test environment..."

# Create test results directory
if (!(Test-Path "test-results")) {
    New-Item -ItemType Directory -Path "test-results" | Out-Null
}

# Initialize test results
$testResults = @{
    ModelPassed = $false
    ServicePassed = $false
    HandlerPassed = $false
    IntegrationPassed = $false
    OverallCoverage = "0%"
}

# Run model tests
Write-Status "Running Model Tests..."
if ($Coverage) {
    $modelTest = go test -v ./internal/model/... -coverprofile=test-results/model-coverage.out 2>&1
} else {
    $modelTest = go test -v ./internal/model/... 2>&1
}

$modelTest | Out-File -FilePath "test-results/model-test.log"
Write-Host $modelTest

if ($LASTEXITCODE -eq 0) {
    Write-Status "✅ Model tests passed!"
    $testResults.ModelPassed = $true
} else {
    Write-Error "❌ Model tests failed!"
}

# Run service tests
Write-Status "Running Service Tests..."
if ($Coverage) {
    $serviceTest = go test -v ./internal/service/... -coverprofile=test-results/service-coverage.out 2>&1
} else {
    $serviceTest = go test -v ./internal/service/... 2>&1
}

$serviceTest | Out-File -FilePath "test-results/service-test.log"
Write-Host $serviceTest

if ($LASTEXITCODE -eq 0) {
    Write-Status "✅ Service tests passed!"
    $testResults.ServicePassed = $true
} else {
    Write-Error "❌ Service tests failed!"
}

# Run handler tests
Write-Status "Running Handler Tests..."
if ($Coverage) {
    $handlerTest = go test -v ./internal/api/handler/... -coverprofile=test-results/handler-coverage.out 2>&1
} else {
    $handlerTest = go test -v ./internal/api/handler/... 2>&1
}

$handlerTest | Out-File -FilePath "test-results/handler-test.log"
Write-Host $handlerTest

if ($LASTEXITCODE -eq 0) {
    Write-Status "✅ Handler tests passed!"
    $testResults.HandlerPassed = $true
} else {
    Write-Error "❌ Handler tests failed!"
}

# Run integration tests
Write-Status "Running Integration Tests..."
if ($Coverage) {
    $integrationTest = go test -v ./tests/... -coverprofile=test-results/integration-coverage.out 2>&1
} else {
    $integrationTest = go test -v ./tests/... 2>&1
}

$integrationTest | Out-File -FilePath "test-results/integration-test.log"
Write-Host $integrationTest

if ($LASTEXITCODE -eq 0) {
    Write-Status "✅ Integration tests passed!"
    $testResults.IntegrationPassed = $true
} else {
    Write-Error "❌ Integration tests failed!"
}

# Run all tests together for overall coverage
if ($Coverage) {
    Write-Status "Running All Tests for Overall Coverage..."
    $overallTest = go test -v ./... -coverprofile=test-results/overall-coverage.out 2>&1
    $overallTest | Out-File -FilePath "test-results/overall-test.log"
    Write-Host $overallTest

    # Generate coverage report
    Write-Status "Generating Coverage Report..."
    go tool cover -html=test-results/overall-coverage.out -o test-results/coverage.html

    # Calculate coverage percentage
    $coverageOutput = go tool cover -func=test-results/overall-coverage.out | Select-String "total"
    if ($coverageOutput) {
        $testResults.OverallCoverage = ($coverageOutput.ToString().Split())[-1]
    }
    Write-Status "Overall Test Coverage: $($testResults.OverallCoverage)"
}

# Run race condition tests
if ($Race) {
    Write-Status "Running Race Condition Tests..."
    $raceTest = go test -race -v ./internal/service/... 2>&1
    $raceTest | Out-File -FilePath "test-results/race-test.log"
    Write-Host $raceTest

    if ($LASTEXITCODE -eq 0) {
        Write-Status "✅ Race condition tests passed!"
    } else {
        Write-Warning "⚠️  Race condition tests found issues!"
    }
}

# Run benchmarks
if ($Bench) {
    Write-Status "Running Benchmarks..."
    $benchTest = go test -bench=. -benchmem ./internal/api/handler/... 2>&1
    $benchTest | Out-File -FilePath "test-results/benchmark.log"
    Write-Host $benchTest
}

# Generate test summary
Write-Status "Generating Test Summary..."
$timestamp = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$summaryContent = @"
Test Summary for Özgür Mutfak API
==================================

Test Results:
- Model Tests: $(if ($testResults.ModelPassed) { "PASSED" } else { "FAILED" })
- Service Tests: $(if ($testResults.ServicePassed) { "PASSED" } else { "FAILED" })
- Handler Tests: $(if ($testResults.HandlerPassed) { "PASSED" } else { "FAILED" })
- Integration Tests: $(if ($testResults.IntegrationPassed) { "PASSED" } else { "FAILED" })

Coverage: $($testResults.OverallCoverage)

Test Files Generated:
- Model Test Log: test-results/model-test.log
- Service Test Log: test-results/service-test.log
- Handler Test Log: test-results/handler-test.log
- Integration Test Log: test-results/integration-test.log
$(if ($Coverage) { "- Overall Test Log: test-results/overall-test.log" })
$(if ($Race) { "- Race Test Log: test-results/race-test.log" })
$(if ($Bench) { "- Benchmark Log: test-results/benchmark.log" })
$(if ($Coverage) { "- Coverage Report: test-results/coverage.html" })

Generated on: $timestamp
"@

$summaryContent | Out-File -FilePath "test-results/test-summary.txt"

# Display summary
Write-Status "Test Summary:"
Write-Host $summaryContent

# Final status
$allPassed = $testResults.ModelPassed -and $testResults.ServicePassed -and $testResults.HandlerPassed -and $testResults.IntegrationPassed

if (-not $allPassed) {
    Write-Error "❌ Some tests failed! Check the logs for details."
    exit 1
} else {
    Write-Status "🎉 All tests passed successfully!"
    Write-Host ""
    if ($Coverage) {
        Write-Status "📊 Coverage Report: test-results/coverage.html"
    }
    Write-Status "📋 Test Summary: test-results/test-summary.txt"
    Write-Host ""
    Write-Status "Test suite completed successfully! 🚀"
    exit 0
}

# Usage examples:
# .\run-tests.ps1                    # Run basic tests
# .\run-tests.ps1 -Coverage          # Run tests with coverage
# .\run-tests.ps1 -Race              # Run tests with race detection
# .\run-tests.ps1 -Bench             # Run tests with benchmarks
# .\run-tests.ps1 -Coverage -Race -Bench  # Run all tests with all options
# .\run-tests.ps1 -TestPattern "User"     # Run only tests matching pattern
