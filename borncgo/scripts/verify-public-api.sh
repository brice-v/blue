#!/usr/bin/env bash
# verify-public-api.sh - Verify public API interface exposure
#
# This script verifies that all public API packages properly expose interfaces
# and that the codebase compiles correctly.
#
# Usage: ./scripts/verify-public-api.sh

set -euo pipefail

echo "=================================================="
echo "🔍 Born ML - Public API Verification"
echo "=================================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track failures
FAILED=0

# Helper function
check() {
    echo -n "  → $1... "
}

pass() {
    echo -e "${GREEN}✓${NC} $1"
}

fail() {
    echo -e "${RED}✗${NC} $1"
    FAILED=$((FAILED + 1))
}

warn() {
    echo -e "${YELLOW}⚠${NC} $1"
}

echo "📋 Step 1: Verify build"
echo "------------------------"

check "Building examples/mnist"
if go build -o /dev/null ./examples/mnist 2>&1; then
    pass "MNIST example builds"
else
    fail "MNIST example build failed"
fi

check "Building examples/mnist-cnn"
if go build -o /dev/null ./examples/mnist-cnn 2>&1; then
    pass "MNIST-CNN example builds"
else
    fail "MNIST-CNN example build failed"
fi

check "Building all binaries (make build)"
if make build > /dev/null 2>&1; then
    pass "All binaries build successfully"
else
    fail "Binary build failed"
fi

echo ""
echo "🧪 Step 2: Run tests"
echo "--------------------"

check "Running unit tests"
if go test ./... -short 2>&1 | grep -q "FAIL"; then
    fail "Some tests failed"
else
    pass "All tests pass"
fi

echo ""
echo "🔎 Step 3: Lint public API packages"
echo "------------------------------------"

check "Linting public API packages"
LINT_OUTPUT=$(golangci-lint run ./optim/... ./loader/... ./autodiff/... ./generate/... ./tokenizer/... ./backend/cpu/... ./nn/... ./tensor/... 2>&1)
if echo "$LINT_OUTPUT" | grep -qE "^[1-9][0-9]* issues?"; then
    fail "Linter found issues"
else
    pass "No linter issues"
fi

echo ""
echo "✅ Step 4: Verify interface exposure"
echo "-------------------------------------"

# Check that key types are aliases (not redefinitions)
check "optim.Optimizer is type alias"
if grep -q "type Optimizer = optim.Optimizer" optim/optim.go; then
    pass "optim.Optimizer correctly aliased"
else
    fail "optim.Optimizer not properly aliased"
fi

check "tokenizer.Tokenizer is type alias"
if grep -q "type Tokenizer = tokenizer.Tokenizer" tokenizer/tokenizer.go; then
    pass "tokenizer.Tokenizer correctly aliased"
else
    fail "tokenizer.Tokenizer not properly aliased"
fi

check "tensor.Backend is type alias"
if grep -q "type Backend = tensor.Backend" tensor/tensor.go; then
    pass "tensor.Backend correctly aliased"
else
    fail "tensor.Backend not properly aliased"
fi

check "nn.Module is type alias"
if grep -q "type Module\[B tensor.Backend\] = nn.Module\[B\]" nn/module.go; then
    pass "nn.Module correctly aliased"
else
    fail "nn.Module not properly aliased"
fi

# Check that loader.ModelReader is an interface (not alias)
check "loader.ModelReader is interface"
if grep -q "type ModelReader interface" loader/loader.go; then
    pass "loader.ModelReader is interface"
else
    fail "loader.ModelReader not defined as interface"
fi

echo ""
echo "📊 Step 5: Verify documentation"
echo "--------------------------------"

check "Package comments exist"
PKG_COUNT=0
for pkg in optim loader autodiff generate tokenizer tensor nn; do
    if head -20 "$pkg/$pkg.go" 2>/dev/null | grep -q "^// Package $pkg"; then
        PKG_COUNT=$((PKG_COUNT + 1))
    fi
done

if [ $PKG_COUNT -ge 6 ]; then
    pass "Package comments present ($PKG_COUNT/7 packages)"
else
    warn "Some package comments missing ($PKG_COUNT/7 packages)"
fi

echo ""
echo "=================================================="
if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✅ All checks passed!${NC}"
    echo "=================================================="
    exit 0
else
    echo -e "${RED}❌ $FAILED check(s) failed${NC}"
    echo "=================================================="
    exit 1
fi
