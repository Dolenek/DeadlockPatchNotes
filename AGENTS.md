# AGENTS.md

## Purpose
This file defines working rules for contributors and coding agents in this repository.
The project is a monorepo with:
- `api/`: Go HTTP API
- `web/`: Next.js frontend
- `scripts/`: fixture generation and asset mirroring scripts

## Scope
Rules in this file apply to hand-written source and docs:
- `api/**/*.go`
- `web/**/*.{ts,tsx,css}`
- `scripts/**/*.{js,mjs}`
- `README.md`, `docs/**/*.md`, `api/sql/**/*.sql`

These rules do not apply to generated or vendor-style artifacts:
- `web/node_modules/**`
- `web/.next/**`
- `web/public/assets/**` (mirrored patch assets)
- `api/internal/patches/data/*.json` (fixtures)
- `web/package-lock.json`

## Code Size Guidance
- Prefer source files under 400 lines.
- Avoid source files above 500 lines unless there is a strong reason.
- Treat functions above ~40 lines as refactor candidates.
- For unavoidable orchestrators/parsers, split logic into helpers/modules and keep the top-level flow easy to read.

## Readability and Naming
- Use descriptive, intent-revealing names.
- Avoid vague names such as `data`, `info`, `helper`, `temp` when a precise name is possible.
- Keep modules focused on one responsibility.

## Architecture Conventions
- Keep API handlers thin; business/data logic belongs in `api/internal/patches` and related packages.
- Keep UI components composable and presentation-focused.
- Keep data-generation scripts modular: parsing, lookup, asset download, and output assembly should be separated.

## Documentation Rules
- `docs/index.md` is the docs entrypoint.
- Canonical docs describe current behavior; avoid historical narratives in canonical pages.
- For substantial feature/refactor work, update docs when behavior, architecture, or workflow changes.
- Link docs from `README.md` instead of duplicating large content.

## Validation Before Merge
Run relevant checks for changed areas:
- Frontend: `cd web && npm run lint && npm run test && npm run build`
- Backend: `cd api && go test ./...`
- Data pipeline changes: `node scripts/generate_patch_fixture.mjs` (when network is available)
- Source size/function guidance report: `node scripts/check_source_limits.mjs`

## Practical Notes
- These are guidance rules, not hard CI gates.
- Prioritize maintainability and clarity while preserving behavior.
