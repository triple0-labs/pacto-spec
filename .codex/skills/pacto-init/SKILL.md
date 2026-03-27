---
name: pacto-init
description: Agent contract for the Pacto init workflow to scaffold project-local Pacto workspace files.
---

<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=codex/skill/pacto-init workflow=init contract=v1 template_sha256=54b527088279673041d42b64130021e33816f124708016229ec9c7875bfbb855 generated_by=pacto generated_at=2026-03-27T13:34:57Z -->
# Pacto Init Skill

Use this skill as an agent contract for the init workflow in Pacto projects.

## Objective

Initialize a project-local Pacto workspace.

## When To Use

Use once per project to create canonical `.pacto/plans` workspace scaffolding.

## Input Contract

### Required Inputs
- User intent answers captured through a short interview (problem, users, goals, constraints, success criteria).

### Optional Inputs
- `--root <path>` to initialize a specific project root.
- `--with-agents` to manage optional `AGENTS.md` hand-off block.
- `--force` to overwrite init-managed files.
- Optional product context details (non-goals, risks, rollout notes) to enrich `prd.md`.

## Execution Contract

- Tool target: codex
- Recommended command: pacto init

## Output Contract
- Creates canonical state directories and template files under `.pacto/plans`.
- Reports created/updated/skipped items.
- `PACTO.md` under plans root remains canonical; `AGENTS.md` is optional hand-off only.
- Optionally creates or updates managed Pacto block in `AGENTS.md`.
- Creates or updates project-root `prd.md` with a basic PRD skeleton based on user interview answers.

## Validation Checklist
- Confirm `.pacto/plans/{current,to-implement,done,outdated}` exist.
- Confirm core docs (`README.md`, `PACTO.md`, template, slash commands) exist.
- When `--with-agents`, confirm managed block markers are present in `AGENTS.md`.
- Confirm the agent asked clarifying intent questions before drafting `prd.md`.
- Confirm `prd.md` includes at least: problem, goals, non-goals, users, scope, success metrics, open questions.

## Failure Modes and Handling
- State path already exists as non-directory.
- Filesystem permission errors while creating workspace.
- Force-less init skips existing managed files by design.
- Insufficient product intent from user to draft a useful PRD.

## Implementation Status

- Status: **Implemented**
- Fallback: If intent is unclear, ask follow-up questions and create `prd.md` with explicit assumptions/open questions before proceeding.
<!-- pacto:managed:end -->
