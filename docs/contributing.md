# Contributing

## Local Development

```bash
go test ./...
```

Build binaries locally:

```bash
go build -o pacto ./cmd/pacto
go build -o pacto-engine ./cmd/pacto-engine
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

## Release Process

Use the canonical release checklist in:

- [RELEASING.md](../RELEASING.md)

Tag flow:

```bash
git tag vX.Y.Z
git push origin vX.Y.Z
```
