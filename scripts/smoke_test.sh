#!/bin/bash

# TUI Smoke Test
# Quick validation script for development workflow

set -e

echo "💨 Running TUI Smoke Test..."
echo "==========================="

# Ensure we're in the right directory
cd "$(dirname "$0")/.."

# Build check
echo "🔨 Build test..."
go build -o uzi . 
echo "✅ Build successful"

# Basic command test
echo ""
echo "📋 Basic command test..."
uzi ls > /dev/null 2>&1
echo "✅ uzi ls works"

# TUI terminal detection test
echo ""
echo "🖥️  TUI terminal detection test..."
echo 'q' | uzi tui 2>&1 | grep -q "TUI requires a terminal environment"
if [ $? -eq 0 ]; then
    echo "✅ TUI terminal detection works"
else
    echo "❌ TUI terminal detection failed"
    exit 1
fi

# Quick unit test
echo ""
echo "🧪 Quick unit tests..."
go test ./test -run "TestTUIComponents|TestConfigValidation" -timeout 10s > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ Core unit tests pass"
else
    echo "❌ Unit tests failed"
    exit 1
fi

# Package import test
echo ""
echo "📦 Package import test..."
go list -deps ./pkg/tui > /dev/null 2>&1
if [ $? -eq 0 ]; then
    echo "✅ TUI package imports correctly"
else
    echo "❌ TUI package import failed"
    exit 1
fi

echo ""
echo "🎉 Smoke Test Passed!"
echo "===================="
echo "Ready for development."
