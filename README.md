# Release Notes Tool

> Enforce and collect release notes on every Git commit across your organization.

Supports **macOS**, **Linux**, and **Windows** with three interaction modes:
- 🖥️ **Web UI**: Beautiful local HTTP form in your browser
- 🖱️ **Fyne GUI**: Native desktop GUI (legacy)
- ⌨️ **Headless**: CLI automation for CI/CD

## Quick Start

### 1. For Developers (One-time Setup)

```bash
# Clone the repo or get the binary
git clone <your-org-repo> release-notes-tool
cd release-notes-tool

# Build (if needed) or use pre-compiled binary
go build -o release-notes .

# Install globally for all repos
./setup-release-notes.sh

# Test
cd any-existing-repo
git commit --allow-empty -m "test"
# → Browser opens with form
```

### 2. For IT Teams (Deploy to Organization)

**Prerequisites**: All developers must have the tool available.

**Option A: Build from source**
```bash
# Clone repo
git clone <your-org-repo>
cd release-notes-tool

# Cross-compile for all platforms
./scripts/build-all.sh  # or use GitHub Actions: push tag v1.0.0

# Distribute binaries
# → macOS: /usr/local/bin/release-notes
# → Windows: C:\Program Files\release-notes.exe (or user PATH)
# → Linux: /usr/local/bin/release-notes
```

**Option B: Use GitHub Releases**
- Tag a commit: `git tag v1.0.0 && git push origin v1.0.0`
- GitHub Actions auto-builds all platforms
- Download binaries from [Releases](../../releases)

**Option C: Package with Homebrew (macOS/Linux)**
```bash
# Create formula (example)
brew tap your-org/tools
brew install release-notes
```

### 3. Installation for IT

After building/downloading binaries:

**macOS/Linux:**
```bash
# Copy binary to system PATH
sudo cp release-notes /usr/local/bin/
sudo chmod +x /usr/local/bin/release-notes

# Or user-local (no sudo needed)
mkdir -p ~/.local/bin
cp release-notes ~/.local/bin/
# Add to ~/.bashrc or ~/.zshrc:
export PATH="$HOME/.local/bin:$PATH"
```

**Windows:**
```cmd
# Option 1: Copy to Git bin directory
copy release-notes.exe "C:\Program Files\Git\cmd\"

# Option 2: Copy to PATH directory
copy release-notes.exe "C:\Windows\System32\"
```

Then each developer runs:
```bash
./setup-release-notes.sh
```

This globally installs the Git hooks for all future repositories.

## Features

### 🌐 Web UI Mode (Default)
- Responsive HTML form with validation
- Auto-opens in default browser
- Supports dark/light browser themes
- Works across all platforms

### 📋 Form Fields
- **Type** (mandatory): Funzionalità, Correzione Bug, Refactoring, Documentazione, Generico
- **Module** (mandatory): AGENDA, B2B, CATALOG, etc. (configurable in `modules.json`)
- **Title** (optional): Short title
- **Description** (mandatory): ≥ 20 characters (can be excluded if checkbox set)
- **Internal Ticket** (optional): e.g. PROJ-123
- **Client Ticket** (optional): e.g. CLI-456
- **Exclude from Release Notes** (optional): Flag to skip this commit

### 🔄 Git Integration

**Hooks Applied:**
```bash
# prepare-commit-msg: initializes release_notes.json
# pre-commit: launches release-notes tool, requires form submission, auto-stages result
```

**Workflow:**
```
$ git commit -m "fix bug"
  ↓
  prepare-commit-msg hook runs (setup)
  ↓
  User opens editor (normal flow)
  ↓
  pre-commit hook intercepts
  ↓
  Browser opens with form → User fills in release notes
  ↓
  Form submitted → release_notes.json appended → auto-staged
  ↓
  Commit completes
```

**Exit Behavior:**
- Form saved ✅ → Commit proceeds
- Form cancelled/closed ❌ → Commit aborted (exit code 1)

### 📊 Output: `release_notes.json`

```json
{
  "releases": [
    {
      "data": "29/01/2026 14:30:00",
      "tipo": "Correzione Bug",
      "modulo": "CORE",
      "descrizione": "Fixed critical memory leak in cache handler",
      "internalTicket": "PROJ-1234",
      "clientTicket": "CLI-5678",
      "commitAuthor": "john.doe",
      "commitDesc": "fix: memory leak in cache.go",
      "commitDate": "2026-01-29 14:25:30 +0100",
      "commitHash": "abc123def456...",
      "excludedFromReleaseNote": false
    }
  ]
}
```

## Configuration

