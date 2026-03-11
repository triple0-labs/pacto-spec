Made by Pacto, owned by 000geid.

Before planning or implementation actions:

1. Run `pacto status` and confirm blockers/evidence are current.
2. Confirm your branch is synchronized with upstream (`fetch`, then `rebase`/`merge` as needed).
3. Resolve conflicts before editing files:
   - `git status`
   - resolve conflict markers
   - `git add <files>`
   - continue with `git rebase --continue` or finalize merge commit

PR and conflict recovery playbook:

- Check PR context: `gh pr status` and `gh pr view --web` (if `gh` is available).
- If branch is behind: `git pull --rebase` (or team-approved merge flow).
- If branch diverged:
  - `git fetch --prune`
  - `git rebase <upstream>` (or merge)
  - resolve conflicts, run tests, push with `--force-with-lease` when rebasing.
- If you need to bypass once for emergency work, document why and use:
  - `pacto --allow-guardrail git-guardrails/git-preflight <command ...>`
