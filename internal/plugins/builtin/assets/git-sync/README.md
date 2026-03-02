# git-sync plugin (built-in sample)

This plugin syncs Git when running `pacto status`.

- Always runs `git fetch --prune <remote>` using current branch upstream.
- Optional pull via `PULL_ON_STATUS=1` in `config.env`.
- Default fail-open (`STRICT_MODE=0`): sync failures warn and allow `pacto status`.
- Strict mode (`STRICT_MODE=1`) blocks on sync failure.
