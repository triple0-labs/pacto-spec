# Contributing

## Local Development

From the repository root:

```bash
go test ./...
```

Optional: the repo ships a [local dev skill](../.agents/skills/pacto-dev-local/SKILL.md) with build/install shortcuts for day-to-day work on the CLI.

Build binaries locally:

```bash
go build -o pacto ./cmd/pacto
```

## Tiny Smoke

Run minimal end-to-end smoke against release artifact:

```bash
make tiny-smoke
```

This verifies a minimal flow in `/tmp/pacto-tiny-smoke/mock`:

```text
status -> new -> status
```

## Documentation

User-facing docs live under [`docs/`](./README.md). When you change CLI behavior or defaults, update [Commands](./commands.md) (and [Integrations](./integrations.md) or [Plugins](./plugins.md) when artifacts or guardrails change), and add a note to [CHANGELOG.md](../CHANGELOG.md).

## Release process

Use the canonical release checklist in:

- [RELEASING.md](../RELEASING.md)

Tag flow:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```
