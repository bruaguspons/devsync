# devsync

A single CLI that keeps a local dev environment in sync with two
sources of truth declared in this repo:

- **Tools** — versions declared in the repo's `mise.toml`, applied via
  `mise use`/`mise uninstall` (backend: `providers/mise`).
- **Skills** — Claude Code skill directories declared under the repo's
  top-level `skills/` folder, installed into `~/.claude/skills`
  (backend: `providers/skills`).

Both resource kinds are diffed and presented in one combined
select/confirm screen, distinguished only by a `[tool]`/`[skill]`
label prefix — there is no separate command per resource kind.

## Usage

```bash
devsync            # diff, select, confirm, apply
devsync -version   # print version and exit
```

If local state already matches desired state for every provider, no
selection screen is shown and the command reports "Everything is
already in sync."

## Architecture

```
main.go           # composition root: wires providers, calls core.Run
core/             # provider-agnostic domain layer (no mise/skills imports)
  resource.go       # Resource, DiffItem, ApplyResult
  provider.go        # Provider interface (Kind/LocalState/DesiredState/Apply)
  diff.go             # generic Diff() algorithm
  ui.go                # huh select/confirm UI
  lock.go               # lock file schema + reconciliation
  run.go                 # orchestration
providers/
  mise/             # mise-managed tools backend
  skills/           # repo skills/ folder backend
```

`core/` never imports a concrete provider package; the dependency
direction is `main.go -> providers/* -> core`, never the reverse.

## Skills sync

- **Desired state**: the repo's top-level `skills/<name>/` directories,
  bundled into the release archive as `skills.snapshot/` (mirroring how
  `mise.toml` ships as `mise.toml.snapshot`) and installed to
  `~/.local/share/devsync/skills.snapshot` by `install.sh`.
- **Local state**: whatever exists under `~/.claude/skills/<name>/`,
  regardless of how it got there.
- **Diffing**: SHA-256 over the full directory tree (relative path +
  content, sorted, NUL-separated), not semver — a rename-only change is
  detected as a hash change.
- **Applying**: install/update overwrites `~/.claude/skills/<name>` with
  the repo's copy (remove-then-copy); remove deletes only that named
  subdirectory. Both operations only ever touch the single named skill
  directory, never its parent.

`devsync` intentionally does **not** read or write
`.atl/skill-registry.md` or the gitignored `skills-lock.json` used by
gentle-ai's own skill-registry mechanism — a skill installed by that
mechanism will show up as an "update" or "removed" candidate the next
time `devsync` runs, and applying it will overwrite/delete those files
without reconciling ownership. This is expected, surfaced (never
silent — the user selects and confirms before anything is deleted),
accepted-risk behavior, not a bug.

## Lock file

`devsync` writes `~/.local/share/devsync/lock.json` after every
successful apply, recording each resource's kind/name/version. It is a
**cache, not a trust source**: `LocalState()` is always queried fresh
on every run (`mise ls --json` for tools, an on-disk hash walk for
skills) and always wins the diff. The lock is only ever used to skip
recomputing a skill's SHA-256 hash when that skill directory's on-disk
mtime hasn't changed since the last successful apply — a pure
performance optimization that can, at worst, cause one extra or
avoided hash recompute, never a wrong reported state.

This lock file is entirely separate from (different path, different
schema, never read/written together with) the gitignored
`skills-lock.json` used by gentle-ai's skill-registry mechanism.

## Testing

```bash
go test ./...
bash install_test.sh
```
