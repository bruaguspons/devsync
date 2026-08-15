# Structure & Architecture Contract (shared)

Normative rules shared by `file-placement` (preventive) and `structure-migration` (remediation).
Per-stack directory layouts live in `structure-layouts-by-stack.md`.

## 1. Detection Order

Detect before deciding. Never assume.

| Step | Evidence to read |
|------|------------------|
| Language / build | `pom.xml`, `build.gradle*`, `package.json`, `go.mod`, `*.csproj`, `pyproject.toml`, `Cargo.toml`, `composer.json` |
| Source roots | build config, then `src/main/<lang>`, `src/`, `app/`, `internal/`, `cmd/` |
| Architecture | signals table below, over the 2–3 deepest populated trees |
| Existing convention | the dominant real pattern, not the aspirational one in docs |

### Architecture signals

| Signal found in tree | Architecture |
|---|---|
| `domain/` + `application/` (or `usecase/`) + `infrastructure/`, or `ports`/`adapters` | Hexagonal / Ports & Adapters |
| Per-context folders each with `domain/{entity,valueobject,repository,event}` + `application` | DDD by bounded context |
| Top-level `controller/`, `service/`, `repository/`, `dto/`, `model/` | Layered (package-by-layer) |
| Top-level `features/` or `modules/` each self-contained | Feature-sliced / Screaming |
| `entities/ features/ widgets/ pages/ shared/` | Feature-Sliced Design (frontend) |
| `atoms/ molecules/ organisms/` | Atomic Design (UI only, never business layout) |
| No consistent signal | Report `unknown`; propose feature-sliced default, do not silently impose it |

**Hard rule:** the detected architecture wins over personal preference. Improve inside it; propose a change of paradigm only as an explicit, separate decision the user approves.

## 2. Placement Rules

1. **Feature first, layer second.** Group by feature / bounded context / aggregate, then by layer *inside* it. Never place feature folders inside layer folders.
2. **Artifacts live next to their owner.** A DTO, mapper, validator, or exception belongs to the feature (and endpoint) it serves — never in a global bucket.
3. **Banned as global buckets:** `dto/`, `dtos/`, `models/`, `utils/`, `helpers/`, `services/`, `managers/`, `impl/`, `common/`, `misc/`, `lib/` at or near the source root. Allowed only as a leaf inside a feature, or as `shared/` with a real cross-feature consumer count ≥ 2.
4. **Dependencies point inward.** Domain imports no framework, no persistence, no transport. Application imports domain + ports only. Infrastructure and transport may import inward, never the reverse.
5. **Tests mirror the source path** in the stack's idiomatic test root.

## 3. Numeric Thresholds

Measure; do not eyeball. Counts exclude generated code and barrel/index files.

**Directory thresholds — binding.** A directory is a defect when it breaches these AND a real cohesion axis exists to split along (entity, aggregate, use case, subdomain). Never split a directory to satisfy a number alone; if no axis exists, report it and leave it.

| Directory metric | Warn | Defect |
|---|---|---|
| Source files in one leaf directory | > 8 | > 12 |
| Distinct entities/aggregates served by one directory (once it has ≥ 4 files) | 2 | > 2 |
| Depth below source root | > 5 | > 7 |

**File metrics — signals only, never a split trigger.** They mean "inspect this file for mixed responsibilities" (§4). A large, deeply cohesive file that passes §4 is correct and stays as it is. Report the signal, state that cohesion held, move on.

| File signal | Inspect when |
|---|---|
| Lines in one file | > 400 |
| Public/exported types per file (domain & application) | > 2 |
| Public methods on one class | > 12 |
| Constructor dependencies | > 7 |

## 4. Responsibility Mixing — the only reason to split a file

This section, not size, decides splits. Split a file when any of these holds:

- It touches more than one of: transport (HTTP/CLI/queue/UI event), business rule, persistence/SQL, mapping/serialization, validation, cross-cutting (auth, cache, logging, config).
- It serves more than one entity/aggregate or more than one unrelated use case.
- Its imports span more than two layers, or a domain file imports framework/infrastructure.
- Name contains `And`, or a vague suffix (`Manager`, `Helper`, `Util`, `Processor`, `Handler` with no bounded subject).
- Two or more public entry points that no caller uses together, changing for different reasons.
- Test file asserting more than one unit's behavior.

Each split proposal names one target file per responsibility and states which caller/reason-to-change drives each. If you cannot name the distinct reason each half would change for, there is no split — say so.

**Cohesion overrides size.** Many small functions serving one responsibility, an exhaustive mapper, a parser table, a state machine, or a value object with rich behavior legitimately produce a long file. Leave it.

## 5. Framework Adaptation

The framework's own conventions are part of the target, not an obstacle. Adapt, never transplant a layout from another ecosystem.

| Step | Requirement |
|---|---|
| Identify framework + major version | From the manifest/lockfile, not from folder names |
| Load its layout conventions | `structure-layouts-by-stack.md`; if the stack is absent or the version's convention is uncertain, fetch current official docs (context7 preferred) before proposing anything |
| Respect its idioms | Naming case, file suffixes, module/registration mechanism, test root, resource/asset location, DI style |
| Respect its magic | Convention-over-configuration paths (route/file-based routing, component scan, autoload, migrations, generators) are constraints — moving those files breaks the app |
| Preserve its unit of modularity | NestJS module, Django app, Rails engine, Go package, .NET project, Angular module, Spring `@Configuration` scope |

A rename that violates a framework naming contract is a bug, not a cleanup. When framework convention and this contract disagree, the framework wins and you state the constraint explicitly.

## 6. Non-Negotiables

- Size is never a reason to split. Mixed responsibilities are.
- Never move or split a file whose language, layer, or owning feature you could not determine — report it as `needs-input`.
- Never break the public API of a published package silently; list every changed import path.
- Framework-mandated locations win over these rules (Next.js `app/`, Rails, Django apps, Go `cmd/`, generated sources). Cite the constraint when it wins.
- Preserve behavior. Structure work is behavior-neutral; a real bug found on the way is a separate report, not an inline fix.
