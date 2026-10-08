# Portal design-system docs and Storybook titles

## Goal / scope

Make portal docs and Storybook match the current UI.

The portal component library is Radix Themes (`@radix-ui/themes`) plus the wrappers in `portal/src/components/v2/`. `@fluentui/react` is not a dependency. `components/v1` exists only as a Storybook `title` on five stories. It is not a source folder.

This plan changes documentation and five Storybook `meta.title` strings. It does not move `portal/src/components/v2/`, does not change component behavior, and does not change `portal/src/util/shades.ts`.

## Out of scope

- `portal/src/util/shades.ts`, `portal/src/util/theme.ts` `deriveColors` / `getShades`, and `portal/src/util/__fixtures__/fluent-shades-golden.json`. `deriveColors()` is called from `portal/src/graphql/portal/DesignScreen/form.ts` and `portal/src/screens/v2/ProjectWizard/form.ts` to compute a brand button's hover/active color. `ColorPicker` calls `parseCSSColor`. The golden fixtures exist so those colors stay byte-identical. Do not delete, rewrite, or rename that algorithm in this work.
- Dead-code removal of `ThemePreviewWidget`, `addLightTheme`, or `addDarkTheme`. Those look unused outside tests. That is a separate change, because `getShades` shares the golden-fixture contract with `deriveColors`.
- Moving or renaming `portal/src/components/v2/`. Storybook's `components/v2/...` group is the auto-title of that folder. Keep it.
- Rewriting portal screens to use different components.

## Compatibility

No config, API, storage, or persisted theme data changes. A Storybook title change makes Chromatic treat the old story as removed and the new title as added. Accept the new baselines. No portal runtime behavior changes.

## Storybook title changes

Edit only the `title` string in `meta`. Leave decorators, args, and render functions as they are.

| File | Current `title` | New `title` |
|---|---|---|
| `portal/src/graphql/portal/LoginMethodChooser.stories.tsx` | `components/v1/LoginMethodChooser` | `portal/LoginMethods/LoginMethodChooser` |
| `portal/src/graphql/portal/LoginMethodIcon.stories.tsx` | `components/v1/LoginMethodChooser/LoginMethodIcon` | `portal/LoginMethods/LoginMethodIcon` |
| `portal/src/graphql/portal/IconRadioCards.stories.tsx` | `components/v1/IconRadioCards` | `portal/LoginMethods/AuthenticationSection` |
| `portal/src/CheckboxWithTooltip.stories.tsx` | `components/v1/Checkbox/CheckboxWithTooltip` | `components/CheckboxWithTooltip` |
| `portal/src/graphql/adminapi/AccountStatusDialog.stories.tsx` | `components/v1/Dialog/AccountStatusDialog` | `portal/Users/AccountStatusDialog` |

`IconRadioCards.stories.tsx` renders `LoginMethodAuthenticationSection`, which imports `portal/src/components/v2/IconRadioCards/IconRadioCards.tsx`. The new title must not be `components/v2/IconRadioCards`, because that name is already the auto-title of `portal/src/components/v2/IconRadioCards/IconRadioCards.stories.tsx`.

Do not add a `title` to stories under `portal/src/components/v2/`. They stay auto-titled from the file path.

Leave these titles unchanged:

- `portal/src/graphql/portal/StarterKitSection.stories.tsx` — `portal/StarterKitSection`
- `portal/src/graphql/portal/CreateOAuthClientScreen/AuthMethodChoice.stories.tsx` — `portal/CreateOAuthClient/AuthMethodChoice`
- `portal/src/graphql/portal/CreateOAuthClientScreen/FrameworkGrid.stories.tsx` — `portal/CreateOAuthClient/FrameworkGrid`
- `portal/src/components/header/AppearanceSwitcher.stories.tsx` — `components/header/AppearanceSwitcher`
- `portal/src/components/auth/UnauthenticatedDialog.stories.tsx` — no explicit `title` (auto-titled from `components/auth/`)

