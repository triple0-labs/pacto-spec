# Plans Workspace

All previous archived and active plan slices were removed from this workspace.

## Current State

Plan status is derived from plan folders and documents via `pacto status`.
This README is informational only.

## Mock Project

Use the mock project for testing Pacto commands:

- [samples/mock-pacto-repo](../samples/mock-pacto-repo/)

Example:

```bash
pacto status --root ./samples/mock-pacto-repo --repo-root .
pacto new to-implement demo --root ./samples/mock-pacto-repo
```