### Custom Modules

Edit `modules.json`:
```json
[
  "AGENDA",
  "B2B",
  "CORE",
  "CUSTOM_MODULE",
  "YOUR_TEAM"
]
```

All modes will use this list.

### Customize Types

Edit `main.go`, line ~85:
```go
tipi := []string{"Funzionalità", "Correzione Bug", "Refactoring", "Documentazione", "Generico"}
```

Then rebuild.

### Web UI Port

Default: `127.0.0.1:9999`

To change, edit `.git-hooks/pre-commit.new`:
```bash
"$RELEASE_TOOL" --serve :8888 --commit-msg "$COMMIT_MSG_FILE" ...
```

## Usage

### Developers

**Normal workflow** (web UI auto-triggers):
```bash
git commit -m "implement new feature"
# Browser opens automatically
# Fill form and click "Save Note"
# Commit completes
```

**Check saved notes:**
```bash
cat release_notes.json | jq '.releases | length'  # Count entries
```

**Re-apply hooks to existing repo:**
```bash
git init  # or rm -rf .git/hooks && git init
# Hooks copied from template dir
```

### Headless / CI Mode

For automated commits (CI/CD):
```bash
release-notes --headless \
  --commit-msg /path/to/msg \
  --tipo "Correzione Bug" \
  --modulo "CORE" \
  --descrizione "Automated fix applied by CI pipeline, more than 20 chars" \
  --excluded=false
```

Exit code 0 = success, non-zero = validation failed (commit aborts).

### Legacy Fyne GUI

If you prefer the desktop GUI:
```bash
release-notes /path/to/commit/msg
# Opens desktop window instead of browser
```

## Troubleshooting

### "Tool not found" in hook
- Check `release-notes` is in `$PATH`: `which release-notes`
- On macOS/Linux: run `setup-release-notes.sh` again and restart terminal
- On Windows: add directory to `PATH` environment variable and restart Git Bash

### Browser doesn't open
- Check firewall allows localhost:9999
- Manually open: `http://127.0.0.1:9999`
- Check browser default in OS settings

### "Commit aborted" but no form shown
- Check Git pre-commit hook executed: `cat .git/hooks/pre-commit`
- Run test: `git commit --allow-empty -m "debug"` and watch output
- Check error logs: `GIT_TRACE=1 git commit -m "test"` (macOS/Linux)

### Module list empty
- Verify `modules.json` exists in repo root
- Fallback: `["Default Module"]` used

### Permission denied on hooks
```bash
chmod +x .git/hooks/pre-commit
chmod +x .git/hooks/prepare-commit-msg
```

## Development

### Build Locally

```bash
go build -o release-notes .
```

### Test Web UI

```bash
./release-notes --serve :9999 --commit-msg /tmp/test-msg
# Opens http://127.0.0.1:9999
```

### Test Headless

```bash
./release-notes --headless \
  --commit-msg /tmp/test-msg \
  --tipo "Generico" \
  --modulo "TEST" \
  --descrizione "Test description with more than twenty characters here" \
  --excluded=false
```

### Run Tests

```bash
go test -v ./...
```

### Build All Platforms (local)

```bash
./scripts/build-all.sh
# Outputs: release-notes-{os}-{arch}.{tar.gz|zip}
```

### Publish Release

```bash
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions auto-builds and uploads to Releases
```

## Architecture

```
main.go
├── Pure functions (reused by all modes)
│   ├── loadModules()
│   ├── getGitData()
│   └── ValidateAndSaveRelease()
├── Web UI mode
│   ├── Embedded HTML/CSS/JS (webUIHTML const)
│   ├── HTTP handlers (/api/init, /api/save)
│   └── Browser auto-open
├── Fyne GUI mode (legacy)
│   └── Desktop window + form
└── Headless mode
    └── CLI flags + JSON write

.git-hooks/
├── prepare-commit-msg.new (init)
└── pre-commit.new (launch tool, require form)

modules.json → canonical module list
release_notes.json → incremental JSON append (auto-staged)
```

## Cross-Platform Considerations

- ✅ Go stdlib + Fyne handle OS differences
- ✅ Browser open uses `open` (macOS), `xdg-open` (Linux), `cmd /c start` (Windows)
- ✅ Git template system works on all platforms
- ✅ Shell scripts provided for setup (developers adapt for their shell)

## Support & Contributing

- Issues: [GitHub Issues](#)
- Docs: See `.github/copilot-instructions.md` for AI agent guidance
- CI/CD: GitHub Actions workflow in `.github/workflows/cross-compile.yml`

## License

[Your License Here]
