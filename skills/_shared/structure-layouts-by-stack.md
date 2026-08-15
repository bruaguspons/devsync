# Target Layouts by Stack (shared reference)

Canonical destinations used by `file-placement` and `structure-migration`.
Rules and thresholds: `structure-architecture-contract.md`.
`<feature>` = feature / bounded context / aggregate (e.g. `billing`, `booking`), never a layer name.

## Java / Kotlin — Spring Boot (hexagonal or DDD)

```
src/main/java/<base>/<feature>/
├── domain/          Invoice.java  InvoiceId.java  InvoiceStatus.java  InvoicePolicy.java
├── application/     IssueInvoiceUseCase.java
│   └── port/        out/InvoiceRepository.java  out/PaymentGateway.java
├── infrastructure/
│   ├── persistence/ JpaInvoiceRepository.java  InvoiceJpaEntity.java  InvoiceEntityMapper.java
│   └── payment/     StripePaymentGateway.java
└── api/
    ├── IssueInvoiceController.java
    ├── dto/         IssueInvoiceRequest.java  InvoiceResponse.java
    └── mapper/      InvoiceApiMapper.java
```

- One controller per use case or per tight endpoint group; never one controller per entity holding 12 endpoints.
- Request/Response DTOs are **per endpoint**, inside that feature's `api/dto/`. A repo-wide `dto/` package is the primary defect this contract exists to remove.
- JPA entities are infrastructure, never domain. Keep a mapper between them.
- Cross-feature code goes to `<base>/shared/` (or `common/config`) only with ≥ 2 consuming features.
- Layered legacy repos: keep `controller/service/repository`, but subdivide each by feature (`controller/billing/…`) instead of flattening.
- Tests: `src/test/java/<base>/<feature>/…` mirroring the path.

## TypeScript — React / frontend

```
src/
├── features/<feature>/
│   ├── components/   InvoiceList.tsx  InvoiceRow.tsx
│   ├── hooks/        useInvoices.ts
│   ├── api/          invoices.client.ts  invoices.schema.ts
│   ├── model/        invoice.types.ts  invoice.mapper.ts
│   └── index.ts      public surface of the feature
└── shared/
    ├── ui/           design-system primitives (atomic design allowed here only)
    ├── lib/          framework-agnostic helpers, one file per concern
    └── config/
```

- One component per file; container/presentational split when a component both fetches and renders.
- No global `types/`, `utils/`, or `services/` barrel; co-locate with the feature.
- Cross-feature import goes through `features/<x>/index.ts`, never deep paths.

## Node — NestJS / Express

`src/modules/<feature>/{domain,application,infrastructure,http}` — NestJS module per feature; DTOs under `http/dto/`.

## Go

`internal/<feature>/{domain,app,adapter}`; entrypoints in `cmd/<binary>/`. Package name = feature, so avoid stutter (`billing.Invoice`). Interfaces are declared by the consumer package.

## .NET

`src/<Product>.<Feature>.Domain|Application|Infrastructure|Api` (project per layer) or `src/<Product>/Features/<Feature>/{Domain,Application,Infrastructure,Api}` for a single project. DTOs under the feature's `Api/Contracts/`.

## Python

`src/<package>/<feature>/{domain,application,infrastructure,api}.py` or package folders when a module passes the size threshold. Pydantic schemas under the feature's `api/schemas/`, not a global `schemas.py`.

## Stack not listed here

Do not transplant the Java shape. Derive the target instead:

1. Identify framework + major version from the manifest/lockfile.
2. Fetch its current official structure conventions (context7 preferred) — layout, naming case, file suffixes, module/registration unit, test root.
3. Keep every convention-driven path exactly where the framework expects it.
4. Apply only the paradigm-independent rules: group by feature, put layers inside the feature, keep DTOs/schemas next to the endpoint they serve, point dependencies inward.
5. State the derived layout and its source before proposing moves.

## Framework-mandated exceptions

Next.js `app/`, Rails `app/*`, Django app layout, Laravel, Android/iOS module conventions, Angular/NestJS module registration, migrations, autoloaded paths, and generated sources keep their prescribed location. Place only *business* code by feature, and state the constraint when it overrides a rule.

Idioms to preserve per ecosystem: file-name casing (`PascalCase.tsx`, `snake_case.py`, `kebab-case.service.ts`), test roots (`src/test/java`, `__tests__`, `*_test.go`, `tests/`), and the framework's own unit of modularity.
