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
.pacto/
├── config.yaml
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

Canonical workflow contract is:

- `.pacto/plans/PACTO.md`

Optional hand-off for AGENTS-compatible tools:

```bash
pacto init --with-agents
```

Non-interactive/agent-friendly init:

```bash
pacto init --no-interactive --tools codex,cursor --yes
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
pacto status
```

For CI automation:

```bash
pacto status --format json --fail-on partial
```

## 4. Execute Planned Work

Use `pacto exec` to advance execution tasks and append execution evidence in plan docs.
(`exec` only runs for plans in `current` state.)

```bash
pacto exec current improve-auth-flow --note "Started implementation"
pacto exec current improve-auth-flow --step 1.1 --evidence src/auth/flow.go
```

## 5. Move Plan State Explicitly

Use `pacto move` for explicit workflow transitions.

```bash
pacto move to-implement improve-auth-flow current
pacto move current improve-auth-flow done --reason "All tasks complete and verified"
```

## 6. Explore Ideas (Optional)

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

Use explicit roots when needed:

```bash
pacto status --root . --repo-root .
pacto new to-implement my-plan --root .
```
