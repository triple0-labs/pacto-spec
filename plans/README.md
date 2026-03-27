# Plans Workspace

All previous archived and active plan slices were removed from this workspace.

## Current State

Plan status is derived from plan folders and documents via `pacto status`.
This README is informational only.

## Temporary Workspace

Generate a fresh workspace in a temporary directory when testing Pacto commands:

```bash
TMP_DIR="$(mktemp -d /tmp/pacto-demo-XXXXXX)"
pacto init --root "$TMP_DIR" --no-interactive --no-install
pacto status --root "$TMP_DIR/.pacto/plans" --repo-root .
pacto new to-implement demo --root "$TMP_DIR/.pacto/plans"
```
