---
name: update-portal-ui
description: Guidelines for updating or designing pages in the portal React frontend (portal/src). Covers component conventions, link rendering rules, i18n patterns, and common pitfalls.
---

Follow this skill when adding, editing, or reviewing UI in `portal/src`.

## All user-facing text must be translated

Every string a user can see or hear must go through `renderToString`/`FormattedMessage`/locale-data (`portal/src/locale-data/en.json`) — never a bare string literal. This applies beyond obvious JSX text nodes:

- **Chart/graph library config**: dataset `label`s, legend text, tooltip callbacks, axis titles passed into `chart.js` (or any charting lib) config objects are still user-facing text, even though they live inside a plain JS config object, not JSX. Wrap them with `renderToString` the same as any other label.
- **Locale-aware formatting of derived values**: any `Intl.DisplayNames`, `Intl.NumberFormat`, `Intl.DateTimeFormat`, or `luxon` `DateTime#toFormat`/`toLocaleString` call that produces user-visible output (country names, chart axis date labels, etc.) must be constructed with the active portal locale, not a hardcoded locale (e.g. `new Intl.DisplayNames(["en"], ...)`) and not left to the library's default. Grep for `.toFormat(`/`.toLocaleString(`/`new Intl.` in your diff and confirm each one is passed (or chained with) the active `locale`, not silently defaulting.
- A string can be "translated" everywhere else in a file and still miss one of these — check every literal individually, don't assume a file is compliant because most of it uses `FormattedMessage`.

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

- [ ] No hardcoded user-facing string literals anywhere in the diff, including non-JSX config objects (chart library `label`/legend/tooltip config, form option lists, etc.) — all go through `renderToString`/`FormattedMessage`.
- [ ] Every `Intl.DisplayNames`/`Intl.NumberFormat`/`Intl.DateTimeFormat`/luxon `toFormat`/`toLocaleString` call that produces user-visible text uses the active portal locale, not a hardcoded or default locale.
- [ ] Links inside `WidgetDescription` or FluentUI `Text` use `Link` or `ExternalLink` from `portal/src`, not from `react-router-dom`.
- [ ] Inline links in `FormattedMessage` `values` use `Link` or `ExternalLink` from `portal/src`.
- [ ] Callbacks that may receive rich content (links, JSX) are typed `React.ReactNode`, not `string`.
- [ ] Run `cd portal && npm run typecheck` — must pass clean.
