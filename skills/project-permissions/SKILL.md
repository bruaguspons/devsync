---
name: project-permissions
description: "Trigger: configurar permisos, settings.local.json, permisos del proyecto, project permissions, allow everything except remote git. Configure a project so Claude Code may do anything except affect the remote repository."
license: Apache-2.0
metadata:
  author: "brunopons"
  version: "1.0"
---

## Activation Contract

Load this skill when the user asks to configure, bootstrap, repair, or audit `.claude/settings.local.json` permissions for a project, or asks to let Claude Code "do everything except touch the remote repo".

Do not load it for global `~/.claude/settings.json` changes, hook authoring, or MCP configuration.

## Hard Rules

- Write only to `<project>/.claude/settings.local.json`. Never modify `.claude/settings.json` or `~/.claude/settings.json`.
- Read the existing file first. Merge arrays; never drop pre-existing `allow`, `deny`, `ask`, `env`, or `hooks` entries.
- The deny list from `assets/settings.local.template.json` is a floor, not a ceiling. Never remove a deny entry, even on user request in the same turn — report the conflict and stop.
- `git commit` stays allowed only because `/commit-changes` needs it. `git commit --amend` stays denied.
- Never enable `defaultMode: "bypassPermissions"`. Always keep `disableBypassPermissionsMode: "disable"`.
- State once, verbatim, that deny rules are prefix matchers: dynamically built commands (`$G push`, `sh -c '…'`, wrapper scripts) are NOT covered. Do not overstate the guarantee.
- Validate with `jq` after writing. A malformed file silently disables every setting in it.

## Decision Gates

| Situation | Action |
|-----------|--------|
| No `.claude/settings.local.json` | Create it from the template as-is |
| File exists | Merge template `deny`/`ask` into it; keep its `allow` plus template `allow` |
| Project is not a git repo | Ask the user to confirm applying permissions anyway before proceeding (deny rules are inert but survive `git init`); do not apply silently. Proceed only on explicit confirmation; stop without writing on decline. |
| `.claude/settings.local.json` not gitignored | Add it to `.gitignore` |
| User wants a deny entry removed | Refuse, name the entry, explain the remote-repo risk |
| User rejects the proposed plan | Ask what to adjust, apply the requested change, re-present the full revised plan; repeat until approved or the user stops. Never write until explicit approval. |

## Execution Steps

1. Resolve the project root (`git rev-parse --show-toplevel`, else cwd). If no `.git` is found, ask the user to confirm applying permissions anyway before proceeding; on confirmation continue and note deny rules are currently inert, on decline stop without writing anything.
2. Read `assets/settings.local.template.json` and the existing `.claude/settings.local.json` if present.
3. Build the merged result in memory, deduplicating rule strings.
4. Present the proposed plan and ask for approval. On rejection, ask what to adjust, apply the requested change to the merged result, and re-present the full revised plan; repeat until approved or the user stops. Never write until explicit approval.
5. On approval, write the result.
6. Ensure `.claude/settings.local.json` is gitignored.
7. Validate: `jq -e '.permissions.deny | length > 0' .claude/settings.local.json`.
8. Smoke-check three representative denials against the written rules: `git push`, `git commit --amend`, `gh pr create`.
9. Tell the user the file is written and that `/hooks` or a restart may be needed for a running session to reload settings.

## Output Contract

Return:
- Absolute path written, and created-vs-merged.
- Count of `allow` / `deny` / `ask` rules.
- The prefix-matcher limitation, stated once.
- Any `.gitignore` change.
- Any deny-removal request refused.

## References

- `assets/settings.local.template.json` — the permission baseline to merge.
- `references/deny-rationale.md` — what each deny group protects and known gaps.