## `portal/docs/storybook.md`

Replace the sidebar-grouping section. The convention becomes:

- `components/v2/<ComponentName>` — reusable components in `portal/src/components/v2/<Name>/`. No explicit `title`. Auto-derived from the path.
- `portal/<Screen>/<Piece>` — a screen or a section of a screen that is not a reusable component. Set `title` in `meta`. Examples: `portal/LoginMethods/LoginMethodChooser`, `portal/Users/AccountStatusDialog`, `portal/CreateOAuthClient/AuthMethodChoice`.
- `components/<Name>` or `components/<area>/<Name>` — a shared component that is not under `components/v2/`. Set `title` only when the path would not produce that name. Example: `CheckboxWithTooltip.stories.tsx` at `portal/src/` sets `title: "components/CheckboxWithTooltip"`. `AppearanceSwitcher` and `UnauthenticatedDialog` already auto-title from their folders.

Delete every `components/v1/...` example, including both `TextFieldWithCopyButton` snippets. `TextFieldWithCopyButton.tsx` does not exist. Replace the v1 example with the `CheckboxWithTooltip` title above.

Replace the "Checklist when adding a story for a v1 component" section with two checklists:

1. New reusable component: add `<Component>.stories.tsx` next to the component under `portal/src/components/v2/<Name>/`. Do not set `title`. Use CSF3. Put shared defaults in `meta.args`. Add `tags: ["autodocs"]`. Wrap width-sensitive components in a fixed-width container.
2. New screen story: co-locate `<Piece>.stories.tsx` next to the screen file. Set `title: "portal/<Screen>/<Piece>"`. Add decorators only for contexts that screen needs (`MemoryRouter`, and `SystemConfigContext` when the screen calls `useSystemConfig()`).

Keep the existing notes that are still true:

- Global preview wrappers in `portal/.storybook/preview.tsx`: `AppLocaleProvider`, v2 `ThemeProvider`, Appearance toolbar.
- `SystemConfigContext` is opt-in per story, via `instantiateSystemConfig(defaultSystemConfig)` from `portal/src/system-config.ts`. It is system config, not a theme.
- Apollo `MockedProvider` and `MemoryRouter` stay per-story when the component calls those hooks.

Delete the sentence that says v1 components live at the root or under `components/users/` and need a `components/v1` title.

## `portal/docs/FRONTEND.md`

### Stack

Replace the FluentUI v8 bullet with:

- Radix Themes (`@radix-ui/themes`) and `@radix-ui/react-icons` — primitives.
- `portal/src/components/v2/` — design-system wrappers around those primitives (buttons, fields, callouts, dialogs).

Change `Storybook 9` to `Storybook 10`.

In the source-layout tree, change the `hook/` comment from `useCopyFeedback` to a hook that exists, `useAppConfigForm`. Delete the paragraph that cites `TextField.tsx`, `TextFieldWithCopyButton.tsx`, and `FormTextField.tsx`. Replace it with:

- New reusable UI goes in `src/components/v2/<Name>/`, one folder per component, story co-located.
- Screen-specific UI stays next to the screen (`src/graphql/portal/`, `src/graphql/adminapi/`, `src/components/<feature>/`).
- `components/v1` is not a directory.

### Providers

Replace the numbered provider list with the order in `portal/src/ReactApp.tsx`:

1. `AppLocaleProvider` — react-intl messages. `portal/src/components/common/AppLocaleProvider.tsx`.
2. Portal `ApolloProvider` — Portal API client.
3. `SystemConfigContext.Provider` — system config from `/api/system-config`. No theme object.
4. Inside `PortalRoot`: v2 `ThemeProvider` (`portal/src/components/v2/ThemeProvider/ThemeProvider.tsx`), which renders Radix `<Theme accentColor="indigo" grayColor="slate" className="contents">`.
5. Admin API `ApolloProvider` — created in `AppRoot`, scoped to the tenant in the URL.
6. React Router.

