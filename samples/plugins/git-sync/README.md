# git-sync plugin (sample)

This sample plugin syncs Git when running `pacto status`.

## Behavior

- Always tries `git fetch --prune <remote>` based on current branch upstream.
- Optional pull: set `PULL_ON_STATUS=1` to run `git pull --ff-only <remote> <branch>`.
- Default is fail-open (`STRICT_MODE=0`): sync failures warn but do not block `pacto status`.
- Set `STRICT_MODE=1` to block on sync failure.

## Install in a project

From your project root:

```bash
mkdir -p .pacto/plugins/git-sync
cp -R /path/to/pacto/samples/plugins/git-sync/* .pacto/plugins/git-sync/
chmod +x .pacto/plugins/git-sync/scripts/sync-status.sh
cp .pacto/plugins/git-sync/config.env.example .pacto/plugins/git-sync/config.env
```

Then validate and enable:

```bash
pacto plugin validate
pacto plugin enable git-sync
```

## Upstream setup

The plugin expects the current branch to have an upstream:

```bash
git branch --set-upstream-to=origin/<branch> <branch>
```

## Bypass for one run

```bash
pacto --allow-guardrail git-sync/status-sync status
```
