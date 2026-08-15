---
name: commit-changes
description: "Trigger: commit, commitear, create commit, commit staged changes, split staged changes. Propose commits from the staged index and commit only after explicit user approval."
license: Apache-2.0
metadata:
  author: "brunopons"
  version: "1.0"
---

## Activation Contract

Load this skill whenever the user asks to commit, commitear, or split already-staged work into commits.

Do not load it for pushing, PR creation, history rewrites, or release tagging.

## Hard Rules

- Commit ONLY what is already staged. Never run `git add`, `git add -A`, `git add -p`, `git commit -a`, or stage new paths that were not in the original index.
- NEVER push, amend, tag, rebase, reset `--hard`, or touch remotes. Stop after the local commit.
- NEVER add `Co-Authored-By`, `Generated with`, `Signed-off-by`, AI names, or any authorship trailer. The only author and committer is the system `git config user.name` / `user.email`.
- Abort with a message if `user.name` or `user.email` is unset — do not pass `--author` or `-c user.*` to invent one.
- NEVER use `--no-verify`. Hooks run as configured; report hook failures verbatim. This holds even if the user explicitly asks to skip hooks — refuse, explain hooks run as configured, and proceed without `--no-verify`.
- NEVER commit staged content matching a sensitive pattern (`.env`, `*.key`, `*.pem`, `*secret*`, `*credential*`, `*token*`) by filename or by scanning the staged diff content itself — flag the match and stop before proposing approval.
- Present the plan, then STOP. Commit only after the user explicitly approves. Silence, an unrelated reply, or an ambiguous answer is NOT approval.
- Commit messages are English Conventional Commits. Subject <=72 chars, imperative mood, no trailing period.
- On any failure mid-plan, stop immediately, restore the original index from the snapshot tree, and report which commits already landed.

## Decision Gates

| Situation | Action |
|-----------|--------|
| Nothing staged | Report it, list unstaged/untracked paths, ask the user to stage. Do not stage for them. |
| Staged diff is one work unit | Propose a single commit. |
| Staged diff mixes unrelated work units | Propose an ordered split, one commit per work unit (see `work-unit-commits`). |
| A file is staged AND has different unstaged content | Do not split that file across commits; keep its staged version in one commit and say so. |
| Detached HEAD, active merge/rebase, or unresolved conflicts | Report the state and stop. |
| On the default branch (`main`/`master`) | Warn once in the plan; proceed only if the user approves anyway. |
| User approves part of the plan | Commit only the approved units; leave the rest staged. |
| A hook modified files | Stop, report the modified paths, and ask before re-staging anything. |
| Staged content matches sensitive pattern (`.env`, `*.key`, `*.pem`, `*secret*`, `*credential*`, `*token*`) | Flag the match, stop before proposing approval. |
| User provides their own commit message | Validate against Conventional Commit format rules (type from the table below, subject <=72 chars, imperative mood, no trailing period, no authorship trailer); reject or ask for correction on violation, do not commit as-is. |

## Execution Steps

1. Read state: `git rev-parse --abbrev-ref HEAD`, `git status --porcelain`, `git diff --cached --stat`, `git diff --cached`, `git log --oneline -10` (to match existing message style).
2. Verify the identity and repo gates from Hard Rules and Decision Gates.
3. Scan the staged content (`git diff --cached`) for sensitive material, not just filenames: `.env`, `*.key`, `*.pem`, `*secret*`, `*credential*`, `*token*`. Flag any match and stop before proposing approval.
4. Group the staged diff into work units and draft a Conventional Commit message per unit, selecting a type from the Conventional Commit Types table below. If the user supplies their own message, validate it against the same format rules before use.
5. Present the plan using the Output Contract, then STOP and wait.
6. On approval, follow `references/staged-index-safety.md`: snapshot the index tree, commit each unit in order, restore on failure.
7. Verify with `git log --oneline -<n>` and `git status --porcelain`, then confirm no push happened and no authorship trailer exists (`git log -1 --format='%an <%ae>%n%b'`).

## Conventional Commit Types

| Type | Use when |
|------|----------|
| `feat` | New feature or user-visible behavior |
| `fix` | Bug fix |
| `refactor` | Restructure without behavior change |
| `test` | Add or fix tests |
| `docs` | Documentation only |
| `chore` | Tooling, deps, config — no behavior change |
| `style` | Formatting, whitespace, no logic change |
| `perf` | Performance improvement |
| `ci` | CI/CD pipeline changes |
| `build` | Build system or external dependency changes |

## Output Contract

Before committing, return the plan:

- Branch, staged file count, and any warning from the Decision Gates.
- For each proposed commit: order number, full message (subject + body), and its file list.
- Anything left staged on purpose.
- The literal line: `Approve to commit? Nothing is committed until you say so.`

Approval examples: "yes", "ok", "approve", "lgtm", "dale", or an equivalent unambiguous affirmative count as approval. Silence, an unrelated reply, or an ambiguous answer (e.g. "looks fine but check X", a topic change, no reply) is NOT approval — treat it per the Hard Rule above and wait or ask for clarification.

After committing, return: short hash + subject per commit, files still staged or unstaged, and an explicit `No push performed.`

## References

- `references/staged-index-safety.md` — index snapshot/restore and multi-commit split procedure.
- `../work-unit-commits/SKILL.md` — how to group a diff into reviewable work units.
