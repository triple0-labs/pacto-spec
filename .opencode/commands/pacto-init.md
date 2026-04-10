<!-- pacto:managed:start -->
<!-- pacto:managed:meta artifact=opencode/command/pacto-init.md workflow=init contract=v1 template_sha256=e2faf55fcd205ab3a14e1fbc819b5361fbd4ec6690f2ddc2dedc81af33df3bc1 generated_by=pacto generated_at=2026-04-08T06:15:23Z -->
# pacto-init

Agent contract for init.

## Objective

Initialize a project-local Pacto workspace.

## Input Contract

### Required Inputs
- User intent answers captured through a short interview (problem, users, goals, constraints, success criteria).

### Optional Inputs
- `--root <path>` to initialize a specific project root.
- `--with-agents` to manage optional `AGENTS.md` hand-off block.
- `--force` to overwrite init-managed files.
- Optional product context details (non-goals, risks, rollout notes) to enrich `prd.md`.

## Execution Contract

- Tool target: opencode
- Recommended command: pacto init

## Output Contract
- Creates canonical state directories and template files under `.pacto/plans`.
- Reports created/updated/skipped items.
- `PACTO.md` under plans root remains canonical; `AGENTS.md` is optional hand-off only.
- Optionally creates or updates managed Pacto block in `AGENTS.md`.
- Creates or updates project-root `prd.md` with a basic PRD skeleton based on user interview answers.

## Validation Checklist
- Confirm `.pacto/plans/{current,to-implement,done,outdated}` exist.
- Confirm core docs (`README.md`, `PACTO.md`) exist.
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
