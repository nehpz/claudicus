#!/bin/bash

# TUI Test Suite Runner
# Runs all automated tests for the TUI implementation

set -e  # Exit on any error

echo "🧪 Running TUI Test Suite..."
echo "=================================="

# Ensure we're in the right directory
cd "$(dirname "$0")/.."

# Check if uzi binary exists
if [ ! -f "uzi" ]; then
    echo "⚠️  Building uzi binary first..."
    go build -o uzi .
    echo "✅ Binary built successfully"
fi

echo ""
echo "📋 Running Regression Tests..."
echo "------------------------------"
go test -v ./test -run "Test.*Regression|TestConfig|TestTUIComponents|TestPortAssignment" -timeout 30s

echo ""
echo "🔗 Running Integration Tests..."  
echo "------------------------------"
go test -v ./test -run "TestUziCommand|TestUziCLI|TestState|TestSession|TestTmux" -timeout 60s

echo ""
echo "🏗️  Testing Build Process..."
echo "----------------------------"
echo "Testing clean build..."
go build -o uzi .
if [ $? -eq 0 ]; then
    echo "✅ Clean build successful"
else
    echo "❌ Build failed"
    exit 1
fi

echo ""
echo "📦 Testing Package Imports..."
echo "----------------------------"
echo "Testing TUI package can be imported..."
go list -deps ./pkg/tui > /dev/null
if [ $? -eq 0 ]; then
    echo "✅ TUI package imports successfully"
else
    echo "❌ TUI package import failed"
    exit 1
fi

echo ""
echo "🔍 Running Static Analysis..."
echo "----------------------------"
echo "Running go vet..."
go vet ./...
if [ $? -eq 0 ]; then
    echo "✅ go vet passed"
else
    echo "❌ go vet found issues"
    exit 1
fi

echo "Running go fmt check..."
UNFORMATTED=$(go fmt ./...)
if [ -z "$UNFORMATTED" ]; then
    echo "✅ Code is properly formatted"
else
    echo "❌ Code formatting issues found:"
    echo "$UNFORMATTED"
    exit 1
fi

echo ""
echo "🎯 Testing TUI Terminal Detection..."
echo "-----------------------------------"
echo "Testing TUI with non-terminal input..."
echo 'q' | uzi tui 2>&1 | grep -q "TUI requires a terminal environment"
if [ $? -eq 0 ]; then
    echo "✅ TUI correctly detects non-terminal environment"
else
    echo "❌ TUI terminal detection not working"
    exit 1
fi

echo ""
echo "🎉 All Tests Passed!"
echo "===================="
echo "✅ Regression tests: PASS"
echo "✅ Integration tests: PASS" 
echo "✅ Build process: PASS"
echo "✅ Package imports: PASS"
echo "✅ Static analysis: PASS"
echo "✅ Terminal detection: PASS"
echo ""
echo "TUI implementation is stable and ready for development."