Delete "system config + FluentUI themes" and "FluentUI `ThemeProvider`".

### Styling

Delete:

- "Prefer FluentUI primitives for v1 screens."
- "Do not mix CSS Modules with `!important` to override FluentUI."

Replace with:

- New UI uses `src/components/v2/`. Import `Button` / `Text` / `Dialog` / `Select` from `@radix-ui/themes` only when no v2 wrapper exists for that control.
- Do not set a literal hex or `white`. Use the Radix tokens already documented in the Colour tokens section.
- CSS Modules style a component. Do not fight Radix with `!important`. Pass Radix props (`variant`, `color`, `size`) or a token.

Leave the Colour tokens section as it is. It already describes Radix steps and `light-theme` / `dark-theme`.

### Common utilities

Delete the `useCopyFeedback` / `TextFieldWithCopyButton` bullet. Copying is `CopyIconButton` in `portal/src/components/v2/CopyIconButton/CopyIconButton.tsx`, which calls `copyToClipboard` from `portal/src/util/clipboard.ts`.

Keep `ExternalLink`, `ErrorBoundSuspense`, and `useSystemConfig()`. Change the `useSystemConfig()` bullet so it no longer says "stories for v1 components". Any story whose component calls `useSystemConfig()` must provide `SystemConfigContext`.

Leave routing, GraphQL, i18n, forms, authentication, performance, and the "When editing a Portal screen" list, except the forms bullet that names components which still exist (`FormContainer`, `BlockerDialog`). Do not document removed form helpers.

## `portal/docs/ARCHITECTURE.md`

In the frontend stack list, apply the same two replacements as `FRONTEND.md`: Radix Themes plus `src/components/v2/` instead of FluentUI v8, and Storybook 10 instead of Storybook 9.

In key files:

- Change `src/hook/` example from `useCopyFeedback` to `useAppConfigForm`.
- Keep the `src/components/v2/` bullet.

In Configuration and theming, delete the `createTheme(...)` / `themes.main` sentence. `portal/src/system-config.ts` does not expose Fluent themes. Replace with:

- System config loads once into `SystemConfigContext`.
- Portal chrome color comes from the v2 `ThemeProvider` (Radix `Theme`) and from CSS variables set by `src/util/appearance.ts` (`light-theme` / `dark-theme` on `<html>`).
- AuthUI brand shades are a separate path: `deriveColors()` in `portal/src/util/theme.ts` calls `portal/src/util/shades.ts`. That function is what the Design screen and project wizard persist as button hover/active colors. It is not the portal's own theme.

## Component guidelines

Rewrite each file below so every named component is a file that exists. Do not leave a row whose import path 404s.

### `button-components.md`

The root wrappers `PrimaryButton`, `DefaultButton`, `ActionButton`, `ButtonWithLoading`, `MessageBarButton`, `CommandBarButton`, and `OutlinedActionButton` do not exist. `menuProps` is not a prop on the v2 buttons.

| Pattern | Component | Path |
|---|---|---|
| Primary action | `PrimaryButton` | `portal/src/components/v2/Button/PrimaryButton/PrimaryButton.tsx` |
| Secondary action | `SecondaryButton` | `portal/src/components/v2/Button/SecondaryButton/SecondaryButton.tsx` |
| Text-style action | `TextButton` | `portal/src/components/v2/Button/TextButton/TextButton.tsx` |
| Light fill on a dark surface | `WhiteButton` | `portal/src/components/v2/Button/WhiteButton/WhiteButton.tsx` |
| Icon-only action | `IconButton` | `portal/src/components/v2/IconButton/IconButton.tsx` |
| Button that looks like a link | `LinkButton` | `portal/src/LinkButton.tsx` |

`PrimaryButton` and `SecondaryButton` both take `loading?: boolean` and disable the button while `loading` is true. There is no `ButtonWithLoading`.

Rules to state:

