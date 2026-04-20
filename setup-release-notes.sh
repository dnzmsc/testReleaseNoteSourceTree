#!/bin/bash
# setup-release-notes.sh
# Configura il git hook prepare-commit-msg per un repository specifico.
#
# Uso:
#   cd /path/to/your/repo
#   /path/to/setup-release-notes.sh
#
# Prerequisiti:
#   - Il binario "release-notes" deve essere nel PATH di sistema
#   - Il repository deve avere un file modules.json (opzionale, usa default se assente)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

echo "📝 Setup Release Notes Hook"
echo "============================"
echo ""

# Verify we're inside a git repository
if ! git rev-parse --show-toplevel &> /dev/null; then
    echo "❌ Errore: non sei dentro un repository Git."
    echo "   Esegui questo script dalla root di un repository."
    exit 1
fi

REPO_ROOT=$(git rev-parse --show-toplevel)
HOOKS_DIR="$REPO_ROOT/.git/hooks"

# Check that the binary is available
if ! command -v release-notes &> /dev/null; then
    echo "⚠️  Il binario 'release-notes' non è nel PATH."
    echo "   Assicurati che i sistemisti lo abbiano installato."
    echo "   Continuo comunque con l'installazione dell'hook..."
    echo ""
fi

# Install hooks
echo "Installazione hook in: $HOOKS_DIR"
mkdir -p "$HOOKS_DIR"

# Backup and install pre-commit hook (launches the GUI, stages release_notes.json)
if [ -f "$HOOKS_DIR/pre-commit" ]; then
    BACKUP="$HOOKS_DIR/pre-commit.bak.$(date +%Y%m%d%H%M%S)"
    cp "$HOOKS_DIR/pre-commit" "$BACKUP"
    echo "  ↳ Hook pre-commit esistente salvato in: $BACKUP"
fi
cp "$SCRIPT_DIR/.git-hooks/pre-commit" "$HOOKS_DIR/pre-commit"
chmod +x "$HOOKS_DIR/pre-commit"
echo "  ✓ Hook pre-commit installato"

# Backup and install prepare-commit-msg hook (ensures release_notes.json exists)
if [ -f "$HOOKS_DIR/prepare-commit-msg" ]; then
    BACKUP="$HOOKS_DIR/prepare-commit-msg.bak.$(date +%Y%m%d%H%M%S)"
    cp "$HOOKS_DIR/prepare-commit-msg" "$BACKUP"
    echo "  ↳ Hook prepare-commit-msg esistente salvato in: $BACKUP"
fi
cp "$SCRIPT_DIR/.git-hooks/prepare-commit-msg" "$HOOKS_DIR/prepare-commit-msg"
chmod +x "$HOOKS_DIR/prepare-commit-msg"
echo "  ✓ Hook prepare-commit-msg installato"

# Create release_notes.json if missing
if [ ! -f "$REPO_ROOT/release_notes.json" ]; then
    echo '{"releases":[]}' > "$REPO_ROOT/release_notes.json"
    echo "  ✓ File release_notes.json creato"
fi

# Create modules.json if missing
if [ ! -f "$REPO_ROOT/modules.json" ]; then
    echo '["Default"]' > "$REPO_ROOT/modules.json"
    echo "  ✓ File modules.json creato (modifica con i moduli del tuo progetto)"
fi

echo ""
echo "✅ Setup completato!"
echo ""
echo "Prossimi passi:"
echo "  1. Modifica modules.json con i moduli del tuo progetto"
echo "  2. Aggiungi modules.json e release_notes.json al repository"
echo "  3. Prova con: git commit --allow-empty -m 'test release notes'"
echo ""
echo "Per disinstallare:"
echo "  rm $HOOKS_DIR/prepare-commit-msg"
