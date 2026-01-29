#!/bin/bash
# test-suite.sh
# Quick test suite for all modes

set -e

REPO_ROOT="$(pwd)"
TEST_TMP="/tmp/release-notes-test"

echo "🧪 Release Notes Tool - Test Suite"
echo "===================================="
echo ""

# Setup
mkdir -p "$TEST_TMP"
echo "Test commit message for validation testing" > "$TEST_TMP/commit-msg.txt"

# Color output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[0;33m'
NC='\033[0m' # No Color

test_result() {
  if [ $? -eq 0 ]; then
    echo -e "${GREEN}✓ PASS${NC}: $1"
  else
    echo -e "${RED}✗ FAIL${NC}: $1"
    exit 1
  fi
}

# Test 1: Build
echo ""
echo "━━━ Test 1: Build ━━━"
go build -o "$TEST_TMP/release-notes-test" .
test_result "Binary builds successfully"

# Test 2: Headless - Valid input
echo ""
echo "━━━ Test 2: Headless Mode (Valid) ━━━"
RELEASE_JSON="$TEST_TMP/release_notes.json"
echo '{"releases":[]}' > "$RELEASE_JSON"

cd "$TEST_TMP"
"$TEST_TMP/release-notes-test" --headless \
  --commit-msg "$TEST_TMP/commit-msg.txt" \
  --json-out "$RELEASE_JSON" \
  --tipo "Correzione Bug" \
  --modulo "CORE" \
  --descrizione "This is a valid test description with sufficient length" \
  --excluded=false
test_result "Headless mode saves valid note"

# Test 3: Verify JSON structure
echo ""
echo "━━━ Test 3: JSON Output Validation ━━━"
if grep -q '"tipo": "Correzione Bug"' "$RELEASE_JSON"; then
  test_result "Saved note has correct type"
else
  echo -e "${RED}✗ FAIL${NC}: Note type not found in JSON"
  exit 1
fi

if grep -q '"modulo": "CORE"' "$RELEASE_JSON"; then
  test_result "Saved note has correct module"
else
  echo -e "${RED}✗ FAIL${NC}: Note module not found in JSON"
  exit 1
fi

# Test 4: Headless - Missing required field
echo ""
echo "━━━ Test 4: Headless Mode (Validation Fails) ━━━"
"$TEST_TMP/release-notes-test" --headless \
  --commit-msg "$TEST_TMP/commit-msg.txt" \
  --json-out "$RELEASE_JSON" \
  --tipo "Generico" \
  --modulo "CORE" \
  --descrizione "too short" \
  --excluded=false && {
  echo -e "${RED}✗ FAIL${NC}: Should have rejected short description"
  exit 1
} || {
  echo -e "${GREEN}✓ PASS${NC}: Correctly rejected invalid description"
}

# Test 5: Headless - Excluded note with short description
echo ""
echo "━━━ Test 5: Excluded Note (allows short desc) ━━━"
"$TEST_TMP/release-notes-test" --headless \
  --commit-msg "$TEST_TMP/commit-msg.txt" \
  --json-out "$RELEASE_JSON" \
  --tipo "Generico" \
  --modulo "CORE" \
  --descrizione "short" \
  --excluded=true
test_result "Excluded notes don't enforce 20-char limit"

# Test 6: Verify incremental append
echo ""
echo "━━━ Test 6: Incremental JSON Append ━━━"
COUNT=$(grep -c '"tipo"' "$RELEASE_JSON")
if [ "$COUNT" -ge 2 ]; then
  test_result "Multiple notes stored (found $COUNT)"
else
  echo -e "${RED}✗ FAIL${NC}: Expected ≥2 notes, found $COUNT"
  exit 1
fi

# Test 7: Exit codes
echo ""
echo "━━━ Test 7: Exit Codes ━━━"
"$TEST_TMP/release-notes-test" --headless \
  --commit-msg "$TEST_TMP/commit-msg.txt" \
  --json-out "$RELEASE_JSON" \
  --tipo "Generico" \
  --modulo "CORE" \
  --descrizione "Valid test note for exit code check" \
  --excluded=false
EXIT_CODE=$?
if [ $EXIT_CODE -eq 0 ]; then
  test_result "Exit code 0 on success"
else
  echo -e "${RED}✗ FAIL${NC}: Expected exit 0, got $EXIT_CODE"
  exit 1
fi

# Test 8: Module loading
echo ""
echo "━━━ Test 8: Module Loading ━━━"
if [ -f "$REPO_ROOT/modules.json" ]; then
  test_result "modules.json exists"
else
  echo -e "${YELLOW}⚠ WARN${NC}: modules.json not found (will use fallback)"
fi

# Test 9: Web UI mode (verify port is free)
echo ""
echo "━━━ Test 9: Web UI Port Check ━━━"
if ! lsof -Pi :9999 -sTCP:LISTEN -t >/dev/null ; then
  echo -e "${GREEN}✓ PASS${NC}: Port 9999 available for web UI"
else
  echo -e "${YELLOW}⚠ WARN${NC}: Port 9999 already in use"
fi

# Cleanup
echo ""
echo "━━━ Cleanup ━━━"
cd "$REPO_ROOT"
rm -rf "$TEST_TMP"
echo -e "${GREEN}✓ PASS${NC}: Cleanup complete"

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}✅ All tests passed!${NC}"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "Ready for deployment 🚀"
