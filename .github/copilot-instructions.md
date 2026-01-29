# Copilot instructions for this repository

This repository contains a cross-platform Go tool for enforcing and collecting release notes on every Git commit. It supports three modes: legacy Fyne GUI, modern web UI (local HTTP server), and headless (CLI).

Quick intent
- Primary: maintain and extend release notes collection (`main.go`), JSON persistence (`release_notes.json`, `modules.json`), and Git hook integration.
- Secondary: ensure cross-platform distribution (Windows/macOS/Linux, x86_64/arm64) and developer onboarding via setup scripts.

High-level architecture
- **Single Go binary** with three modes:
  - **Legacy Fyne mode** (`--` or no flags): Interactive GUI with Fyne.
  - **Web UI mode** (`--serve :PORT --commit-msg FILE`): Local HTTP server with embedded HTML/CSS/JS, auto-opens browser.
  - **Headless mode** (`--headless --commit-msg FILE --tipo X --modulo Y --descrizione Z`): CLI automation, no UI.
- **Git hook integration**: `prepare-commit-msg` initializes; `pre-commit` launches tool and stages `release_notes.json`.
- **Persistent state**: `release_notes.json` (incremental append) and `modules.json` (module list).
- **Cross-repo**: Git template system (`~/.git-templates/hooks/`) applies hooks globally to all repos.

Project-specific patterns and conventions
- **Refactored architecture**: Validation and save logic extracted into pure functions (`ValidateAndSaveRelease()`, `loadModules()`, `getGitData()`) for reuse across modes.
- **Web UI**: Embedded as HTML string (`webUIHTML` const); includes responsive CSS and vanilla JS for form handling.
- **Flag parsing**: Main dispatch logic (`main()`) examines flags to choose mode: legacy Fyne → web UI → headless → error.
- **Atomic writes**: Release notes saved via temp file + rename to prevent corruption if write fails mid-operation.
- **Browser detection**: `openBrowser()` uses platform-specific commands (macOS: `open`, Linux: `xdg-open`, Windows: `cmd /c start`).
- **Git template approach**: Developers run `setup-release-notes.sh` once; thereafter all new/re-initialized repos inherit hooks via `git config --global init.templatedir`.
- **Exit codes**: Headless and hook return 0 on success, non-zero on error → aborts commit.

Key files to inspect when editing
- `main.go` — all three modes, validation rules, HTTP handlers, and Fyne GUI.
- `.git-hooks/prepare-commit-msg.new` — initializes `release_notes.json` if missing.
- `.git-hooks/pre-commit.new` — launches tool in web UI mode, stages result.
- `setup-release-notes.sh` — one-time developer setup (copies hooks to Git templates, installs binary).
- `.github/workflows/cross-compile.yml` — builds for all platforms on tag.
- `modules.json` — canonical module list (used by all modes).

Build / run / debug notes
- Build: `go build -o release-notes .` produces single executable.
- Web UI test: `./release-notes --serve :9999 --commit-msg .git/COMMIT_EDITMSG` → opens browser at http://127.0.0.1:9999.
- Headless test: `./release-notes --headless --commit-msg .git/COMMIT_EDITMSG --tipo "Correzione Bug" --modulo "CORE" --descrizione "Fixed critical issue with data parsing" --excluded=false`.
- Git hook test: `cd test-repo && git commit --allow-empty -m "test"` → should trigger pre-commit hook → web UI launches.
- Requires: Go 1.24+ (or adjust in `go.mod`); Fyne dependencies (auto-fetched); no external services.

Validation rules (enforced in `ValidateAndSaveRelease()`)
- If `ExcludedFromReleaseNote == false`:
  - `Tipo`, `Modulo` are required (non-empty).
  - `Descrizione` must be ≥ 20 non-space characters.
- If `ExcludedFromReleaseNote == true`:
  - `Descrizione` must be non-empty (but no length requirement).
- Apply same rules in Fyne, web form, and headless.

Behavioral contracts & gotchas
- `modules.json` missing → defaults to `["Default Module"]`; log warning.
- `getGitData()` fallbacks: `git config user.name` fails → use `$USER`; `git log` fails → use `time.Now()`.
- Web server binds to `127.0.0.1:9999` (hardcoded for now); change if needed.
- Hook assumes `release-notes` is in `$PATH`; setup script installs it via `/usr/local/bin` or `~/.local/bin`.
- Single-process write assumption on `release_notes.json`; add file locking if concurrent access needed.

Suggested edits policy for AI agents
- When adding new modes or CLI flags, update `main()` dispatch logic and test all three paths.
- Pure functions (`ValidateAndSaveRelease`, `loadModules`, `getGitData`) are reused; changes affect all modes.
- Web UI form validation mirrors backend validation; keep them in sync.
- Atomic writes are important for reliability; preserve temp file pattern.
- Cross-platform paths: use `/` and let Go handle OS differences; avoid hardcoded separators.
- Test each mode after edits: Fyne GUI, web UI (browser test), headless CLI.

Examples extracted from the codebase
- Validation: `if !isExcluded && len(strings.TrimSpace(release.Descrizione)) < 20 { return error }`.
- Web form error display: JavaScript clears errors, validates, updates DOM with results.
- Atomic write: `WriteFile(tmpFile) → Rename(tmp, target) → Remove(tmp) on error`.
- Browser open: `exec.Command()` with platform-specific args; errors logged but don't fail.

CI/CD & Distribution
- GitHub Actions workflow (`.github/workflows/cross-compile.yml`) builds on tag (`v*`).
- Produces platform-specific binaries: `release-notes-{os}-{arch}.{tar.gz|zip}`.
- Setup script (`setup-release-notes.sh`) copies hooks and binary; developers run once globally.
- For IT teams: test setup script on both macOS and Windows before rollout.

If you need more detail
- Ask me to extract specific functions, add tests (mock `exec.Command`), or scaffold CI steps.
- Consider adding feature flags for web port, or env vars for hook behavior customization.
