# Getting Started

This guide covers first-run setup and the default Pacto SDD loop.

## Prerequisites

- `pacto` installed (see [README](../README.md#install)).
- A repository where plans should be tracked.

## 1. Initialize Workspace

```bash
pacto init
```

Default output location:

```text
./.pacto/plans/
```

Generated structure:

```text
.pacto/plans/
├── current/
├── to-implement/
├── done/
├── outdated/
├── README.md
├── PACTO.md
├── PLANTILLA_PACTO_PLAN.md
└── SLASH_COMMANDS.md
```

## 2. Create a Plan Slice

```bash
pacto new to-implement improve-auth-flow
```

Creates:

- `<plans-root>/to-implement/improve-auth-flow/README.md`
- `<plans-root>/to-implement/improve-auth-flow/PLAN_IMPROVE_AUTH_FLOW_<YYYY-MM-DD>.md`

Also updates `<plans-root>/README.md` counts and links.

## 3. Verify State and Evidence

```bash
pacto status --format table
```

For CI automation:

```bash
pacto status --format json --fail-on partial
```

## 4. Explore Ideas (Optional)

Use `pacto explore` for ideation before creating formal plan slices.

```bash
pacto explore auth-refresh --title "Auth refresh ideas"
pacto explore auth-refresh --note "Compare token vs session model"
pacto explore --list
```

## Root Auto-Discovery

`pacto status` and `pacto new` auto-discover plan roots from current directory and parents.

Resolution pattern:

- direct state-folder root (`current`, `to-implement`, `done`, `outdated`)
- `./.pacto/plans`
- `./plans`

Use explicit roots when needed:

```bash
pacto status --plans-root ./.pacto/plans --repo-root .
pacto new to-implement my-plan --root ./.pacto/plans
```
