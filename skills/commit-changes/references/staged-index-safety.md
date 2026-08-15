# Staged Index Safety

Procedure for committing the staged index without losing staged state. Applies only after the user approved the plan.

## 1. Snapshot the index

```bash
git write-tree            # -> <SNAPSHOT_TREE>, records the exact staged state
git diff --cached --name-only > /dev/null   # keep the path list in context too
```

Keep `<SNAPSHOT_TREE>` for the whole operation. It restores partial (hunk-level) staging that `git add <path>` cannot rebuild.

## 2. Single-commit case

```bash
git commit -m "<subject>" -m "<body>"
```

No index manipulation. Preferred whenever the staged diff is one work unit.

## 3. Multi-commit split

Repeat per approved unit, in the planned order:

```bash
git read-tree <SNAPSHOT_TREE>          # index == original staged state
git reset --quiet                       # unstage everything (worktree untouched)
git add -- <unit paths>                 # re-stage only this unit's paths
git commit -m "<subject>" -m "<body>"
```

Simpler equivalent when every staged file matches its worktree content:

```bash
git reset --quiet
git add -- <unit-1 paths> && git commit -m "..."
git add -- <unit-2 paths> && git commit -m "..."
```

Constraints:

- A path may appear in exactly one unit.
- A path whose staged content differs from the worktree MUST NOT be split; `git add <path>` would pull in the unstaged edits. Keep it whole in one commit and disclose it in the plan.
- Never use `git add -A`, `git add .`, `git add -p`, or `git commit -a`.

## 4. Failure and restore

If any commit or hook fails:

```bash
git read-tree <SNAPSHOT_TREE>   # restore the original staged set (minus what already committed)
git status --porcelain
```

Then stop and report: commits that landed, the failing unit, and the verbatim hook/git error. Do not retry with `--no-verify`.

## 5. Post-commit verification

```bash
git log --oneline -<n>
git log -<n> --format='%an <%ae> | %cn <%ce>' # must be the system git identity only
git log -<n> --format='%b' | grep -Ei 'co-authored-by|generated with|claude|anthropic|assistant'
```

The grep MUST return nothing. If it matches, report it — do not amend to fix it; amending rewrites history and is out of scope.
