# Releasing Pacto (Go + npm)

This project ships:

- Go binaries via GitHub Releases (`v*` tags).
- npm wrapper package `@triple0-labs/pacto-spec`.

## One-time setup

1. Add npm token in GitHub repo secrets:
   - `NPM_TOKEN` (automation token with publish rights to `@triple0-labs` scope).
2. Ensure npm scope/package ownership is correct:
   - package: `@triple0-labs/pacto-spec`
3. Ensure release workflow is healthy:
   - `.github/workflows/release.yml`
4. Ensure npm workflow exists:
   - `.github/workflows/npm-publish.yml`

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
   - publishes to npm with provenance

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
