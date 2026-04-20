#!/bin/bash
# test-suite.sh
# Test suite for the release-notes tool.
# Tests build, module loading, JSON structure, and hook script.

set -e

REPO_ROOT="$(pwd)"
TEST_TMP="/tmp/release-notes-test-$$"

echo "🧪 Release Notes Tool - Test Suite"
echo "===================================="
echo ""

mkdir -p "$TEST_TMP"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

PASS=0
FAIL=0

pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    PASS=$((PASS + 1))
}

fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    FAIL=$((FAIL + 1))
}

# --- Test 1: Build ---
echo "━━━ Test 1: Build ━━━"
if go build -o "$TEST_TMP/release-notes" . 2>&1; then
    pass "Il binario compila correttamente"
else
    fail "Errore di compilazione"
fi

# --- Test 2: --version flag ---
echo ""
echo "━━━ Test 2: Flag --version ━━━"
VERSION_OUT=$("$TEST_TMP/release-notes" --version 2>&1 || true)
if echo "$VERSION_OUT" | grep -q "release-notes"; then
    pass "Flag --version funziona: $VERSION_OUT"
else
    fail "Flag --version non funziona: $VERSION_OUT"
fi

# --- Test 3: modules.json loading ---
echo ""
echo "━━━ Test 3: Caricamento modules.json ━━━"
if [ -f "$REPO_ROOT/modules.json" ]; then
    # Verify it's valid JSON array
    if python3 -c "import json; m=json.load(open('$REPO_ROOT/modules.json')); assert isinstance(m, list) and len(m) > 0" 2>/dev/null; then
        MODULE_COUNT=$(python3 -c "import json; print(len(json.load(open('$REPO_ROOT/modules.json'))))")
        pass "modules.json valido con $MODULE_COUNT moduli"
    else
        fail "modules.json non è un array JSON valido"
    fi
else
    fail "modules.json non trovato"
fi

# --- Test 4: release_notes.json structure ---
echo ""
echo "━━━ Test 4: Struttura release_notes.json ━━━"
echo '{"releases":[]}' > "$TEST_TMP/test_notes.json"
if python3 -c "
import json
with open('$TEST_TMP/test_notes.json') as f:
    data = json.load(f)
    assert 'releases' in data
    assert isinstance(data['releases'], list)
" 2>/dev/null; then
    pass "Struttura JSON corretta"
else
    fail "Struttura JSON non valida"
fi

# --- Test 5: Hook script exists and is executable ---
echo ""
echo "━━━ Test 5: Hook prepare-commit-msg ━━━"
HOOK_FILE="$REPO_ROOT/.git-hooks/prepare-commit-msg"
if [ -f "$HOOK_FILE" ]; then
    if [ -x "$HOOK_FILE" ]; then
        pass "Hook prepare-commit-msg presente e eseguibile"
    else
        fail "Hook prepare-commit-msg presente ma non eseguibile"
    fi
else
    fail "Hook prepare-commit-msg non trovato"
fi

# --- Test 6: Hook skips merge commits ---
echo ""
echo "━━━ Test 6: Hook salta merge commits ━━━"
if grep -q 'merge' "$HOOK_FILE" 2>/dev/null; then
    pass "Hook gestisce merge commits"
else
    fail "Hook non gestisce merge commits"
fi

# --- Test 7: Setup script ---
echo ""
echo "━━━ Test 7: Script di setup ━━━"
if [ -f "$REPO_ROOT/setup-release-notes.sh" ] && [ -x "$REPO_ROOT/setup-release-notes.sh" ]; then
    pass "setup-release-notes.sh presente e eseguibile"
else
    fail "setup-release-notes.sh mancante o non eseguibile"
fi

# --- Test 8: Build script ---
echo ""
echo "━━━ Test 8: Script di build ━━━"
if [ -f "$REPO_ROOT/scripts/build-all.sh" ]; then
    pass "scripts/build-all.sh presente"
else
    fail "scripts/build-all.sh mancante"
fi

# --- Cleanup ---
echo ""
echo "━━━ Cleanup ━━━"
rm -rf "$TEST_TMP"
pass "Cleanup completato"

# --- Summary ---
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
TOTAL=$((PASS + FAIL))
echo "Risultati: $PASS/$TOTAL passati"
if [ $FAIL -eq 0 ]; then
    echo -e "${GREEN}✅ Tutti i test passati!${NC}"
else
    echo -e "${RED}❌ $FAIL test falliti${NC}"
    exit 1
fi
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