- Import these wrappers. Do not import `Button` from `@radix-ui/themes` in a screen when one of these fits.
- One primary action per section or dialog footer. Other actions are `SecondaryButton` or `TextButton`.
- Destructive confirm uses `ConfirmationDialog` `confirmColor="red"`, not a separate destructive button component.
- Navigation uses `Link` / `ExternalLink`. `LinkButton` is a click handler that looks like a link.
- Labels go through `FormattedMessage` / `renderToString`.

Reference `portal/src/graphql/portal/EndpointDirectAccessScreen.tsx` only if it still imports v2 `PrimaryButton`. Reference `portal/src/components/common/DeleteConfirmationDialog.tsx` for dialog footer actions. Delete references to `FieldList.tsx`, `ShowError.tsx` message-bar buttons, `CommandBarPrimaryButton.tsx`, and `menuProps` call sites.

### `input-text-fields.md`

`FormTextField.tsx`, root `TextField.tsx`, and `TextFieldWithCopyButton.tsx` do not exist.

| Pattern | Component | Path |
|---|---|---|
| Label, hint, error around a control | `FormField` | `portal/src/components/v2/FormField/FormField.tsx` |
| Editable text | `TextField` | `portal/src/components/v2/TextField/TextField.tsx` |
| Copy an immutable value | `CopyIconButton` | `portal/src/components/v2/CopyIconButton/CopyIconButton.tsx` |

`TextField` accepts `type="password"`. Secret entry that needs reveal stays `PasswordField` (`portal/src/PasswordField.tsx`), documented in `input-password-fields.md`.

Point the "existing references" at `portal/src/graphql/adminapi/UserProfileForm.tsx` (v2 `TextField` + `FormField`) and `portal/src/graphql/portal/EditOAuthClientForm.tsx` (v2 `TextField` + `CopyIconButton`). Delete the Admin API endpoint / Project ID reference if that screen no longer uses a copy field; confirm with a grep for `CopyIconButton` in `AdminAPIConfigurationScreen.tsx` while editing, and cite that file only when the grep hits.

### `input-copy-text.md`

`TextWithCopyButton` and `useCopyFeedback` do not exist. One row: `CopyIconButton` for a value the user copies (identifier, URL, client ID). Do not use it for an editable draft. Cite `EditOAuthClientForm.tsx`. Delete the `ApplicationsConfigurationScreen.tsx` / `useCopyFeedback` reference.

### `input-password-fields.md`

Change the `TextField` import from `../../TextField` to `portal/src/components/v2/TextField/TextField.tsx`. Keep `PasswordField` at `portal/src/PasswordField.tsx`.

### `input-dropdowns.md`

`FormDropdown`, `SearchableDropdown`, and `CommandBarDropdown` do not exist. There is no shared dropdown wrapper.

| Pattern | Component | Import |
|---|---|---|
| Single choice from a short list | `Select` | `@radix-ui/themes` |
| Row or toolbar action menu | `DropdownMenu` | `@radix-ui/themes` |

Cite `portal/src/graphql/portal/CustomTextConfigurationScreen.tsx` (`Select.Root`) and `portal/src/UserProfileAttributesList.tsx` (`DropdownMenu`). State that option labels go through i18n, and that `Select` is controlled with `value` / `onValueChange`.

### `selection-checkbox.md`

| Pattern | Component | Import |
|---|---|---|
| Independent flag | `Checkbox` | `@radix-ui/themes` |
| Flag with a tooltip | `CheckboxWithTooltip` | `portal/src/CheckboxWithTooltip.tsx` |
| Flag that reveals content below | `CheckboxWithContentLayout` | `portal/src/CheckboxWithContentLayout.tsx` |

`CheckboxWithTooltip` already composes Radix `Checkbox` and v2 `Tooltip`. Keep the Login method configuration screen as the reference only where it still renders those two local components.

### `selection-radio.md`

`ChoiceGroup` does not exist.

