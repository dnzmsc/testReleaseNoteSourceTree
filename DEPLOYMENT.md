# Release Notes Tool - Implementation Guide

## What Changed?

### ✅ Problems Solved

1. **Cross-platform support**: Now builds for Windows, macOS (Intel & ARM), Linux (x86_64 & ARM64)
2. **Beautiful UI**: Modern web-based form instead of CLI-only
3. **Global hook distribution**: Single setup command applies to all repos
4. **Headless mode**: CI/CD automation without UI
5. **Atomic writes**: No JSON corruption on write failures
6. **Error handling**: Robust fallbacks for Git operations

### 📝 New Files

```
.github/workflows/cross-compile.yml     # Auto-build on tags (CI/CD)
.github/copilot-instructions.md         # Updated with new architecture
.git-hooks/prepare-commit-msg.new       # Simplified init hook
.git-hooks/pre-commit.new               # Web UI launch hook
setup-release-notes.sh                  # Developer & IT setup script
README.md                               # Complete documentation
DEPLOYMENT.md                           # This file
```

### 🔧 Modified Files

```
main.go                                 # Complete refactoring:
                                        # - 3 modes (Fyne, Web, Headless)
                                        # - Pure validation functions
                                        # - Embedded web UI (HTML/CSS/JS)
                                        # - HTTP server
                                        # - CLI flag parsing
                                        # - Browser auto-open
go.mod                                  # No new dependencies (Go stdlib + Fyne)
modules.json                            # Unchanged (config)
release_notes.json                      # Unchanged (output format)
```

## Architecture Overview

```
Release Notes Tool (Single Binary)
├── Mode: Web UI (NEW) ⭐
│   └── --serve :PORT --commit-msg FILE
│       ├── Embedded HTML/CSS/JS form
│       ├── HTTP handlers (/api/init, /api/save)
│       └── Auto-opens browser
├── Mode: Headless (NEW) ⭐
│   └── --headless --commit-msg FILE --tipo X --modulo Y --descrizione Z
│       └── Pure CLI, no UI
├── Mode: Fyne GUI (Legacy)
│   └── (no flags) or custom flag
│       └── Desktop window
└── Pure Functions (Reused)
    ├── loadModules()
    ├── getGitData()
    ├── ValidateAndSaveRelease()
    └── openBrowser()
```

## Deployment Workflow

### Phase 1: Build & Publish (IT Team)

```bash
# Option A: Manual build for all platforms
cd release-notes-tool
./scripts/build-all.sh  # or go build for each GOOS/GOARCH
# Outputs: release-notes-macos-x86_64.tar.gz, etc.

# Option B: Auto-build via GitHub Actions (Recommended)
git tag v1.0.0
git push origin v1.0.0
# → GitHub Actions builds all 6 binaries automatically
# → Artifacts available in Releases page
```

### Phase 2: Install Binary (IT or Developers)

**macOS/Linux:**
```bash
# Install globally
sudo curl -fsSL https://github.com/your-org/release-notes/releases/download/v1.0.0/release-notes-macos-arm64.tar.gz \
  | sudo tar xz -C /usr/local/bin/

# Or for non-admin users
curl -fsSL https://...release-notes-macos-arm64.tar.gz | tar xz -C ~/.local/bin/
```

**Windows (PowerShell):**
```powershell
# Download and add to PATH
Invoke-WebRequest -Uri https://github.com/your-org/release-notes/releases/download/v1.0.0/release-notes-windows-x86_64.zip `
  -OutFile release-notes.zip
Expand-Archive -Path release-notes.zip -DestinationPath 'C:\Program Files\release-notes\'
# Add to PATH via System Properties or:
$env:PATH += ';C:\Program Files\release-notes\'
```

### Phase 3: Setup (Each Developer, One-time)

```bash
# Clone repo or just run the setup script
./setup-release-notes.sh

# This:
# 1. Copies hooks to ~/.git-templates/hooks/ (or Windows equivalent)
# 2. Configures git config --global init.templatedir
# 3. Installs binary to PATH

