#!/bin/bash

# Pre-deployment test script
# Runs all tests for frontend and backend

set -e  # Exit on error

echo "========================================="
echo "  Folda Finances - Pre-Deployment Tests"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track overall status
OVERALL_STATUS=0

# Backend Tests
echo "${YELLOW}Running Backend Tests...${NC}"
echo "----------------------------------------"
cd backend
export CGO_ENABLED=1
if go test ./... -v; then
    echo "${GREEN}✓ Backend tests passed${NC}"
else
    echo "${RED}✗ Backend tests failed${NC}"
    OVERALL_STATUS=1
fi
cd ..
echo ""

# Frontend Tests
echo "${YELLOW}Running Frontend Tests...${NC}"
echo "----------------------------------------"
cd frontend
if npm test -- --run; then
    echo "${GREEN}✓ Frontend tests passed${NC}"
else
    echo "${RED}✗ Frontend tests failed${NC}"
    OVERALL_STATUS=1
fi
cd ..
echo ""

# Summary
echo "========================================="
if [ $OVERALL_STATUS -eq 0 ]; then
    echo "${GREEN}✓ All tests passed! Ready to deploy.${NC}"
else
    echo "${RED}✗ Some tests failed. Fix issues before deploying.${NC}"
    exit 1
fi
echo "========================================="