| Pattern | Component | Path or import |
|---|---|---|
| Card with title and subtitle | `RadioCards` | `portal/src/components/v2/RadioCards/RadioCards.tsx` |
| Card with an icon | `IconRadioCards` | `portal/src/components/v2/IconRadioCards/IconRadioCards.tsx` |
| Compact single choice | `RadioGroup` | `@radix-ui/themes` |

Cite `LoginMethodConfigurationScreen.tsx` for `IconRadioCards` and `AnonymousUsersConfigurationScreen.tsx` for `RadioGroup`. Delete the `EndpointDirectAccessScreen` / `CreateOAuthClientScreen` `ChoiceGroup` claims.

### `selection-toggle.md`

Point `Toggle` at `portal/src/components/v2/Toggle/Toggle.tsx`. Replace the `ChoiceGroup` alternative with `RadioCards`, `IconRadioCards`, or Radix `RadioGroup`.

### `overlay-modal-dialogs.md`

| Pattern | Component | Path |
|---|---|---|
| Confirm / cancel | `ConfirmationDialog` | `portal/src/components/v2/ConfirmationDialog/ConfirmationDialog.tsx` |
| Delete confirm | `DeleteConfirmationDialog` | `portal/src/components/common/DeleteConfirmationDialog.tsx` |
| Unsaved-changes block | `BlockerDialog` | `portal/src/BlockerDialog.tsx` |
| Route leave | `NavigationBlockerDialog` | `portal/src/NavigationBlockerDialog.tsx` |
| Show / loading state | `useConfirmationDialog` | `portal/src/hook/useConfirmationDialog.tsx` |

`DeleteConfirmationDialog` renders `ConfirmationDialog`. Do not document a raw Fluent `Dialog` or `ReauthDialog` (`ReauthDialog.tsx` does not exist). One-off dialogs that are not confirm/cancel use `Dialog` from `@radix-ui/themes`. `ConfirmationDialog` uses `open` / `onOpenChange`. Cancel side effects go in `onCancel` or `onOpenChange(false)`, matching the comment in that file. Destructive confirm sets `confirmColor="red"`.

### `feedback-message-bars.md`

| Pattern | Component | Path |
|---|---|---|
| Inline status | `Callout` | `portal/src/components/v2/Callout/Callout.tsx` |
| Error-styled wrapper kept for existing imports | `RedMessageBar` | `portal/src/RedMessageBar.tsx` (renders `Callout` `type="error"`) |
| Parsed API / form errors | `ErrorMessageBar` | `portal/src/ErrorMessageBar.tsx` |
| Plan / feature disabled | `FeatureDisabledCallout` | `portal/src/components/v2/FeatureDisabledCallout/FeatureDisabledCallout.tsx` |

`CalloutType` is `"error" | "success" | "warning" | "info"`. New code imports `FeatureDisabledCallout` directly. `FeatureDisabledMessageBar.tsx` only re-exports that callout for old import sites; do not list it as the component to use.

`BlueMessageBar.tsx` and `MessageBarButton.tsx` do not exist. Delete those rows. In-callout actions are a `Link`, `ExternalLink`, or v2 button passed as `text`, not a message-bar button wrapper.

### `nav-link-components.md`

The three components stay: `Link` (`portal/src/Link.tsx`), `ExternalLink` (`portal/src/ExternalLink.tsx`), `LinkButton` (`portal/src/LinkButton.tsx`).

Replace the FluentUI explanation with the reason in `Link.tsx`:

- Tailwind preflight sets `a { color: inherit; text-decoration: inherit }`, so a bare `<a>` looks like body text.
- `WidgetDescription` renders Radix `Text` (`portal/src/WidgetDescription.tsx`). Radix `Text` remaps the accent scale, which would turn a link gray.
- `Link`, `ExternalLink`, and `LinkButton` pass Radix `Link` with `color="indigo"` so the link stays link-colored inside that `Text`.

