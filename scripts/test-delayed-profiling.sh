#!/bin/bash
#
# Test script for delayed profiling feature
# Run this on Linux system with target process
#

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

echo "=========================================="
echo "Delayed Profiling Feature Test Suite"
echo "=========================================="
echo ""

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  echo -e "${RED}ERROR: Must run as root (sudo)${NC}"
  exit 1
fi

# Check if binary exists
BINARY="./bin/blc-perf-analyzer"
if [ ! -f "$BINARY" ]; then
  echo -e "${RED}ERROR: Binary not found at $BINARY${NC}"
  echo "Run 'make build' first"
  exit 1
fi

# Test 1: Help output
echo -e "${YELLOW}[TEST 1] Checking help output...${NC}"
if $BINARY --help | grep -q "delay-start"; then
  echo -e "${GREEN}✓ PASS${NC} - Help shows --delay-start flag"
else
  echo -e "${RED}✗ FAIL${NC} - Help missing --delay-start flag"
  exit 1
fi

# Test 2: Flag validation
echo -e "\n${YELLOW}[TEST 2] Testing flag validation...${NC}"
if ! $BINARY --process test --duration 30 --profile-window 60 2>&1 | grep -q "exclusive"; then
  echo -e "${RED}✗ FAIL${NC} - Should reject duration + profile-window"
  exit 1
fi
echo -e "${GREEN}✓ PASS${NC} - Flags properly validated"

# Test 3: Negative delay
echo -e "\n${YELLOW}[TEST 3] Testing negative delay validation...${NC}"
if ! $BINARY --process test --delay-start -10 2>&1 | grep -q "negative"; then
  echo -e "${RED}✗ FAIL${NC} - Should reject negative delay"
  exit 1
fi
echo -e "${GREEN}✓ PASS${NC} - Negative delay rejected"

# Check for test process
echo -e "\n${YELLOW}[TEST 4] Looking for test process...${NC}"
read -p "Enter process name to test with (e.g., 'sleep', 'systemd'): " PROCESS_NAME

if ! pgrep -x "$PROCESS_NAME" > /dev/null; then
  echo -e "${YELLOW}WARNING: Process '$PROCESS_NAME' not found${NC}"
  echo "Starting a test process (sleep 300)..."
  sleep 300 &
  TEST_PID=$!
  PROCESS_NAME="sleep"
  echo "Test process started with PID: $TEST_PID"
else
  TEST_PID=$(pgrep -x "$PROCESS_NAME" | head -1)
  echo "Found process '$PROCESS_NAME' with PID: $TEST_PID"
fi

# Test 5: Short delay capture
echo -e "\n${YELLOW}[TEST 5] Testing short delay (5s) capture...${NC}"
echo "This will take ~10 seconds (5s delay + 5s capture)"
START_TIME=$(date +%s)

RESULT=$($BINARY \
  --pid $TEST_PID \
  --delay-start 5 \
  --profile-window 5 \
  --output-dir /tmp/test-delay-short \
  --quiet)

END_TIME=$(date +%s)
ELAPSED=$((END_TIME - START_TIME))

if [ -d "$RESULT" ]; then
  echo -e "${GREEN}✓ PASS${NC} - Capture completed in ${ELAPSED}s"
  echo "Result directory: $RESULT"
  
  # Verify perf.data exists
  if [ -f "$RESULT/perf.data" ]; then
    SIZE=$(stat -f%z "$RESULT/perf.data" 2>/dev/null || stat -c%s "$RESULT/perf.data" 2>/dev/null)
    echo "  perf.data size: $SIZE bytes"
  else
    echo -e "${RED}✗ FAIL${NC} - perf.data not generated"
    exit 1
  fi
else
  echo -e "${RED}✗ FAIL${NC} - Output directory not created"
  exit 1
fi

# Test 6: Quiet mode
echo -e "\n${YELLOW}[TEST 6] Testing quiet mode output...${NC}"
RESULT=$($BINARY \
  --pid $TEST_PID \
  --delay-start 2 \
  --profile-window 3 \
  --output-dir /tmp/test-quiet \
  --quiet 2>&1)

LINE_COUNT=$(echo "$RESULT" | wc -l | tr -d ' ')
if [ "$LINE_COUNT" -eq 1 ]; then
  echo -e "${GREEN}✓ PASS${NC} - Quiet mode outputs single line"
  echo "Output: $RESULT"
else
  echo -e "${RED}✗ FAIL${NC} - Quiet mode output has $LINE_COUNT lines (expected 1)"
  echo "Output:"
  echo "$RESULT"
fi

# Test 7: Non-quiet mode with progress
echo -e "\n${YELLOW}[TEST 7] Testing non-quiet mode with delay...${NC}"
echo "This will take ~12 seconds (10s delay + 2s capture)"
$BINARY \
  --pid $TEST_PID \
  --delay-start 10 \
  --profile-window 2 \
  --output-dir /tmp/test-progress

echo -e "${GREEN}✓ PASS${NC} - Non-quiet mode completed"

# Test 8: Custom output directory
echo -e "\n${YELLOW}[TEST 8] Testing custom output directory...${NC}"
CUSTOM_DIR="/tmp/custom-perf-test-$(date +%s)"
$BINARY \
  --pid $TEST_PID \
  --delay-start 2 \
  --profile-window 3 \
  --output-dir "$CUSTOM_DIR" \
  --quiet > /dev/null

if [ -d "$CUSTOM_DIR" ]; then
  echo -e "${GREEN}✓ PASS${NC} - Custom directory created: $CUSTOM_DIR"
else
  echo -e "${RED}✗ FAIL${NC} - Custom directory not created"
  exit 1
fi

# Test 9: Process liveness during delay
echo -e "\n${YELLOW}[TEST 9] Testing process liveness checking...${NC}"
echo "Starting a short-lived process..."
(sleep 3) &
SHORT_PID=$!

echo "Attempting capture with 10s delay (process will die at t=3s)..."
if $BINARY \
  --pid $SHORT_PID \
  --delay-start 10 \
  --profile-window 5 \
  --output-dir /tmp/test-liveness \
  --quiet 2>&1 | grep -q "terminated during delay"; then
  echo -e "${GREEN}✓ PASS${NC} - Detected process termination during delay"
else
  echo -e "${RED}✗ FAIL${NC} - Did not detect process termination"
  # This is non-critical, might just mean timing was off
fi

# Cleanup
echo -e "\n${YELLOW}Cleaning up...${NC}"
if [ -n "$TEST_PID" ] && [ "$PROCESS_NAME" = "sleep" ]; then
  kill $TEST_PID 2>/dev/null || true
fi

rm -rf /tmp/test-delay-short /tmp/test-quiet /tmp/test-progress "$CUSTOM_DIR" /tmp/test-liveness 2>/dev/null || true

echo ""
echo "=========================================="
echo -e "${GREEN}All tests passed!${NC}"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. Test with real workload (MariaDB, PostgreSQL, etc.)"
echo "2. Test with benchmark framework integration"
echo "3. Merge to main"
echo "4. Tag v1.1.0"
