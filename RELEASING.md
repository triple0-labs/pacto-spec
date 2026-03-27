# Releasing Pacto

This project ships:

- Go binaries via GitHub Releases (`v*` tags).

## One-time setup

1. Ensure release workflow is healthy:
   - `.github/workflows/release.yml`

## Standard release flow

1. Ensure the intended release changes are committed on `main`, then push:

```bash
git push origin main
```

2. Create and push release tag:

```bash
git tag v<x.y.z>
git push origin v<x.y.z>
```

## What happens in CI

1. `Release` workflow runs on `v*` tag push and publishes Go release artifacts.

## Post-release verification

```bash
pacto update --check
pacto update --yes
pacto version
```
