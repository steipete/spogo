---
summary: "Release checklist for spogo (GitHub release binaries via GoReleaser)"
---

# Releasing `spogo`

Always do **all** steps below (CI + changelog + tag + GitHub release assets). No partial releases.

Shortcut (if you want scripts later): create them to mirror this doc.

Assumptions:
- Repo: `steipete/spogo`
- Binary: `spogo`
- GoReleaser config: `.goreleaser.yaml`

## 0) Prereqs
- Clean working tree on `main`.
- Go toolchain installed (version from `go.mod`).
- CI is green.

## 1) Verify build is green
```sh
./scripts/lint.sh
./scripts/check-coverage.sh 90
```

Confirm GitHub Actions `CI` is green for the commit you’re tagging:
```sh
gh run list -L 5 --branch main
```

## 2) Update changelog
- Update `CHANGELOG.md` for the version you’re releasing.

Example heading:
- `## 0.1.0 - 2026-01-02`

## 3) Commit, tag & push
```sh
git checkout main
git pull

# commit changelog + any release tweaks
git commit -am "release: vX.Y.Z"

git tag -a vX.Y.Z -m "Release X.Y.Z"
git push origin main --tags
```

## 4) Verify GitHub release artifacts
The tag push triggers `.github/workflows/release.yml` (GoReleaser). Ensure it completes successfully and the release has assets.

```sh
gh run list -L 5 --workflow release.yml
gh release view vX.Y.Z
```

Ensure GitHub release notes are not empty (mirror the changelog section).

If the workflow needs a rerun:
```sh
gh workflow run release.yml -f tag=vX.Y.Z
```

## Notes
- GoReleaser publishes binaries for macOS, Linux, and Windows.
