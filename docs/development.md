# Development Workflow

## Prerequisites
- Node.js 20+
- npm
- Go 1.22+

## Run API
```bash
cd api
go mod tidy
go run ./cmd/server
```

Default API URL: `http://localhost:8080`.

## Run Web
```bash
cd web
npm install
npm run dev
```

Default web URL: `http://localhost:3000`.

## Quality Checks
Frontend:
```bash
cd web
npm run lint
npm run test
npm run build
```

Backend:
```bash
cd api
go test ./...
```

Source guidance check:
```bash
node scripts/check_source_limits.mjs
```

## Generate Fixture + Assets
```bash
node scripts/generate_patch_fixture.mjs
```

Notes:
- This command requires network access.
- It updates fixture JSON and mirrored assets for the configured patch slug.
