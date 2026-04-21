# Screen layout content

Use the shared screen layout components so portal pages have consistent structure, spacing, and loading/error behavior.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Top-level page scroll container | `ScreenLayoutScrollView` | `../../ScreenLayoutScrollView` | A screen needs the standard scrollable body area under portal chrome |
| Screen content container | `ScreenContent` | `../../ScreenContent` | A screen renders title/description/widgets in the standard content width |
| Screen heading | `ScreenTitle` | `../../ScreenTitle` | The page needs a main H1 heading |
| Screen intro text | `ScreenDescription` | `../../ScreenDescription` | The page needs a short explanatory paragraph under title |
| Header/chrome area | `ScreenHeader` | `../../ScreenHeader` | Top-level pages need the shared portal header/navigation area |
| Query-state gates | `ShowLoading`, `ShowError` | `../../ShowLoading`, `../../ShowError` | Async page data must render loading/error before main content |

## Rules

- For data-driven screens, gate content with `ShowLoading` and `ShowError` before rendering `ScreenContent`.
- Prefer this structure for most settings pages: `ScreenLayoutScrollView` -> `ScreenContent` -> (`NavBreadcrumb`) -> `ScreenTitle` -> `ScreenDescription` -> widgets/forms.
- Use one `ScreenTitle` per screen content area; avoid multiple H1-level titles in a single page body.
- Keep screen title/description strings in i18n (`FormattedMessage` / `renderToString`), not hard-coded text.
- Use `ScreenContent` `layout` prop intentionally:
- `auto-rows` (default) for form/widget pages.
- `list` for list-heavy pages where items should flow with list spacing behavior.
- If a page needs a special header block above content body, pass it via `ScreenContent` `header` prop instead of ad-hoc wrappers.
- Reuse `ScreenHeader` at route/shell level; feature screens should focus on body content and avoid re-implementing page chrome.

## Width policy (narrow vs full-width)

- Default to narrow content for settings/detail/editing flows. In practice this is the common 8-column content span on desktop (`grid-column: 1 / span 8`).
- Use full-width content for list/table/dense overview pages where horizontal space improves scanability. In practice this is 12-column span plus list-style content layout (`ScreenContent layout="list"` + `grid-column: 1 / span 12`).
- For drill-down flows, it is expected to go from full-width list to narrower detail editor (for example applications list -> application edit page).
- Use split-width variants only when information architecture needs it (for example primary editing area plus side quick-start/help column).
- Unless there is a strong UX reason, do not introduce custom width behavior outside these patterns.

## Existing references

- `portal/src/graphql/portal/GoogleTagManagerConfigurationScreen.tsx` uses `ShowLoading`/`ShowError` and `ScreenContent` with breadcrumb, title, and widget content.
- `portal/src/graphql/portal/HookConfigurationScreen.tsx` uses `ScreenTitle` and `ScreenDescription` as standard screen heading structure.
- `portal/src/graphql/portal/ApplicationsConfigurationScreen.tsx` uses full-width list layout (`ScreenContent layout="list"` with 12-column content widgets).
- `portal/src/graphql/portal/EditOAuthClientScreen.tsx` uses narrower/split detail editing layout after entering a specific application.
- `portal/src/graphql/portal/AppsScreen.tsx` uses `ScreenHeader` and `ScreenLayoutScrollView` for page shell + scrollable content.
- `portal/src/graphql/portal/ProjectRootScreen.tsx` shows loading-only redirect flow with `ShowLoading`.
- `portal/src/ScreenContent.tsx` defines shared container behavior and `layout` options (`list`, `auto-rows`).
