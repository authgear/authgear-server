# Storybook

Storybook is used to develop and preview Portal UI components in isolation.

- Run locally: `npm run storybook` (inside `portal/`).
- Config: `portal/.storybook/main.ts`, `portal/.storybook/preview.tsx`.
- Stories are discovered via the glob `../src/**/*.stories.@(js|jsx|mjs|ts|tsx)`.

## Sidebar grouping convention

- `components/v2/<ComponentName>` — reusable components in `portal/src/components/v2/<Name>/`. Titles are auto-derived from the file path. Do not set `title`.
- `portal/<Screen>/<Piece>` — a screen or a section of a screen that is not a reusable component. Set `title` in `meta`. Examples: `portal/LoginMethods/LoginMethodChooser`, `portal/Users/AccountStatusDialog`, `portal/CreateOAuthClient/AuthMethodChoice`.
- `components/<Name>` or `components/<area>/<Name>` — a shared component that is not under `components/v2/`. Set `title` only when the file path would not produce that name. `CheckboxWithTooltip.stories.tsx` lives at `portal/src/` and sets `title`. `AppearanceSwitcher` and `UnauthenticatedDialog` auto-title from `components/header/` and `components/auth/`.

Example shared component that does not live under `components/v2/`:

```tsx
const meta = {
  title: "components/CheckboxWithTooltip",
  component: CheckboxWithTooltip,
  // ...
} satisfies Meta<typeof CheckboxWithTooltip>;
```

Nested subgroups are allowed via `/`, e.g. `portal/LoginMethods/LoginMethodIcon`.

## File placement

Co-locate `<Component>.stories.tsx` next to `<Component>.tsx`. Do not set `title` on a story under `portal/src/components/v2/`. A screen story that lives next to its screen sets `title: "portal/<Screen>/<Piece>"` instead of moving the screen into `components/v2/`.

## Global setup in `preview.tsx`

Every story is wrapped with:

- `AppLocaleProvider` — provides `react-intl` messages. Required by anything using `FormattedMessage` / `Context` from `src/intl`.
- `ThemeProvider` (v2) — Radix `Theme` from `src/components/v2/ThemeProvider/ThemeProvider.tsx`.
- An **Appearance** toolbar (Light / Dark / Device) — calls `setAppearance()` from `src/util/appearance.ts`, the same path the portal uses, so the `light-theme`/`dark-theme` class lands on `<html>` and `useAppearance()` consumers update.

## Providers a story may still need

Some screens depend on contexts that are not globally provided. Add them as **component-level decorators** in the story file, not globally, so design-system stories stay free of them.

Common ones:

- `SystemConfigContext` — components that call `useSystemConfig()` throw if no value is provided. This is system config, not a theme. Use `instantiateSystemConfig(defaultSystemConfig)` from `src/system-config.ts` to get a fully-populated value.
- Apollo `MockedProvider` — required if the component runs GraphQL queries/mutations. Supply `mocks` matching the query shape.
- React Router (`MemoryRouter`) — required if the component uses `useParams`, `Link`, `useNavigate`, etc.

Example decorator for a screen story that calls `useSystemConfig()`:

```tsx
const systemConfig = instantiateSystemConfig(defaultSystemConfig);

const meta = {
  title: "portal/LoginMethods/LoginMethodChooser",
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
} satisfies Meta;
```

If a provider ends up being needed by most screen stories, promote it to `.storybook/preview.tsx` instead of repeating it.

## Story authoring conventions

- Use CSF3 (`Meta` / `StoryObj`) — match the style in `src/components/v2/Badge/Badge.stories.tsx`.
- Put shared defaults in `meta.args`; vary per-story via `args`.
- Add `tags: ["autodocs"]` to generate a docs page.
- For form-like or width-sensitive components, wrap the story in a fixed-width container — `layout: "centered"` otherwise collapses them.
- Cover the meaningful branches: default, empty, disabled, and any boolean props that toggle rendering.

## Checklist when adding a reusable component story

1. Add `<Component>.stories.tsx` next to the component under `portal/src/components/v2/<Name>/`.
2. Do not set `title`.
3. Use CSF3. Put shared defaults in `meta.args`.
4. Add `tags: ["autodocs"]`.
5. Wrap width-sensitive components in a fixed-width container.

## Checklist when adding a screen story

1. Co-locate `<Piece>.stories.tsx` next to the screen file.
2. Set `title: "portal/<Screen>/<Piece>"`.
3. Add decorators only for contexts that screen needs (`MemoryRouter`, and `SystemConfigContext` when the screen calls `useSystemConfig()`).
4. Wrap in a sized container if the component is width-sensitive.
5. Add stories for each meaningful variant, not just the default.
