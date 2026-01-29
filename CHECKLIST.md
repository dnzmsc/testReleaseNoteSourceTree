# Pre-Deployment Checklist

> Use this checklist before deploying to your team.

## ✅ Code Quality

- [ ] All tests pass: `./test-suite.sh`
- [ ] No compilation warnings (ignore duplicate `-lobjc` on macOS)
- [ ] Code compiles for all platforms: `scripts/build-all.sh`
- [ ] No unused imports or variables
- [ ] JSON parsing/validation reviewed

## ✅ Functionality

- [ ] Web UI mode works: opens browser on commit
- [ ] Headless mode works: CLI-only with exit codes
- [ ] Fyne GUI works: fallback UI functional
- [ ] Form validates correctly: rejects <20 char descriptions
- [ ] Excluded flag works: allows short descriptions when checked
- [ ] Git integration: hooks trigger on commit
- [ ] release_notes.json appends incrementally (no overwrites)
- [ ] Atomic writes: no corruption if process crashes mid-write

## ✅ Cross-Platform

- [ ] Tested on macOS (Intel and/or ARM)
- [ ] Tested on Windows (PowerShell or Git Bash)
- [ ] Tested on Linux (Ubuntu/Debian or equivalent)
- [ ] Browser open works on all platforms
- [ ] Path handling works (forward slashes)
- [ ] Binary sizes reasonable (<50MB each)

## ✅ Deployment

- [ ] GitHub Releases created with binaries (or build locally)
- [ ] `setup-release-notes.sh` tested on target platforms
- [ ] Hooks correctly installed to `~/.git-templates/hooks/`
- [ ] `git init` in new repo applies hooks automatically
- [ ] Existing repos can manually apply hooks
- [ ] Binary installed to PATH correctly

## ✅ Documentation

- [ ] README.md complete and accurate
- [ ] DEPLOYMENT.md reviewed by IT/DevOps
- [ ] IMPLEMENTATION_SUMMARY.md matches reality
- [ ] `.github/copilot-instructions.md` updated
- [ ] Troubleshooting section covers common issues
- [ ] Setup instructions are step-by-step clear

## ✅ Edge Cases

- [ ] Missing `modules.json` → uses fallback ✓
- [ ] Git command failures → graceful fallback ✓
- [ ] Browser open failure → no crash (user can manually open)
- [ ] Commit message with special characters → handles correctly
- [ ] Very long descriptions → doesn't break form
- [ ] Rapid successive commits → each creates own note entry

## ✅ Performance

- [ ] Web server starts quickly (<1s)
- [ ] Form submission completes (<2s)
- [ ] JSON append doesn't slow down with large files
- [ ] No memory leaks (runs fine multiple times)

## ✅ Security

- [ ] Server binds to localhost only (127.0.0.1)
- [ ] No sensitive data in logs
- [ ] File permissions preserved (0644)
- [ ] Atomic writes prevent corruption

## ✅ Team Readiness

- [ ] Rollout plan communicated
- [ ] FAQ prepared for support team
- [ ] IT tested setup on their environment
- [ ] Example repos created for testing
- [ ] Feedback channel open (Slack/email)

## ✅ Version Control

- [ ] Main.go version string (if used)
- [ ] Git tag created: `git tag v1.0.0`
- [ ] Tag pushed: `git push origin v1.0.0`
- [ ] GitHub Actions triggered successfully
- [ ] All 6 binaries built and uploaded

## ✅ Final Sign-Off

- [ ] Tech lead approved
- [ ] IT/DevOps approved
- [ ] Product owner notified
- [ ] Ready for team communication

---

## If Any Checkbox is Unchecked

**Do not deploy yet.** Identify the issue, fix it, and re-test:

```bash
# Re-run tests
./test-suite.sh

# Re-build if code changed
go build -o release-notes .

# Test specific mode
./release-notes --serve :9999 --commit-msg /tmp/msg  # Web UI
./release-notes --headless ...                        # Headless
./release-notes /tmp/msg                              # Fyne GUI
```

---

## Deployment Timeline

| Phase | Time | Owner | Task |
|-------|------|-------|------|
| 1. Pre-check | 15 min | Dev | Verify checklist ✓ |
| 2. Build | 5 min | Dev/CI | Cross-compile binaries |
| 3. IT Review | 30 min | IT | Test setup scripts |
| 4. Announcement | 1 day | PM/Dev | Notify team |
| 5. Distribute | 1-2 days | IT | Provide binaries/docs |
| 6. Developer Setup | 5 min/dev | Developers | Run `setup-release-notes.sh` |
| 7. Monitor | 1 week | Dev+Support | Track issues, iterate |

---

## Support Escalation

**Level 1**: Developer
- Binary not found → Update PATH
- Hooks not applying → Re-run setup-release-notes.sh

**Level 2**: IT Support
- Windows PATH issues → Check System Properties
- Network firewall → Verify port 9999 accessible
- Installation permissions → Use admin/sudo as needed

**Level 3**: Engineering
- Code issues → File GitHub issue
- Validation edge cases → Review main.go ValidateAndSaveRelease()
- CI/CD failures → Check .github/workflows/cross-compile.yml

---

**Deployment Status**: Ready ✅  
**Date Approved**: [DATE]  
**Approved By**: [NAME]

