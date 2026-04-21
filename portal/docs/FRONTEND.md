# Portal Frontend

React 18 SPA under `portal/src/`. Built with Vite, served from `portal/dist/`.

## Stack

- React 18 + React Router v5 (`react-router-dom`)
- Apollo Client — two instances (Portal API, Admin API)
- FluentUI v8 (`@fluentui/react`) — dominant component library
- Tailwind CSS v3 + CSS Modules (`*.module.css`) — styling
- react-intl — i18n (`FormattedMessage`, `renderToString`)
- Monaco Editor — config / template editor
- Chart.js — analytics widgets
- Storybook 9 — component preview (see `portal/docs/storybook.md`)
- GraphQL Codegen — typed queries/mutations

## Build

```sh
cd portal
npm run start        # Vite dev server with HMR
npm run build        # production build → portal/dist/
npm run typecheck    # tsc --noEmit
npm run gentype      # regenerate GraphQL types
npm run storybook    # Storybook on :6006
```

## Source layout

```
portal/src/
  index.tsx                 # Entry: icon init, ChartJS, Monaco, render <ReactApp/>
  ReactApp.tsx              # Top-level providers + router
  AppRoot.tsx               # Per-tenant routes; Admin API Apollo client
  ScreenLayout.tsx          # Chrome: sidebar + content
  ScreenNav.tsx             # Sidebar nav
  system-config.ts          # SystemConfig types + defaults + instantiate
  context/
    SystemConfigContext.ts  # useSystemConfig()
    AppContext.ts           # current app (tenant) context
  graphql/
    portal/                 # Portal API — schema, queries, mutations, screens
    adminapi/               # Admin API — schema, queries, mutations, screens
  components/
    v2/                     # New design-system components (co-located stories)
    applications/
    audit-log/
    auth/
    billing/
    common/
    design/
    ipblocklist/
    onboarding/
    project-wizard/
    roles-and-groups/
    saml/
    sms-provider/
    users/
  hook/                     # Reusable hooks (useCopyFeedback, ...)
  util/                     # Pure helpers
  locale-data/              # i18n JSON per locale
```

v1 components often live at the root of `src/` (e.g. `TextField.tsx`, `TextFieldWithCopyButton.tsx`, `FormTextField.tsx`). v2 components live under `src/components/v2/<Name>/` with their own folder.

## Providers (top-down)

Mounted in `ReactApp.tsx` / `AppRoot.tsx`:

1. `SystemConfigContext.Provider` — system config + FluentUI themes.
2. `IntlProvider` (via `AppLocaleProvider`) — react-intl messages for the active locale.
3. FluentUI `ThemeProvider` — colours/typography from system config.
4. Portal `ApolloProvider` — Portal API client.
5. Admin API `ApolloProvider` — created inside `AppRoot`, scoped to the tenant in the URL.
6. React Router (`BrowserRouter`, `Routes`).

Anything that calls `useSystemConfig()`, `useIntl()`, `useQuery()`, `useMutation()`, or `useParams()` depends on these. Storybook reproduces them selectively — see `portal/docs/storybook.md`.

## Routing

React Router v5. Top-level routes in `ReactApp.tsx`; per-app routes in `AppRoot.tsx` under `/project/:appID`. Screens are `lazy()`-loaded, wrapped in `Suspense` + an error boundary (`ErrorBoundSuspense` / `FlavoredErrorBoundSuspense`).

Use `useParams<{ appID: string }>()` to get the active tenant inside a screen.

## GraphQL pattern

1. Author the query/mutation in a `.graphql` file under `src/graphql/<portal|adminapi>/query|mutations/`.
2. Run `npm run gentype` → produces `*.generated.ts` with a typed hook.
3. Import the hook in the screen:

```tsx
const { data, loading, error } = useAppAndSecretConfigQuery({
  variables: { id: appID },
});
```

4. For mutations, refetch affected queries (`refetchQueries: [{ query: ..., variables: ... }]`) or write to the Apollo cache directly.
5. Errors surface through the app-wide error renderer (`ErrorRenderer.tsx`) — wrap with `FormContainer` / `FormErrorMessageBar` as appropriate.

**Do not hand-edit** `*.generated.ts`.

## Styling conventions

- Prefer FluentUI primitives (`Text`, `PrimaryButton`, `TextField`, `Dropdown`, `Dialog`, `DetailsList`, …) for v1 screens. See the `update-portal-ui` skill for pitfalls (e.g. `Text` + inline links, link components).
- For v2 screens, use components from `src/components/v2/`.
- Tailwind (v3) handles one-off utilities (`flex`, spacing, widths, colours) — no JIT config tricks required.
- CSS Modules (`Foo.module.css`) for component-scoped styles; imported as `styles.xxx`.
- Do **not** mix CSS Modules with `!important` to override FluentUI — extend the component's theme or pass `styles` props.

## i18n conventions

- All user-facing strings go through react-intl.
- In JSX: `<FormattedMessage id="Some.key" values={{ ... }} />`.
- For string props (`label`, `placeholder`, `ariaLabel`): `renderToString("Some.key")` from our intl wrapper.
- Keys land in `src/locale-data/en.json`; other locales are updated via the translation workflow — do not edit them manually.
- Inline links in translations must use the `ExternalLink` or `<a>` render props pattern — see the `update-portal-ui` skill.

## Forms

- `FormContainer` / `FormContainerBase` — standard form shell with save bar, unsaved-changes blocker, error surfacing.
- `useFormField` / form state hooks — immer-backed drafts against the original GraphQL data.
- `BlockerDialog` — unsaved-changes prompt on navigation.
- Validation surfaces through `FormErrorMessageBar` / `FormErrorMessageText`.

## Common utilities

- `useCopyFeedback` — copy-to-clipboard with FluentUI callout feedback; used by `TextFieldWithCopyButton`.
- `ExternalLink` — `<a target="_blank" rel="noopener">` with the correct icon.
- `ErrorBoundSuspense` / `FlavoredErrorBoundSuspense` — error + suspense wrapping for lazy routes.
- `useSystemConfig()` — throws if not under `SystemConfigContext.Provider`; Storybook stories for v1 components must provide this (see `portal/docs/storybook.md`).

## Authentication

The Portal uses Authgear itself for admin login. The backend proxies GraphQL requests, attaching the admin session. The frontend mostly treats auth as "401 → redirect"; `UnauthenticatedDialogContext` handles session-expiry UX.

## Performance practices

- Route components are `lazy()`-loaded.
- Heavy third-party deps (Monaco, Chart.js) are wired in `index.tsx` and tree-shaken where possible.
- Apollo cache policies avoid refetching tenant-scoped data when navigating between screens of the same tenant.
- Immer `setAutoFreeze(false)` is set globally in `index.tsx` to work around a prod-only bug in repeated `produce` calls.

## When editing a Portal screen

1. Read `CLAUDE.md` / `AGENTS.md` and `portal/docs/storybook.md` if you're adding stories.
2. Invoke the `update-portal-ui` skill — it covers link rendering, `Text` pitfalls, and i18n patterns that are easy to get wrong.
3. If you change a `.graphql` file, run `npm run gentype`.
4. After the change: `npm run typecheck` at minimum; `npm run build` for anything non-trivial.
