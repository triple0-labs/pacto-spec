# Releasing Pacto (Go + npm)

This project ships:

- Go binaries via GitHub Releases (`v*` tags).
- npm wrapper package `@triple0-labs/pacto-spec`.

## One-time setup (Trusted Publishing)

1. Ensure npm scope/package ownership is correct:
   - package: `@triple0-labs/pacto-spec`
2. Configure npm Trusted Publisher for this package:
   - Provider: GitHub Actions
   - Owner: `triple0-labs`
   - Repository: `pacto-spec`
   - Workflow file: `.github/workflows/npm-publish.yml`
   - Environment: _(none required)_
3. Ensure release workflow is healthy:
   - `.github/workflows/release.yml`
4. Ensure npm workflow is present with OIDC permissions:
   - `.github/workflows/npm-publish.yml` (`id-token: write`)

No `NPM_TOKEN` secret is required when Trusted Publishing is configured correctly.

## Standard release flow

1. Bump `package.json` version to target release version:

```bash
npm version <x.y.z> --no-git-tag-version
```

2. Commit and push:

```bash
git add package.json
git commit -m "chore(release): bump npm wrapper to v<x.y.z>"
git push origin main
```

3. Create and push release tag:

```bash
git tag v<x.y.z>
git push origin v<x.y.z>
```

## What happens in CI

1. `Release` workflow runs on `v*` tag push and publishes Go release artifacts.
2. `NPM Publish` workflow runs on `release.published`:
   - checks out the same tag
   - verifies `package.json.version == <tag without v>`
   - verifies release has `checksums.txt` and `pacto_*.tar.gz` artifacts
   - publishes to npm with provenance via OIDC Trusted Publishing

If any validation fails, npm publish is blocked.

## Post-release verification

```bash
npx -y @triple0-labs/pacto-spec@<x.y.z> version
```

Optional global install check:

```bash
npm i -g @triple0-labs/pacto-spec@<x.y.z>
pacto version
```
