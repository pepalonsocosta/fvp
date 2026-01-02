# Release Guide

## Manual Release Process

### 1. Update CHANGELOG.md

Add a new section for your release version:

```markdown
## [1.2.3] - 2024-01-15

### Added
- New feature X

### Changed
- Improved Y

### Fixed
- Bug fix Z
```

### 2. Commit and Push Changes

```bash
git add CHANGELOG.md
git commit -m "chore: prepare release v1.2.3"
git push origin main
```

### 3. Create and Push Tag

```bash
# Create annotated tag
git tag -a v1.2.3 -m "Release v1.2.3"

# Push tag to remote
git push origin v1.2.3
```

### 4. Trigger GitHub Release Workflow

1. Go to **GitHub → Actions → Release** workflow
2. Click **"Run workflow"**
3. Enter the version (e.g., `v1.2.3` or `1.2.3`)
4. Click **"Run workflow"**

The workflow will:
- ✅ Validate the tag exists
- ✅ Build binaries for all Linux architectures (amd64, arm64, arm, 386)
- ✅ Generate checksums
- ✅ Create GitHub release with changelog
- ✅ Upload all binaries as release assets

## Version Bump Types

- **Patch**: `v1.0.0` → `v1.0.1` (bug fixes)
- **Minor**: `v1.0.0` → `v1.1.0` (new features, backward compatible)
- **Major**: `v1.0.0` → `v2.0.0` (breaking changes)

## Quick Reference

```bash
# Full release process
vim CHANGELOG.md                    # Update changelog
git add CHANGELOG.md
git commit -m "chore: prepare release v1.2.3"
git push origin main
git tag -a v1.2.3 -m "Release v1.2.3"
git push origin v1.2.3
# Then trigger workflow on GitHub Actions
```
