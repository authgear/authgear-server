# Link components

Use the correct link component based on navigation type and rendering context.

## Which component to use

| Component | Import path | Use when |
|---|---|---|
| `Link` | `../../Link` (or relative path to `portal/src/Link.tsx`) | Internal navigation (React Router) |
| `ExternalLink` | `../../ExternalLink` | External URLs (`href`, opens in new tab) |
| `LinkButton` | `../../LinkButton` | A button that visually looks like a link |

## Rules

- Never use `Link` from `react-router-dom` directly in portal UI.
- Whenever a link appears inside `WidgetDescription`, FluentUI `Text`, or any wrapper that internally uses FluentUI `Text`, use `Link` or `ExternalLink` from `portal/src`.
- For inline links in i18n text (`FormattedMessage`), use XML-like tags in translation strings and provide render functions through `values`.
- Use `Link` for internal routes and `ExternalLink` for external URLs in all i18n inline-link render functions.
- For callbacks that render label/description content with links, use `React.ReactNode` instead of `string`.

## Why this matters

- `portal/src/Link.tsx` and `portal/src/ExternalLink.tsx` wrap FluentUI link components and preserve link styling inside FluentUI `Text`.
- `react-router-dom` `Link` renders a plain `<a>` tag, which may lose expected visual style inside FluentUI text wrappers.

## Existing references

- `portal/src/Link.tsx` internal FluentUI-compatible link wrapper.
- `portal/src/ExternalLink.tsx` external FluentUI-compatible link wrapper.
- `portal/src/graphql/portal/EndpointDirectAccessScreen.tsx` inline i18n links in `FormattedMessage`.