Never use `Link` from `react-router-dom` or `portal/src/ReactRouterLink.tsx` for a visible portal link. `ReactRouterLink` renders a bare `<a>`.

### `layout-advanced-sections.md`

`FoldableDiv.tsx` does not exist. The only row is `Accordion` at `portal/src/components/common/Accordion.tsx`. Replace the FluentUI `Text` sentence with the same Radix `Text` rule as `nav-link-components.md`. Cite `portal/src/graphql/portal/EditOAuthClientForm.tsx` for `Accordion`. Delete the `AddUserScreen` / `SingleSignOnConfigurationWidget` `FoldableDiv` references.

### `list-add-actions.md`

`ActionButton.tsx` and `FieldList.tsx` do not exist.

| Pattern | Component | Path |
|---|---|---|
| Editable list of strings | `TextFieldList` | `portal/src/components/v2/TextFieldList/TextFieldList.tsx` |
| Any other append action | `SecondaryButton` or `TextButton` | v2 button paths above |

`TextFieldList` already renders a `SecondaryButton` for add and takes `onListItemAdd`, `onListItemChange`, `onListItemDelete`, and `addButtonLabelMessageID`.

### `layout-screen-content.md` and `screen-scaffold-template.md`

`ScreenTitle.tsx`, `ScreenDescription.tsx`, and `Widget.tsx` do not exist. `ScreenLayoutScrollView.tsx`, `ScreenContent.tsx`, `ScreenHeader.tsx`, `NavBreadcrumb.tsx`, `ShowLoading.tsx`, and `ShowError.tsx` do.

Rewrite both docs to the structure of `portal/src/graphql/portal/AnonymousUsersConfigurationScreen.tsx`:

- `ShowLoading` / `ShowError` gate the screen.
- `ScreenContent` wraps the page.
- Sections are `SettingsSectionCard` (`portal/src/components/v2/SettingsSectionCard/SettingsSectionCard.tsx`).
- The form shell is `FormContainer` (`portal/src/FormContainer.tsx`).
- The save bar is `SaveFunctionBar` (`portal/src/components/v2/SaveFunctionBar/SaveFunctionBar.tsx`).
- Headings are Radix `Heading`, not `ScreenTitle`.

The scaffold code sample must import those modules. Do not leave an import of `Widget`, `ScreenTitle`, or `ScreenDescription`.

### Guidelines with no FluentUI or deleted-component claims

Do not edit `portal/docs/components-guidelines/` files other than the ones named above, unless a grep for `FluentUI`, `@fluentui`, `TextFieldWithCopyButton`, `ChoiceGroup`, `ScreenTitle`, or `Widget` inside that file hits. If it hits, fix that sentence in the same commit. Do not restyle unrelated guidance.

## `.claude/skills/update-portal-ui/SKILL.md`

This skill is what later edits follow. Update it in the same commit as `nav-link-components.md`, or the skill will reintroduce the FluentUI rule.

In "Link components":

- Keep the table of `Link`, `ExternalLink`, and `LinkButton`.
- Replace "no FluentUI wrapper" / `FluentLink` with the Radix `Link` `color="indigo"` plus Tailwind preflight explanation from `Link.tsx`.
- Replace "WidgetDescription wraps FluentUI `Text`" with "WidgetDescription renders Radix `Text`".
- Keep the rule: inside `WidgetDescription` or Radix `Text`, use `portal/src` `Link` or `ExternalLink`, not `react-router-dom`'s `Link` and not `ReactRouterLink`.

In "Passing rich content to callbacks":

- Delete the `ChoiceGroup` / `onRenderLabel` / `IChoiceGroupOption` sample. `ChoiceGroup` does not exist.
- Keep the rule that a callback which renders a link is typed `React.ReactNode`, not `string`. Use a plain function-argument example, not a Fluent type.

In the verification checklist, change "FluentUI `Text`" to "Radix `Text` (including `WidgetDescription`)".

