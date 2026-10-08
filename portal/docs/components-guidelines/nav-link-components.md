# Link components

Use the correct link component based on navigation type and rendering context.

## Which component to use

| Component | Import path | Use when |
|---|---|---|
| `Link` | `portal/src/Link.tsx` | Internal navigation (React Router) |
| `ExternalLink` | `portal/src/ExternalLink.tsx` | External URLs (`href`, opens in a new tab) |
| `LinkButton` | `portal/src/LinkButton.tsx` | A button that visually looks like a link |

## Rules

- Never use `Link` from `react-router-dom` directly in portal UI.
- Never use `portal/src/ReactRouterLink.tsx` for a visible link. It renders a bare `<a>`.
- Whenever a link appears inside `WidgetDescription` or Radix `Text`, use `Link` or `ExternalLink` from `portal/src`.
- For inline links in i18n text (`FormattedMessage`), use XML-like tags in translation strings and provide render functions through `values`.
- Use `Link` for internal routes and `ExternalLink` for external URLs in those render functions.
- For callbacks that render label or description content with links, use `React.ReactNode` instead of `string`.

## Why this matters

- Tailwind preflight sets `a { color: inherit; text-decoration: inherit }`, so a bare `<a>` looks like body text.
- `WidgetDescription` renders Radix `Text` (`portal/src/WidgetDescription.tsx`). Radix `Text` remaps the accent scale, which would turn a link gray.
- `Link`, `ExternalLink`, and `LinkButton` render Radix `Link` with `color="indigo"`, so the link stays link-colored inside that `Text`.

## Existing references

- `portal/src/Link.tsx` internal link. It pins `color="indigo"`.
- `portal/src/ExternalLink.tsx` external link, same color pin.
- `portal/src/graphql/portal/EndpointDirectAccessScreen.tsx` uses inline i18n links in `FormattedMessage`.
