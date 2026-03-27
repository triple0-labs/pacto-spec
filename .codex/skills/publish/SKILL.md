---
name: publish
description: Release workflow for publishing a new pacto version (tag + GitHub Release). Use when preparing or executing a versioned release.
---

# Publish Skill

## Objective

Ship a new `pacto` release safely with correct tag placement, CI status, and post-release verification.

## Input Contract

### Required Inputs
- Target version `X.Y.Z` (without `v`).

### Optional Inputs
- Scope of release notes/checks.
- Whether to perform local upgrade verification after publish.

## Execution Contract

- Tool target: codex
- Repo root: `/home/diego/sandbox/pacto`
- Canonical doc: `RELEASING.md`

## Workflow

1. Preflight
- `cd /home/diego/sandbox/pacto`
- `git status --short`
- `go test ./internal/integrations ./internal/app`

2. Create/repoint tag to correct commit
- `git tag -d v<X.Y.Z>` (only if local tag already exists)
- `git tag v<X.Y.Z>`
- Verify: `git show --no-patch --oneline v<X.Y.Z>`

3. Push
- `git push origin main`
- `git push origin v<X.Y.Z>`

4. CI + release checks
- `gh run list --repo triple0-labs/pacto-spec --limit 10`
- `gh api repos/triple0-labs/pacto-spec/releases/tags/v<X.Y.Z> --jq '{tag_name: .tag_name, draft: .draft, prerelease: .prerelease, published_at: .published_at, assets: (.assets|length)}'`

5. Post-release verification (optional but recommended)
- `pacto update --check`
- `pacto update --yes`
- `pacto version`

## Validation Checklist

- `v<X.Y.Z>` points to intended commit.
- `main` and tag pushed successfully.
- Release exists and has assets.
- Local update flow can see and install the release.

## Failure Modes and Handling

- Tag on wrong commit:
  - Recreate tag on correct commit, push tag again.
- Release not found (404 by tag):
  - Wait for `Release` workflow completion and retry.
- `pacto update --check` still shows old version:
  - Confirm release is published and assets exist; retry after a short delay.
