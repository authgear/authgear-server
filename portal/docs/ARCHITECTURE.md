# Portal Architecture

The Portal is the tenant-admin web app for Authgear — tenant owners use it to configure their project (auth methods, OAuth clients, branding, users, roles, billing, etc.).

It is a React SPA that talks to two GraphQL endpoints:
- **Portal API** (`pkg/portal/graphql/`) — tenant/project metadata, billing, collaborators, domains, app configuration.
- **Admin API** (`pkg/admin/graphql/`) — per-tenant resources (users, sessions, identities, roles, groups, audit logs).

The Portal backend also brokers authentication against the Authgear server itself — the admin logs in with Authgear and receives an admin session used to call both GraphQL endpoints.

## Stack

**Frontend (`portal/src/`)**
- React 18 + React Router v5 (`react-router-dom`)
- Apollo Client — two clients, one per GraphQL endpoint (`makeClient` for Admin API, `usePortalClient` for Portal API)
- FluentUI v8 (`@fluentui/react`) — primary component library (buttons, dialogs, inputs, layout)
- Tailwind CSS v3 — utility styling, alongside CSS Modules (`*.module.css`)
- react-intl — i18n; messages under `portal/src/locale-data/`
- Monaco Editor — embedded code editor (templates, JSON config)
- Chart.js — analytics widgets
- Vite — build tool; outputs to `portal/dist/`
- GraphQL Codegen — typed operations from `.graphql` files into `*.generated.ts`
- Storybook 9 — component preview (see `portal/docs/storybook.md`)

**Backend (`cmd/portal/`, `pkg/portal/`)**
- Go HTTP server
- GraphQL handler (`pkg/portal/graphql/`)
- Delegates tenant-scoped data to the Admin API of the target Authgear server

## Key files and folders

**Frontend**
- `src/index.tsx` — entry: `initializeIcons`, ChartJS registration, Monaco worker wiring, immer `setAutoFreeze(false)`, render `<ReactApp />`
- `src/ReactApp.tsx` — top-level providers (System config, locale, Apollo, routing)
- `src/AppRoot.tsx` — routes for a specific app (tenant); wires Admin API Apollo client scoped to the app ID in the URL
- `src/ScreenLayout.tsx`, `src/ScreenNav.tsx` — chrome around each screen (sidebar, top bar)
- `src/context/SystemConfigContext.ts` — global system config; hydrated from `/api/system-config`
- `src/system-config.ts` — `PartialSystemConfig`, `defaultSystemConfig`, `instantiateSystemConfig`
- `src/graphql/portal/` — Portal API: schema, queries, mutations, screens (e.g. `AdminAPIConfigurationScreen.tsx`)
- `src/graphql/adminapi/` — Admin API: schema, queries, mutations, screens
- `src/components/` — shared UI by area (`applications`, `audit-log`, `users`, `v2`, …)
- `src/components/v2/` — new design-system components; own folder per component with `.stories.tsx`
- `src/locale-data/` — translation JSON per locale
- `src/hook/` — reusable hooks (e.g. `useCopyFeedback`)
- `src/util/` — pure helpers

**Backend**
- `cmd/portal/main.go` — entry point
- `pkg/portal/graphql/` — GraphQL schema and resolvers
- `pkg/portal/` — services (app config, collaborators, domains, billing, …)

## Build

```sh
cd portal
npm run start        # Vite dev server (proxies to backend)
npm run build        # production build → portal/dist/
npm run typecheck    # tsc --noEmit
npm run gentype      # regenerate *.generated.ts from .graphql
npm run storybook    # Storybook on :6006
```

## Routing

Top-level routes in `ReactApp.tsx` → per-app routes in `AppRoot.tsx`. Screens are `lazy()`-loaded to keep the initial bundle small. The active app ID comes from the URL (`/project/:appID/...`) and scopes the Admin API Apollo client.

## GraphQL data flow

1. A screen imports a generated hook, e.g. `useAppAndSecretConfigQuery` from `graphql/portal/query/appAndSecretConfigQuery.generated.ts`.
2. `.graphql` files next to the generated file are the source of truth; `npm run gentype` regenerates the TypeScript.
3. Two Apollo clients:
   - **Portal client** (`usePortalClient()`) — app-level metadata.
   - **Admin API client** (`makeClient(appID)`) — scoped to one tenant; created inside `AppRoot`.
4. Mutations typically use `useMutation(…Document)` + `refetchQueries`; some screens use optimistic updates via immer-backed form state.

## Configuration and theming

- System config (feature flags, branding, available languages, embedded Authgear config for admin login) loads once and lives in `SystemConfigContext`.
- FluentUI themes are computed from system config via `createTheme(...)` in `src/system-config.ts` and exposed as `themes.main`, `themes.inverted`, `themes.destructive`, `themes.actionButton`, `themes.verifyButton`, `themes.defaultButton`.
- v2 components use a separate `ThemeProvider` under `src/components/v2/ThemeProvider/` providing CSS variables for the new design system.

## i18n

- Messages live in `src/locale-data/<locale>.json`.
- Use `<FormattedMessage id="..." />` for rendered text, `renderToString("...")` (from our intl wrapper) for string props like `label`/`placeholder`.
- When translations contain inline links, follow the conventions in the `update-portal-ui` skill.

## Testing

- `npm test` — Jest unit tests.
- `@storybook/addon-vitest` — Vitest runs story interaction tests.
- E2E tests live at repo root under `e2e/tests/` (Playwright-driven YAML) — **not** inside `portal/`.

## Generated code

Never hand-edit:
- `src/graphql/**/*.generated.ts` — rerun `npm run gentype`.
- `src/locale-data/*.json` for non-English locales — translation workflow owns these.
