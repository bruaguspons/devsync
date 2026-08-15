---
name: structure-migration
description: "Trigger: reorganize repo structure, folder with too many files, god file, split responsibilities, architecture audit, reorganizar carpetas. Audit and migrate an existing tree to feature-first clean structure."
license: Apache-2.0
metadata:
  author: "brunopons"
  version: "1.0"
---

## Activation Contract

Load when asked to audit or reorganize an existing tree: bloated bucket folders (`dto/`, `utils/`, `services/`), directories mixing several entities, files mixing responsibilities, inconsistent architecture, or "clean up the structure of this project". For placing a *new* file, use `file-placement` instead.

## Hard Rules

- Read `../_shared/structure-architecture-contract.md` and `../_shared/structure-layouts-by-stack.md` first; its detection order, placement rules, thresholds, and mixing criteria are binding.
- Detect framework + major version and adapt to its conventions, naming idioms, and convention-driven paths. When uncertain, fetch its current official docs (context7 preferred) before proposing any move. Never transplant another ecosystem's layout.
- Split files only for mixed responsibilities, never for size. Report large cohesive files as inspected-and-correct; do not propose moving them.
- Behavior-neutral only. Moves, renames, splits, and import updates — no logic changes, no API redesign, no dependency upgrades. Bugs found become a separate report.
- Produce and present the plan BEFORE touching any file. Execute only after explicit user approval.
- Migrate in independently reviewable batches, one feature or one violation class per batch. Never one sweeping commit.
- Use `git mv` when the repo is git-tracked, so history follows. Never push, tag, or commit outside `commit-changes`.
- After each batch: update every import/reference, then run the project's build/typecheck and tests. A red batch is reverted or fixed before the next batch starts.
- Never move a file whose language, layer, or owning feature is undetermined — list it as `needs-input`.
- Report every measurement with real numbers (file counts, LOC, entity counts). No impressionistic claims.

## Decision Gates

| Finding | Remediation |
|---|---|
| Bucket folder over density threshold | Redistribute into the owning features; delete the bucket |
| Directory mixing > 2 entities | Split into one subfolder per entity/aggregate |
| Over-threshold directory with no real cohesion axis | Leave; report why no split applies |
| Feature folders nested inside layer folders | Invert: feature first, layer inside |
| Large file, one responsibility | Leave untouched; report as cohesive |
| File mixing responsibilities | Split one file per responsibility, naming the distinct reason each changes |
| Framework naming/convention conflict | Framework wins; state the constraint |
| Domain importing framework/persistence | Extract a port; move the adapter to infrastructure |
| Architecture detected `unknown` | Present 2 candidate target layouts with tradeoffs; user picks before any move |
| Framework-mandated location | Leave in place; state the constraint |

## Execution Steps

1. Detect stack, framework + version, source roots, and architecture; record the evidence and the framework's binding conventions.
2. Directory inventory: per directory, count source files and distinct entities/aggregates served. Flag against the directory thresholds.
3. File inspection: use the file signals only to pick candidates, then read each candidate and judge it against the responsibility-mixing criteria. Split proposals come from that judgment, never from the signal.
4. Classify findings by violation class and rank by blast radius (references in / references out).
5. Emit the migration plan: batches, and per batch `from → to` moves, splits with target names and the reason each target changes, expected reference updates, and verification command.
6. Stop and request approval.
7. Execute batch by batch: move/split, update references, run build + tests, report the batch result. Halt on failure.
8. After the last batch, re-run the inventory and report residual violations and deliberate exceptions.

## Output Contract

Return:
- Detected stack, framework + version, architecture (with evidence), and source roots.
- Findings table: path, violation class, evidence. Directory findings cite the count; file findings cite the mixed responsibilities, not the line count.
- Files inspected and kept: path + the single responsibility that justifies their size.
- Batched migration plan with explicit `from → to` and split targets.
- Per executed batch: moves applied, references updated, build/test result.
- `needs-input` items, framework exceptions, and residual violations.

## References

- `../_shared/structure-architecture-contract.md` — detection order, placement rules, thresholds, responsibility smells.
- `../_shared/structure-layouts-by-stack.md` — canonical target layouts per stack.
- `../file-placement/SKILL.md` — preventive placement for new files.
