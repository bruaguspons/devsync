# Deny Rationale and Known Gaps

## How the rules are evaluated

- `deny` beats `allow` beats `ask`. A denied command is refused before any prompt appears.
- Bash rules are **prefix matchers** over the command string, not a shell parser.
- `"Bash"` with no parentheses allows every Bash invocation not caught by `deny` or `ask`.
- `disableBypassPermissionsMode: "disable"` prevents `--dangerously-skip-permissions` from erasing the deny list. Without it the whole baseline is bypassable in one flag.

## Deny groups

| Group | Protects against |
|-------|------------------|
| `git push`, `git remote *`, `git update-ref`, `git send-email`, `git svn dcommit` | Any write that leaves this checkout |
| `git commit --amend`, `rebase`, `reset --hard`, `filter-branch`, `filter-repo`, `reflog expire`, `gc --prune`, `stash drop/clear` | History rewrite and unrecoverable local loss |
| `git tag -a/-s/-m/-f/-d` | Creating or moving publishable tags. Bare `git tag` stays in `ask` so `git tag --list` remains usable |
| `git branch -d/-D/-m/-c/-f/-u/--set-upstream-to` | Branch destruction and upstream rewiring. Bare `git branch` stays in `ask` so listing works |
| `gh` / `glab` / `hub` write subcommands, `gh api -X`, `gh api --method` | PRs, issues, releases, secrets, workflow dispatch, repo mutation |
| `npm/pnpm/yarn/cargo/twine/gem/mvn publish`, `docker push`, `docker login` | Publishing artifacts to external registries |
| `sudo`, `git config --global/--system` | Escaping project scope and rewriting machine-wide git identity |
| `Read`/`Edit` on `.env*`, `.ssh`, `.aws/credentials`, `.netrc`, `.npmrc`, `.git-credentials`, `.git/**` | Secret exfiltration and hand-editing git internals to reach a remote |

`git commit` (without `--amend`) is deliberately allowed: `/commit-changes` is the single sanctioned write.

## Known gaps — do not oversell this baseline

Prefix matching is defeated by any command whose text does not literally start with the denied prefix:

- indirection — `G=git; $G push origin main`
- nested shells — `sh -c 'git push'`, `bash -lc "gh pr create"`
- wrapper scripts — `./scripts/deploy.sh`, `make release`, a `package.json` script that pushes
- aliases and git config — `git config alias.p push` then `git p`
- non-Bash tools — a `Write` to `.git/hooks/pre-commit`, though `Edit(./.git/**)` covers the common case

This baseline is a **guardrail against accident and drift**, not a sandbox against a determined actor. A `PreToolUse` hook inspecting the full command string is the stronger option if that threat model matters later.

Because the file lives at `.claude/settings.local.json` and is user-writable, anyone with filesystem access can loosen it. Treat it as policy, not enforcement.
