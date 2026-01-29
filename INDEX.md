# 📚 Project File Index

## 🎯 Quick Navigation

**For Different Roles:**
- 👨‍💻 **Developers**: Start with [README.md](README.md) → [setup-release-notes.sh](setup-release-notes.sh)
- 🏢 **IT/DevOps**: Read [DEPLOYMENT.md](DEPLOYMENT.md) → [CHECKLIST.md](CHECKLIST.md)
- 🤖 **AI Agents**: See [.github/copilot-instructions.md](.github/copilot-instructions.md)
- 🧪 **QA/Testing**: Run [test-suite.sh](test-suite.sh)

---

## 📂 File Structure

```
release-notes-tool/
│
├── 📄 README.md                          ← START HERE (Features, quick start, usage)
├── 📄 DEPLOYMENT.md                      ← IT rollout guide (phases, config, troubleshooting)
├── 📄 IMPLEMENTATION_SUMMARY.md           ← High-level changes & improvements
├── 📄 CHECKLIST.md                       ← Pre-deployment verification
├── 📄 INDEX.md                           ← This file
│
├── 🔧 Source Code
│   └── main.go                           ← Single binary (Web UI + Headless + Fyne GUI)
│
├── 🐚 Scripts
│   ├── setup-release-notes.sh            ← ONE-TIME: Install hooks globally
│   ├── test-suite.sh                     ← Verify all modes work
│   └── scripts/build-all.sh              ← Cross-compile for all platforms
│
├── ⚙️ Configuration
│   └── modules.json                      ← List of modules (edit to customize)
│
├── 📊 Data
│   └── release_notes.json                ← Output file (incremental append)
│
├── 🔗 Git Hooks
│   ├── .git-hooks/prepare-commit-msg.new  ← Initialize release_notes.json
│   └── .git-hooks/pre-commit.new          ← Launch tool, require form, auto-stage
│
├── 🚀 CI/CD
│   ├── .github/copilot-instructions.md    ← AI agent guidance
│   └── .github/workflows/cross-compile.yml ← GitHub Actions: build all platforms
│
└── 📦 Builds (after running scripts/build-all.sh)
    ├── release-notes-macos-x86_64.tar.gz
    ├── release-notes-macos-arm64.tar.gz
    ├── release-notes-linux-x86_64.tar.gz
    ├── release-notes-linux-arm64.tar.gz
    ├── release-notes-windows-x86_64.zip
    └── release-notes-windows-arm64.zip
```

---

## 📋 File Descriptions

### Documentation

| File | Purpose | Audience | Read Time |
|------|---------|----------|-----------|
| **README.md** | Complete user guide, features, quick start, troubleshooting | Everyone | 15 min |
| **DEPLOYMENT.md** | Rollout instructions, phases, IT setup, config | IT/DevOps | 20 min |
| **IMPLEMENTATION_SUMMARY.md** | What changed, improvements, comparison | Tech leads | 10 min |
| **CHECKLIST.md** | Pre-deployment verification | QA/PM | 5 min |
| **INDEX.md** | This file — navigation guide | Everyone | 5 min |

### Source Code

| File | Purpose | Lines | Language |
|------|---------|-------|----------|
| **main.go** | Single binary with 3 modes: Web UI, Headless, Fyne GUI | 950+ | Go 1.24 |

**Key Sections of main.go:**
- Lines 1-50: Imports & data structures
- Lines 52-100: `Release` & `ReleaseFile` structs
- Lines 102-130: `loadModules()` — load module config
- Lines 132-180: `getGitData()` — fetch Git commit info
- Lines 182-240: `ValidateAndSaveRelease()` — pure validation + save (reused)
- Lines 242-290: `openBrowser()` — cross-platform browser launch
- Lines 292-500: `webUIHTML` const — embedded HTML/CSS/JS form
- Lines 502-600: HTTP handlers for web UI (`/api/init`, `/api/save`)
- Lines 602-700: `runWebMode()` — HTTP server + browser open
- Lines 702-750: `runHeadlessMode()` — CLI automation
- Lines 752-950: `runFyneMode()` — Desktop GUI (legacy)

### Scripts

| File | Purpose | Trigger | Audience |
|------|---------|---------|----------|
| **setup-release-notes.sh** | Install binary & hooks globally (one-time per developer) | `./setup-release-notes.sh` | Developers |
| **test-suite.sh** | Run comprehensive test suite (validate build) | `./test-suite.sh` | QA/CI |
| **scripts/build-all.sh** | Cross-compile for all 6 platforms + create archives | `./scripts/build-all.sh` | Build engineers |

### Configuration

| File | Purpose | Editable | Default |
|------|---------|----------|---------|
| **modules.json** | List of release note modules/categories | ✅ Yes | 26 modules (AGENDA, B2B, CORE, ...) |

### Git Hooks (in `.git-hooks/`)

