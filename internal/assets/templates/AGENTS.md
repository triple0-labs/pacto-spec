## Pacto Workflow

This file is an optional hand-off for AGENTS-compatible assistants.
Use Pacto workspace artifacts as the source of truth for planning and execution tracking:

- `./.pacto/plans/PACTO.md` (preferred when present)

When this repository is the Pacto tool itself, treat all plan state as product-internal work (not customer project work).

1. Run `pacto status` before implementation work.
2. Create new plan slices with `pacto new <state> <slug>`.
3. Use clear Pacto-internal slugs and titles (for example, `pacto-<area>-<change>`) to avoid cross-project ambiguity.
4. Keep plan status, blockers, and evidence up to date in plan documents.
5. Use `pacto status --format json` for machine-readable checks.
