# ProcureFlow Web

Vite, React, TypeScript, and Tailwind frontend for ProcureFlow.

## Local Development

```bash
npm install
npm run dev
```

Create `web/.env` when the API is not available at the default URL:

```env
VITE_API_BASE_URL=http://localhost:8080
```

## Structure

```text
src/
├── app/                 # providers and router
├── features/            # business state by domain
├── routes/              # route-level pages
├── shared/api/          # API wrapper and generated OpenAPI types
├── shared/auth/         # session persistence
├── shared/components/   # layout and reusable UI
├── shared/hooks/
├── shared/lib/
├── shared/types/
└── styles/
```

## API Contract

The backend OpenAPI file is the source of truth:

```text
../internal/interfaces/http/apidocs/openapi.yaml
```

Generate TypeScript types with:

```bash
npm run generate:api
```

The app currently includes a small handwritten API wrapper in `src/shared/api/client.ts` so early screens can work before generated clients are adopted throughout the UI.
