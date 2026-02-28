---
name: pacto-update
description: Agent contract for the Pacto update workflow to refresh previously installed managed artifacts.
---

<!-- pacto:managed:start -->
# Pacto Update Skill

Use this skill as an agent contract for the update workflow in Pacto projects.

## Objective

Refresh previously installed managed Pacto artifacts.

## When To Use

Use after upgrading Pacto to refresh managed blocks in generated artifacts.

## Input Contract

### Required Inputs
- Previously installed tool artifacts, or explicit `--tools` target.

### Optional Inputs
- `--tools <all|none|csv>` for explicit tool selection.
- `--force` to overwrite unmanaged files when needed.

## Execution Contract

- Tool target: codex
- Recommended command: pacto update [--tools <all|none|csv>] [--force]

## Output Contract
- Updates managed blocks in skill and command artifacts in place.
- Reports created/updated/skipped/failed counts.
- Preserves unmanaged files unless `--force` is set.

## Validation Checklist
- Confirm managed marker replacement happened for existing files.
- Review skipped unmanaged warnings and decide if force is appropriate.
- Spot-check one skill and one command artifact for expected template updates.

## Failure Modes and Handling
- Unsupported or invalid tool selection.
- Unmanaged files skipped without `--force`.
- Filesystem write errors.

## Implementation Status

- Status: **Implemented**
- Fallback: Use `--force` only when intentional overwrite of unmanaged files is acceptable.
<!-- pacto:managed:end -->
