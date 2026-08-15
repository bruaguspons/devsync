---
name: file-placement
description: "Trigger: creating a new file, where should this go, new endpoint/DTO/component/service, adding a class, splitting a file. Place new code by feature and keep one responsibility per file."
license: Apache-2.0
metadata:
  author: "brunopons"
  version: "1.0"
---

## Activation Contract

Load BEFORE writing any new source file, and before adding a second responsibility to an existing one. Applies when: creating a class, module, component, DTO/schema, endpoint, use case, mapper, test, or config; when the natural destination is a bucket folder (`dto/`, `utils/`, `services/`, `models/`, `common/`).

Skip only for single-line edits inside an already-correct file.

## Hard Rules

- Read `../_shared/structure-architecture-contract.md` and `../_shared/structure-layouts-by-stack.md` before choosing a path. Its detection order, placement rules, thresholds, and smells are binding.
- Detect stack, framework major version, and existing architecture from the repo first. Match the dominant real convention and the framework's own idioms (naming case, suffixes, module unit, test root); never introduce a second competing layout. When the framework's convention is uncertain, fetch its current official docs (context7 preferred) before choosing a path.
- One responsibility per file. If the planned file would serve transport + business rule + persistence + mapping, or more than one entity/use case, create separate files instead.
- Never split by size. A long file is fine when it serves one responsibility; only mixed responsibilities justify more files. Do not pre-split a cohesive unit to keep files small.
- Never add a file to a directory already at defect density — create the feature/entity subfolder in the same change.
- DTOs, schemas, mappers, validators and exceptions go inside the owning feature (and endpoint), never a repo-wide bucket.
- State the chosen path and its justification in one line before writing. Do not ask permission for a path the contract determines.

## Decision Gates

| Situation | Action |
|---|---|
| Feature already exists | Place inside it, in the matching layer folder |
| New feature / bounded context | Create the feature folder with only the layers this change needs |
| Target directory over threshold | Split it by feature/entity now, then place |
| Genuinely cross-feature code, ≥ 2 consumers | `shared/` (or stack equivalent), one file per concern |
| Cross-feature code, 1 consumer | Keep it inside that consumer |
| Framework dictates location or naming | Obey the framework; state the constraint |
| One responsibility, large file | Keep it in one file; size alone never splits |
| Layer or owning feature undeterminable | Ask one targeted question; do not guess |

## Execution Steps

1. Detect language, framework + version, source root, and architecture per the shared contract.
2. Name the feature / bounded context this code belongs to.
3. Name each responsibility in the planned change, with the reason each would change. One file per responsibility — no more, no fewer.
4. Resolve target paths from `structure-layouts-by-stack.md` (or derive them from the framework's official conventions); check directory density and depth.
5. Create missing directories, then write the files with their idiomatic test siblings, using the framework's naming idioms.
6. If the change revealed an existing violation you did not fix, report it as a follow-up; hand large cleanups to `structure-migration`.

## Output Contract

Return:
- Detected stack + architecture (with the evidence file).
- Feature/bounded context chosen.
- One line per created file: `path — responsibility`.
- Any threshold split performed as part of the placement.
- Follow-up violations observed but out of scope.

## References

- `../_shared/structure-architecture-contract.md` — detection order, placement rules, thresholds, responsibility smells.
- `../_shared/structure-layouts-by-stack.md` — canonical target layouts per stack.
- `../structure-migration/SKILL.md` — remediation of existing structure.
