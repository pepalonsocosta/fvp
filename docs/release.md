# Release Management

## Version Bump Types

- **Patch**: `v1.0.0` → `v1.0.1`
- **Minor**: `v1.0.0` → `v1.1.0`
- **Major**: `v1.0.0` → `v2.0.0`

## How to Release

1. Merge PRs using **"Create a merge commit"** to preserve commit history
2. Go to GitHub → **Actions** → **Release** workflow
3. Click **"Run workflow"**
4. Select version bump type (patch/minor/major)
5. Click **"Run workflow"**

The release includes all commits since the last release tag. GoReleaser automatically builds binaries and generates the changelog.
