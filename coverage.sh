#!/bin/bash

# Generate coverage report
go test -coverprofile=coverage.out ./...

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html

# Show function-level coverage
echo "=== Function Coverage ==="
go tool cover -func=coverage.out

# Calculate total coverage
TOTAL_COVERAGE=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
echo "=== Total Coverage: ${TOTAL_COVERAGE}% ==="

# Set minimum coverage threshold
MIN_COVERAGE=70
if (( $(echo "$TOTAL_COVERAGE < $MIN_COVERAGE" | bc -l) )); then
    echo "❌ Coverage ${TOTAL_COVERAGE}% is below minimum ${MIN_COVERAGE}%"
    exit 1
else
    echo "✅ Coverage ${TOTAL_COVERAGE}% meets minimum ${MIN_COVERAGE}%"
fi