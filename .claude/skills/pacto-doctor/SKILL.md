---
name: pacto-doctor
description: Agent contract for the Pacto doctor workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=claude/skill/pacto-doctor workflow=doctor contract=v1 template_sha256=f54f3fa4baa01d1c804a2c72946e58807639b581307a18905660fa56ad86bd4d generated_by=pacto generated_at=2026-04-30T10:35:53Z -->
# Pacto Doctor Skill

Use this skill as an agent contract for the doctor workflow in Pacto projects.

## Objective

Audit managed integration artifacts for drift and legacy patterns.

## When To Use

Use when generated skills might be stale after manual edits or CLI/template upgrades.

## Input Contract

### Required Inputs
- None when tool auto-detection succeeds.

### Optional Inputs
- `--root <path>` to target a specific project root.
- `--tools <all|none|csv>` for explicit tool selection.
- `--format table|json` for human vs automation output.
- `--fail-on none|drift|legacy|any` for CI enforcement.

## Execution Contract

- Tool target: claude
- Recommended command: pacto doctor [--root <path>] [--tools <all|none|csv>] [--format table|json] [--fail-on none|drift|legacy|any]

## Output Contract
- Reports managed artifact status (`ok|missing|unmanaged|legacy_managed|stale|meta_mismatch|legacy_pattern`).
- Provides recommended remediation action (typically `pacto update --artifacts`).
- Emits summary counters for drift and legacy findings.

## Validation Checklist
- Confirm analyzed root and tool selection match user intent.
- Review drift/legacy findings before remediation.
- If CI use case, set explicit `--fail-on` policy.

## Failure Modes and Handling
- No tools detected when `--tools` is omitted.
- Invalid tool list or fail-on/format flag values.
- Filesystem read errors while auditing artifacts.

## Implementation Status

- Status: **Implemented**
- Fallback: If findings are expected but unresolved, run `pacto update --artifacts` and re-run `pacto doctor`.
<!-- pacto:managed:end -->