Do not change the i18n, config-form, query, or dead-code sections of the skill. Those are still accurate. The dead-code section already names Radix `rt-*` classes.

## File-level change plan

Docs and the skill, no behavior change:

- `portal/docs/FRONTEND.md`
- `portal/docs/ARCHITECTURE.md`
- `portal/docs/storybook.md`
- `portal/docs/components-guidelines/button-components.md`
- `portal/docs/components-guidelines/input-text-fields.md`
- `portal/docs/components-guidelines/input-copy-text.md`
- `portal/docs/components-guidelines/input-password-fields.md`
- `portal/docs/components-guidelines/input-dropdowns.md`
- `portal/docs/components-guidelines/selection-checkbox.md`
- `portal/docs/components-guidelines/selection-radio.md`
- `portal/docs/components-guidelines/selection-toggle.md`
- `portal/docs/components-guidelines/overlay-modal-dialogs.md`
- `portal/docs/components-guidelines/feedback-message-bars.md`
- `portal/docs/components-guidelines/nav-link-components.md`
- `portal/docs/components-guidelines/layout-advanced-sections.md`
- `portal/docs/components-guidelines/list-add-actions.md`
- `portal/docs/components-guidelines/layout-screen-content.md`
- `portal/docs/components-guidelines/screen-scaffold-template.md`
- `.claude/skills/update-portal-ui/SKILL.md`

Story titles, no render change:

- `portal/src/graphql/portal/LoginMethodChooser.stories.tsx`
- `portal/src/graphql/portal/LoginMethodIcon.stories.tsx`
- `portal/src/graphql/portal/IconRadioCards.stories.tsx`
- `portal/src/CheckboxWithTooltip.stories.tsx`
- `portal/src/graphql/adminapi/AccountStatusDialog.stories.tsx`

No generated files. No `make export-schemas`. No `.vettedpositions`.

## Test plan

No new unit or e2e tests. Titles are strings. Docs are not executed.

After the title edit:

- `cd portal && npm run typecheck`
- Grep `portal/` for `components/v1`, `FluentUI`, `@fluentui`, `TextFieldWithCopyButton`, `useCopyFeedback`, `ChoiceGroup`, and `FormTextField`. The only allowed hits are historical comments that name the removed library on purpose: `portal/src/util/shades.ts`, `portal/src/util/theme.ts`, `portal/src/util/theme.test.ts`, and the two `replacing the FluentUI TextFieldWithCopyButton` comments in `EditOAuthClientForm.tsx` and `EditOAuthClientFormFrameworkQuickStart.tsx`. Do not delete those comments in this work.
- With Storybook already running, confirm the sidebar no longer has a `components/v1` group, `components/v2` is unchanged, and the five retitled stories render.

## Fixed decisions

- Keep the `components/v2` sidebar group. It matches the source folder.
- Remove the `components/v1` title prefix. Do not hide or delete those stories.
- Login-method stories, including the one currently titled `IconRadioCards`, move under `portal/LoginMethods/`.
- `AccountStatusDialog` moves under `portal/Users/`.
- `CheckboxWithTooltip` becomes `components/CheckboxWithTooltip`.
- `shades.ts` stays, untouched.
- Component guidelines name only files that exist today. Missing Fluent wrappers are not given new shims.

## Atomic commit plan

### Commit 1 — Storybook titles and convention

Purpose: stop Storybook from presenting a v1 generation.

Files:

- the five `*.stories.tsx` files listed above
- `portal/docs/storybook.md`

No generated files.

### Commit 2 — Docs and the portal UI skill

Purpose: stop docs and the `update-portal-ui` skill from telling the next edit to use FluentUI or deleted components.

Files: every doc and `.claude/skills/update-portal-ui/SKILL.md` listed in the file-level plan, except `storybook.md` (that landed in commit 1).

No generated files.

Suggested subjects:

- `doc: Retitle portal Storybook stories off components/v1`
- `doc: Describe the portal UI as Radix and components/v2`
