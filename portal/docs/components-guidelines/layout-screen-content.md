# Screen layout content

Use the shared screen layout components so portal pages have consistent structure, spacing, and loading/error behavior.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Top-level page scroll container | `ScreenLayoutScrollView` | `portal/src/ScreenLayoutScrollView.tsx` | A screen needs the standard scrollable body area under portal chrome |
| Screen content container | `ScreenContent` | `portal/src/ScreenContent.tsx` | A screen renders its sections in the standard content width |
| Page heading | `Heading` | `@radix-ui/themes` | The page needs one `h1`. Use `as="h1"` |
| Page intro text | `Text` | `@radix-ui/themes` | A short explanation under the heading (`as="p"`) |
| Settings section | `SettingsSectionCard` | `portal/src/components/v2/SettingsSectionCard/SettingsSectionCard.tsx` | A titled block of settings |
| Save bar | `SaveFunctionBar` | `portal/src/components/v2/SaveFunctionBar/SaveFunctionBar.tsx` | The screen edits config and needs save / discard |
| Header/chrome area | `ScreenHeader` | `portal/src/ScreenHeader.tsx` | A top-level page needs the shared portal header |
| Query-state gates | `ShowLoading`, `ShowError` | `portal/src/ShowLoading.tsx`, `portal/src/ShowError.tsx` | Async page data must render loading or error before main content |

## Rules

- For data-driven screens, gate content with `ShowLoading` and `ShowError` before rendering `ScreenContent`.
- A settings page is `ShowLoading` / `ShowError`, then `FormContainer`, then `ScreenContent`, a Radix `Heading`, `SettingsSectionCard` sections, and `SaveFunctionBar`. See `AnonymousUsersConfigurationScreen.tsx`.
- Use one `h1` per screen content area.
- Keep title, description, and labels in i18n (`FormattedMessage` / `renderToString`).
- Use `ScreenContent`'s `layout` prop intentionally: `auto-rows` (default) for form pages, `list` for list-heavy pages.
- If a page needs a block above the content body, pass it via `ScreenContent`'s `header` prop.
- Reuse `ScreenHeader` at route/shell level. Feature screens should not re-implement page chrome.

## Width policy (narrow vs full-width)

- Default to narrow content for settings, detail, and editing flows. In practice this is the common 8-column content span on desktop (`grid-column: 1 / span 8`).
- Use full-width content for list, table, and dense overview pages. In practice this is a 12-column span plus `ScreenContent layout="list"`.
- For drill-down flows, go from a full-width list to a narrower detail editor.
- Use a split width only when the page needs a primary editing area plus a side column.
- Do not introduce a custom width outside these patterns unless the layout cannot express it.

## Existing references

- `portal/src/graphql/portal/AnonymousUsersConfigurationScreen.tsx` is the settings-screen structure: `ShowLoading`, `ShowError`, `FormContainer`, `ScreenContent`, Radix `Heading`, `SettingsSectionCard`, `SaveFunctionBar`.
- `portal/src/graphql/adminapi/UsersScreen.tsx` uses `ScreenContent layout="list"`.
- `portal/src/graphql/portal/VerifyDomainScreen.tsx` uses `ScreenLayoutScrollView`.
- `portal/src/graphql/portal/AppsScreen.tsx` uses `ScreenHeader`.
- `portal/src/ScreenContent.tsx` defines `layout` (`list`, `auto-rows`) and the `header` prop.
