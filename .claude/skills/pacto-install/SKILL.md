---
name: pacto-install
description: Agent contract for the Pacto install workflow.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=claude/skill/pacto-install workflow=install contract=v1 template_sha256=e2468149e4d7709ef17bea3f5672869f0a31c79e47554fa9056edb284119f28f generated_by=pacto generated_at=2026-04-30T10:35:53Z -->
# Pacto Install Skill

Use this skill as an agent contract for the install workflow in Pacto projects.

## Objective

Install managed Pacto Agent Skills for supported tools.

## When To Use

Use to bootstrap Pacto-generated skills for compatible AI tools.

## Input Contract

### Required Inputs
- Either detectable tool directories or explicit `--tools` selection.

### Optional Inputs
- `--tools <all|none|csv>` for explicit selection.
- `--force` to overwrite unmanaged existing files.

## Execution Contract

- Tool target: claude
- Recommended command: pacto install [--tools <all|none|csv>] [--force]

## Output Contract
- Generates managed skill files per workflow and selected tool.
- Returns per-file outcome summary: created, updated, skipped, failed.

## Validation Checklist
- Confirm selected/detected tools match user intent.
- Check warnings for unmanaged file skips.
- Confirm generated skills are wrapped with managed markers.

## Failure Modes and Handling
- No tools detected when `--tools` is omitted.
- Invalid `--tools` argument values.
- Filesystem write failures for target tool paths.

## Implementation Status

- Status: **Implemented**
- Fallback: If detection fails, rerun with explicit `--tools` list.
<!-- pacto:managed:end -->
