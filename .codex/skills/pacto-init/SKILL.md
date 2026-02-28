---
name: pacto-init
description: Agent contract for the Pacto init workflow to scaffold project-local Pacto workspace files.
---

<!-- pacto:managed:start -->
# Pacto Init Skill

Use this skill as an agent contract for the init workflow in Pacto projects.

## Objective

Initialize a project-local Pacto workspace.

## When To Use

Use once per project to create canonical `.pacto/plans` workspace scaffolding.

## Input Contract

### Required Inputs
- None for current directory initialization.

### Optional Inputs
- `--root <path>` to initialize a specific project root.
- `--with-agents` to manage `AGENTS.md` guidance block.
- `--force` to overwrite init-managed files.

## Execution Contract

- Tool target: codex
- Recommended command: pacto init

## Output Contract
- Creates canonical state directories and template files under `.pacto/plans`.
- Reports created/updated/skipped items.
- Optionally creates or updates managed Pacto block in `AGENTS.md`.

## Validation Checklist
- Confirm `.pacto/plans/{current,to-implement,done,outdated}` exist.
- Confirm core docs (`README.md`, `PACTO.md`, template, slash commands) exist.
- When `--with-agents`, confirm managed block markers are present in `AGENTS.md`.

## Failure Modes and Handling
- State path already exists as non-directory.
- Filesystem permission errors while creating workspace.
- Force-less init skips existing managed files by design.

## Implementation Status

- Status: **Implemented**
- Fallback: If files are skipped unexpectedly, rerun with `--force` only for init-managed artifacts.
<!-- pacto:managed:end -->
