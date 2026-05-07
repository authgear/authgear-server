# Storybook

Storybook is used to develop and preview Portal UI components in isolation.

- Run locally: `npm run storybook` (inside `portal/`).
- Config: `portal/.storybook/main.ts`, `portal/.storybook/preview.tsx`.
- Stories are discovered via the glob `../src/**/*.stories.@(js|jsx|mjs|ts|tsx)`.

## Sidebar grouping convention

Stories are grouped by generation in the sidebar:

- `components/v2/<ComponentName>` — new components under `portal/src/components/v2/`. Titles are auto-derived from the file path, no explicit `title` is needed.
- `components/v1/<ComponentName>` — legacy components that live elsewhere under `portal/src/` (often at the root, or under subfolders like `components/users/`). These need an explicit `title` in `meta` to land in the right group.

Example v1 story:

```tsx
const meta = {
  title: "components/v1/TextFieldWithCopyButton",
  component: TextFieldWithCopyButton,
  // ...
} satisfies Meta<typeof TextFieldWithCopyButton>;
```

Nested subgroups are allowed via `/`, e.g. `components/v1/forms/TextFieldWithCopyButton`.

## File placement

Co-locate `<Component>.stories.tsx` next to `<Component>.tsx`. Do not move v1 components into a new folder just to get auto-titling — set `title` in `meta` instead.

## Global setup in `preview.tsx`

Every story is wrapped with:

- `AppLocaleProvider` — provides `react-intl` messages. Required by anything using `FormattedMessage` / `Context` from `src/intl`.
- `ThemeProvider` (v2) — sets CSS variables for the v2 design tokens.
- `initializeIcons()` from `@fluentui/react` — registers the FluentUI icon set. Without this, `IconButton` / `Icon` render blank.

## Providers needed by v1 components

v1 components often depend on contexts that are not globally provided. Add them as **component-level decorators** in the story file, not globally, to keep v2 stories clean.

Common ones:

- `SystemConfigContext` — many v1 components call `useSystemConfig()` and throw if no value is provided. Use `instantiateSystemConfig(defaultSystemConfig)` from `src/system-config.ts` to get a fully-populated value.
- Apollo `MockedProvider` — required if the component runs GraphQL queries/mutations. Supply `mocks` matching the query shape.
- React Router (`MemoryRouter`) — required if the component uses `useParams`, `Link`, `useNavigate`, etc.

Example decorator for a v1 component using `SystemConfigContext`:

```tsx
const systemConfig = instantiateSystemConfig(defaultSystemConfig);

const meta = {
  title: "components/v1/TextFieldWithCopyButton",
  component: TextFieldWithCopyButton,
  decorators: [
    (Story) => (
      <SystemConfigContext.Provider value={systemConfig}>
        <div style={{ width: 480 }}>
          <Story />
        </div>
      </SystemConfigContext.Provider>
    ),
  ],
  // ...
} satisfies Meta<typeof TextFieldWithCopyButton>;
```

If a provider ends up being needed by most v1 stories, promote it to `.storybook/preview.tsx` instead of repeating it.

## Story authoring conventions

- Use CSF3 (`Meta` / `StoryObj`) — match the style in `src/components/v2/Badge/Badge.stories.tsx`.
- Put shared defaults in `meta.args`; vary per-story via `args`.
- Add `tags: ["autodocs"]` to generate a docs page.
- For form-like or width-sensitive components, wrap the story in a fixed-width container — `layout: "centered"` otherwise collapses them.
- Cover the meaningful branches: default, empty, disabled, and any boolean props that toggle rendering (e.g. `hideCopyButton`).

## Checklist when adding a story for a v1 component

1. Create `<Component>.stories.tsx` next to the component.
2. Set `title: "components/v1/<Component>"` in `meta`.
3. Add decorators for any contexts the component requires (`SystemConfigContext`, `MockedProvider`, `MemoryRouter`).
4. Wrap in a sized container if the component is width-sensitive.
5. Add stories for each meaningful variant, not just the default.