| File | Trigger | Action | Exit |
|------|---------|--------|------|
| **prepare-commit-msg.new** | Before commit editor opens | Initialize `release_notes.json` if missing | Always 0 |
| **pre-commit.new** | Before commit is finalized | Launch release-notes tool, require form, auto-stage JSON | 0=success, 1=abort |

### CI/CD

| File | Trigger | Action | Outputs |
|------|---------|--------|---------|
| **.github/copilot-instructions.md** | Reference | AI agent guidance for maintaining codebase | — |
| **.github/workflows/cross-compile.yml** | On tag `v*` push | Build 6 binaries, create GitHub Release | .tar.gz, .zip archives |

---

## 🔄 Typical Workflows

### Workflow 1: First-Time Developer Setup
```
1. Clone repo (or download binary)
2. Read: README.md (5 min)
3. Run: ./setup-release-notes.sh (1 min)
4. Done! Hooks auto-apply to all future repos
```

### Workflow 2: Create Release (IT)
```
1. Read: DEPLOYMENT.md (15 min)
2. Build: ./scripts/build-all.sh (2 min)
3. Tag: git tag v1.0.0 && git push origin v1.0.0
4. Wait: GitHub Actions builds all 6 binaries (5 min)
5. Verify: Check GitHub Releases page
6. Communicate: Share download link with team
```

### Workflow 3: Test Before Deployment
```
1. Run: ./test-suite.sh
2. Review: All tests pass ✅
3. Check: CHECKLIST.md all items verified
4. Deploy!
```

### Workflow 4: Troubleshoot Issue
```
1. Reproduce: cd your-repo && git commit -m "test"
2. Check logs: Look at error message
3. Search: DEPLOYMENT.md → Troubleshooting section
4. Or: README.md → Troubleshooting section
5. If unsure: Check .github/copilot-instructions.md for tech details
```

---

## 🏃 Quick Commands

### Build
```bash
go build -o release-notes .
```

### Test All Modes
```bash
./test-suite.sh
```

### Build for All Platforms
```bash
./scripts/build-all.sh
```

### Setup (Developer)
```bash
./setup-release-notes.sh
```

### Test Web UI
```bash
echo "test" > /tmp/msg
./release-notes --serve :9999 --commit-msg /tmp/msg
# Opens http://127.0.0.1:9999
```

### Test Headless
```bash
./release-notes --headless \
  --commit-msg /tmp/msg \
  --tipo "Correzione Bug" \
  --modulo "CORE" \
  --descrizione "Test description with enough characters"
```

---

## 📊 Metrics

| Metric | Value |
|--------|-------|
| **Binary size** | ~30 MB (Fyne overhead) |
| **Startup time** | <1 second |
| **Form submission** | <2 seconds |
| **Platforms** | 6 (Win/Mac/Linux × x86/ARM) |
| **Modes** | 3 (Web UI, Headless, Fyne) |
| **External dependencies** | 0 (pure Go stdlib + Fyne) |
| **Tests** | 9 (all passing) |

---

## 🔐 File Permissions

```bash
chmod +x setup-release-notes.sh
chmod +x test-suite.sh
chmod +x scripts/build-all.sh
chmod +x .git-hooks/prepare-commit-msg.new
chmod +x .git-hooks/pre-commit.new
```

---

## 🚀 Before You Deploy

- [ ] Read **README.md**
- [ ] Read **DEPLOYMENT.md** (if IT)
- [ ] Run **test-suite.sh**
- [ ] Check **CHECKLIST.md**
- [ ] Build with `./scripts/build-all.sh`
- [ ] Test on target platforms

---

## 📞 Support Matrix

| Question | Answer Location |
|----------|-----------------|
| "How do I use this?" | README.md § Usage |
| "How do I install?" | README.md § Quick Start |
| "Why isn't it working?" | README.md § Troubleshooting |
| "How do I deploy?" | DEPLOYMENT.md |
| "What changed?" | IMPLEMENTATION_SUMMARY.md |
| "Are we ready?" | CHECKLIST.md |
| "How does it work?" | .github/copilot-instructions.md |
| "Code details?" | main.go (inline comments) |

---

## 🎓 Learning Path

**Beginner (User)**
1. README.md — Understand what it does
2. Run `./setup-release-notes.sh` — Get it working
3. Commit and use → form appears automatically

**Intermediate (Developer)**
1. README.md § Configuration — Customize modules
2. main.go § Validation rules — Understand constraints
3. test-suite.sh — Verify behavior

**Advanced (Maintainer)**
1. IMPLEMENTATION_SUMMARY.md — Architecture overview
2. main.go — Full code review
3. .github/copilot-instructions.md — Add features
4. .github/workflows/cross-compile.yml — Understand CI/CD

---

**Last Updated**: 29 January 2026  
**Version**: 1.0.0  
**Status**: ✅ Ready for Deployment
