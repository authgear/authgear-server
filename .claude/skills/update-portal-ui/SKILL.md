---
name: update-portal-ui
description: Guidelines for updating or designing pages in the portal React frontend (portal/src). Covers component conventions, link rendering rules, i18n patterns, and common pitfalls.
---

Follow this skill when adding, editing, or reviewing UI in `portal/src`.

## Link components

The portal has three link components. Use the right one — using the wrong one causes links to render as unstyled plain text inside certain wrappers.

| Component | Import path | Use when |
|---|---|---|
| `Link` | `../../Link` (or relative path to `portal/src/Link.tsx`) | Internal navigation (React Router) |
| `ExternalLink` | `../../ExternalLink` | External URLs (`href`, opens in new tab) |
| `LinkButton` | `../../LinkButton` | A button that visually looks like a link |

**Never** use `Link` from `react-router-dom` directly — it renders a plain `<a>` tag with no FluentUI styling.

### Why this matters: the WidgetDescription / Text trap

`WidgetDescription` wraps its children in a FluentUI `Text` component. FluentUI's `Text` overrides the colour of plain `<a>` tags to match surrounding text, making links invisible as links.

- `portal/src/Link.tsx` and `portal/src/ExternalLink.tsx` both wrap FluentUI's `FluentLink`, which keeps its own link styling even inside `Text`. ✓
- `react-router-dom`'s `Link` renders a bare `<a>` — styling is stripped inside `Text`. ✗

**Rule:** Whenever a link appears inside `WidgetDescription`, `Text` (FluentUI), or any component that internally wraps FluentUI `Text`, use `Link` or `ExternalLink` from `portal/src`, not from `react-router-dom`.

### Inline links inside FormattedMessage (i18n)

To embed a clickable link inside a translated string:

1. In the translation string (`portal/src/locale-data/en.json`), use an XML-like tag:
   ```
   "my-key": "Read the <docLink>documentation</docLink> for details."
   ```

2. In the component, pass a render function in `FormattedMessage` `values` whose key matches the tag name exactly:
   ```tsx
   <FormattedMessage
     id="my-key"
     values={{
       // eslint-disable-next-line react/no-unstable-nested-components
       docLink: (chunks: React.ReactNode) => (
         <ExternalLink href="https://docs.authgear.com/...">
           {chunks}
         </ExternalLink>
       ),
     }}
   />
   ```

3. Use `Link` for internal routes, `ExternalLink` for external URLs. **Never** use react-router-dom's `Link` here.

### Passing rich content to callbacks that accept descriptions

Some components (e.g. FluentUI `ChoiceGroup` via `onRenderLabel`) accept a label-render callback. If the description contains a link, the callback must accept `React.ReactNode`, not `string`:

```tsx
// Correct — accepts ReactNode so JSX can be passed
const onRenderLabel = useCallback((description: React.ReactNode) => {
  return (option?: IChoiceGroupOption) => (
    <div>
      <Text>{option?.text}</Text>
      <Text>{description}</Text>
    </div>
  );
}, []);

// Then pass FormattedMessage directly — no cast needed
onRenderLabel(
  <FormattedMessage id="..." values={{ reactRouterLink: ... }} />
)
```

**Never** cast JSX to string with `as any as string` — the link will not render correctly.

## Verification checklist

Before submitting a portal UI change:

- [ ] Links inside `WidgetDescription` or FluentUI `Text` use `Link` or `ExternalLink` from `portal/src`, not from `react-router-dom`.
- [ ] Inline links in `FormattedMessage` `values` use `Link` or `ExternalLink` from `portal/src`.
- [ ] Callbacks that may receive rich content (links, JSX) are typed `React.ReactNode`, not `string`.
- [ ] Run `cd portal && npm run typecheck` — must pass clean.