# Test:
git init test-repo
cd test-repo
git commit --allow-empty -m "test"
# → Browser opens with form (or Fyne GUI if --serve not available)
```

### Phase 4: Apply to Existing Repos

For repos created before setup:

```bash
# Option A: Re-initialize (safe for repos with no pending changes)
git init

# Option B: Manual copy
cp ~/.git-templates/hooks/* your-repo/.git/hooks/

# Verify:
ls -la your-repo/.git/hooks/
```

## Configuration

### Change Modules
Edit `modules.json` in the repo root:
```json
["AGENDA", "B2B", "CORE", "MY_MODULE"]
```

### Change Release Types
Edit `main.go` line ~85 and rebuild:
```go
tipi := []string{"Feature", "Bug", "Refactor", "Docs"}
```

### Change Web Port
Edit `.git-hooks/pre-commit.new` and rebuild:
```bash
"$RELEASE_TOOL" --serve :8888  # instead of :9999
```

## Testing Checklist

### 1. Build
```bash
go build -o release-notes .
./release-notes --help  # Verify binary works
```

### 2. Web UI Mode
```bash
echo "Test message" > /tmp/test-msg
./release-notes --serve :9999 --commit-msg /tmp/test-msg
# Browser opens → fill form → verify release_notes.json appended
```

### 3. Headless Mode
```bash
./release-notes --headless \
  --commit-msg /tmp/test-msg \
  --tipo "Correzione Bug" \
  --modulo "CORE" \
  --descrizione "Test with sufficient length"
# Exit code 0 = success
```

### 4. Fyne GUI Mode (Optional)
```bash
./release-notes /tmp/test-msg
# Desktop window opens
```

### 5. Git Hook Integration
```bash
cd /tmp/test-repo && git init
cp .git-hooks/prepare-commit-msg.new .git/hooks/prepare-commit-msg
cp .git-hooks/pre-commit.new .git/hooks/pre-commit
chmod +x .git/hooks/*

git commit --allow-empty -m "test"
# Verify hook triggers
```

## Rollout Recommendations

### Small Team (< 20 devs)
1. Build locally: `go build -o release-notes .`
2. Share binary via email or cloud storage
3. Each dev: `./setup-release-notes.sh`
4. Done!

### Medium Team (20-100 devs)
1. Create GitHub Release with binaries
2. Add to internal wiki/docs
3. Send announcement with setup instructions
4. IT support on-hand for Windows users

### Large Organization (100+ devs)
1. Package via Homebrew, Chocolatey, or internal package manager
2. Pre-install on developer machines via MDM/automation
3. Add to onboarding checklist
4. Provide IT ticket template for support

## Troubleshooting

### "release-notes: command not found"
```bash
# Verify binary in PATH
which release-notes

# If not found:
# 1. Check it exists: ls ~/.local/bin/release-notes
# 2. Add to PATH: export PATH="$HOME/.local/bin:$PATH" >> ~/.bashrc
# 3. Restart shell or: source ~/.bashrc
```

### Browser doesn't open on Linux
```bash
# Fallback: open manually
# Terminal will show URL if browser open fails
xdg-open http://127.0.0.1:9999
```

### Hooks not applied to existing repo
```bash
# Re-run setup to update template
./setup-release-notes.sh

# Then in the repo:
git init  # Re-initialize to copy new hooks
# Or manually:
cp ~/.git-templates/hooks/* .git/hooks/
chmod +x .git/hooks/*
```

### Validation errors (headless mode)
```bash
# Test validation:
./release-notes --headless \
  --commit-msg /tmp/msg \
  --tipo "Generico" \
  --modulo "CORE" \
  --descrizione "a"  # Too short!
# Exit code 1 = validation failed (check output)
```

## Next Steps

1. **Test locally** on macOS, Windows, Linux (each platform if possible)
2. **Build and publish** binaries to GitHub Releases
3. **Announce** setup procedure to team
4. **Monitor** first week for issues
5. **Iterate** on modules.json, types, or UI based on feedback

## Support

- Questions? Check `.github/copilot-instructions.md`
- Issues? See README.md troubleshooting
- CI/CD logs? Check `.github/workflows/cross-compile.yml`

---

**Version**: 1.0.0  
**Last Updated**: 29 January 2026
