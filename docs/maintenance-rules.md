# Maintenance Rules

## Goal

Keep docs aligned with runtime behavior so contributors and automation agents can treat docs as current contracts.

## Canonical Entry

- `docs/index.md` is the docs entrypoint.
- New technical docs should be linked from `docs/index.md`.
- `README.md` should link to docs instead of duplicating deep technical detail.

## Change-Impact Matrix

When these code areas change, update these docs in the same PR:

- API routes, query parsing, response payloads, pagination/error defaults:
  - `docs/api-contracts.md`
  - `docs/domain-model-rules.md` when payload semantics change
- Timeline hydration or entity timeline builders:
  - `docs/domain-model-rules.md`
  - `docs/frontend-behavior.md` if rendering expectations change
- Ingestion parsing, alias behavior, source precedence, sync lifecycle:
  - `docs/ingestion-parser-rules.md`
  - `docs/runtime-overview.md`
  - `docs/domain-model-rules.md` if payload structure changes
- Web route behavior, fetch/cache/error handling, global layout/nav/components:
  - `docs/frontend-behavior.md`
- Compose/env/runtime scripts/cron/backup behavior:
  - `docs/ops-and-scripts.md`
  - `docs/development.md` when local workflow changes
  - `README.md` when operator-facing steps change
- Fixture pipeline (`scripts/generate_patch_fixture.mjs`, `scripts/patch_fixture/*`):
  - `docs/ops-and-scripts.md`
  - `README.md` if usage expectations change
- Repo topology or subsystem responsibilities:
  - `docs/runtime-overview.md`
  - `docs/architecture.md`
  - `docs/index.md` if navigation changes

## PR Checklist

Before merge, verify:

1. `docs/index.md` links remain correct.
2. API behavior changes are reflected in `docs/api-contracts.md`.
3. Parsing/hydration/timeline changes are reflected in rules docs.
4. Env vars, script behavior, ports, and defaults are reflected in ops/development docs.
5. `README.md` remains consistent with canonical docs.

## Authoring Rules

- Describe current behavior only.
- Prefer behavior contracts and defaults over implementation narration.
- Include fallback behavior and failure-path semantics when they affect outputs.
- Keep docs consistent with tests where tests encode runtime contracts.
