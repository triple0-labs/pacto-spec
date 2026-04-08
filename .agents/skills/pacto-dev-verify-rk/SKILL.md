---
name: pacto-dev-verify-rk
description: Audit Pacto usage and configuration in the saas-platform-rk project, including skills, prompts, and workspace contracts. Use when checking alignment/drift without reviewing plan content itself.
---

# Pacto Dev Verify RK Skill

## Objective

Assess whether `saas-platform-rk` is aligned with current Pacto workflow artifacts and contracts, while avoiding judgments about plan business content.

## Scope

- Target project: `/home/diego/work/alt94/saas-platform-rk`
- Focus: `.agents`, `.cursor`, `.pacto/plans` contract files and generated artifacts
- Out of scope: plan implementation quality/content decisions inside individual plan slices

## Execution Contract

- Tool target: codex
- Main verification command:
  - `cd /home/diego/sandbox/pacto && go run ./cmd/pacto doctor --root /home/diego/work/alt94/saas-platform-rk --format table`

## Audit Workflow

1. Discover artifact surfaces
- List:
  - `/home/diego/work/alt94/saas-platform-rk/.agents/skills`
  - `/home/diego/work/alt94/saas-platform-rk/.cursor/skills`
  - `/home/diego/work/alt94/saas-platform-rk/.cursor/commands`
  - `/home/diego/work/alt94/saas-platform-rk/.pacto/plans`

2. Run authoritative drift audit
- Use `pacto doctor` and capture summary counts: `ok|drift|legacy|missing|unmanaged|errors`.

3. Validate Codex skill format quickly
- Check each `SKILL.md` has YAML frontmatter at top (`--- ... ---`).

4. Spot-check contract drift patterns
- Look for stale phrases in prompts/skills such as:
  - `PLAN_<TOPIC>_<YYYY-MM-DD>`
  - root index-update claims for `pacto new`/`pacto move`
  - deprecated `--plans-root`
  - legacy wrapper files like `pacto-plan.md`

5. Compare workspace contracts (optional)
- Diff project `.pacto/plans/PACTO.md` and `SLASH_COMMANDS.md` against current templates to classify intentional customization vs stale behavior.

## Output Contract

- Return a concise report with:
  - current doctor summary
  - concrete misalignment files
  - whether drift is intentional customization or stale generated artifact
  - minimal remediation options (`manual patch` vs `update --artifacts`)

## Validation Checklist

- Paths inspected are under `saas-platform-rk`.
- `pacto doctor` run completed without command error.
- Findings include exact files, not generic statements.
- Recommendation preserves user preference about plan ordering/content.

## Failure Modes and Handling

- `pacto doctor` unavailable:
  - run from `/home/diego/sandbox/pacto` via `go run ./cmd/pacto doctor ...`.
- False positives from intentional custom docs:
  - mark as intentional and do not auto-remediate.
- Regeneration would overwrite custom prompt policy:
  - recommend upstream template change before `update --artifacts`.
