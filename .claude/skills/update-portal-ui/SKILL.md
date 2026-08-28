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

**Never** use a component that renders a bare `<a>`. Two exist, and naming only
the first is how this rule gets passed while being violated:

- `Link` from `react-router-dom`
- `ReactRouterLink` from `portal/src/ReactRouterLink.tsx` — a *local* file, so it
  looks like a portal component, but it returns `<a {...rest} href={href} …>`
  with no FluentUI wrapper

Tailwind preflight ships `a{color:inherit;text-decoration:inherit}`, so a bare
`<a>` inherits the surrounding colour and loses its underline: it renders as
ordinary body text with no affordance that it is clickable. `portal/src/Link.tsx`
and `ExternalLink.tsx` wrap FluentUI's `FluentLink`, which keeps its own styling.

The test is what the component renders, not where it is imported from. If in
doubt, open it and look for a bare `<a>`.

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

## Config-backed forms (`useAppConfigForm`)

Screens that edit `authgear.yaml` share three traps. All three shipped in the DCR
settings tab and were caught only in review.

**Never restate a server-side config default.** `effectiveAppConfig` resolves to
`Context.Config.AppConfig`, which is post-`SetDefaults`, and `SetFieldDefaults`
force-allocates whole sections — so `form.effectiveConfig` carries concrete values
even when `authgear.yaml` omits the section entirely. Read placeholders and
fallbacks from there. A local `const DEFAULT_FOO = 1800` duplicating a Go constant
will drift, and the drift is silent because nothing type-checks it. If you think
you need one, probe the real value first:

```go
// throwaway test in pkg/lib/config
c := &SomeConfig{}; c.SetDefaults(); fmt.Printf("%+v\n", c)
```

**Bind schema validation errors to their field.** Any input written to app config
must pass `parentJSONPointer` + `fieldName` (the v2 `TextField` and `FormTextField`
both support them, plus `errorRules`). Without them a `minimum`/`type` violation
falls through to the generic error bar with no indication which field caused it.
The pointer is the parent object's, e.g.
`/oauth/dynamic_client_registration/default_client_config` + `access_token_lifetime_seconds`.

**Do not save from an individual control.** `saveWith(fn)` applies `fn` to
`currentState` and saves *the whole form* — including edits elsewhere on the screen,
and on sibling tabs that share the same form model — then clears the dirty flag, so
the save bar disappears as if nothing had been pending. A toggle that "should take
effect immediately" still belongs behind the Save button; use `setState` like every
other control.

## User-facing copy must match enforced behaviour

Precision about *who* a permission covers is a correctness question, not a style one.
Verify a claim against the code that enforces it, not against the config field's name:
`allow_dynamic_third_party_client_access` gates only clients that are dynamic **and**
third-party (`handler_authz.go`: `client.IsDynamicClient() && client.IsThirdParty()`),
so "dynamically registered clients" and "third-party clients" are each wrong — one
lets in dynamic first-party clients, the other static third-party ones.

- After correcting one such string, grep for siblings making the same claim. These
  travel in families and get fixed one at a time otherwise.
- Go `Description` fields on GraphQL types ship in the published `schema.graphql` —
  they are customer-facing docs, and are worth the same check.
- A setting with both a create and an edit surface must **share** the string, not
  own a near-duplicate key. Diverging labels for one setting is a bug: grep the
  message id family (`Foo.bar.label` vs `CreateFoo.bar.label`) and reuse one key.
- Copy shown for a state that is not yet in effect must say so, or be hidden until
  it is. A warning that is false when displayed teaches admins to dismiss it.

## Verification checklist

Before submitting a portal UI change:

- [ ] No hardcoded user-facing string literals anywhere in the diff, including non-JSX config objects (chart library `label`/legend/tooltip config, form option lists, etc.) — all go through `renderToString`/`FormattedMessage`.
- [ ] Every `Intl.DisplayNames`/`Intl.NumberFormat`/`Intl.DateTimeFormat`/luxon `toFormat`/`toLocaleString` call that produces user-visible text uses the active portal locale, not a hardcoded or default locale.
- [ ] Links inside `WidgetDescription` or FluentUI `Text` use `Link` or `ExternalLink` from `portal/src` — not `react-router-dom`'s `Link` and not `portal/src/ReactRouterLink`, both of which render a bare `<a>`.
- [ ] Inline links in `FormattedMessage` `values` use `Link` or `ExternalLink` from `portal/src`.
- [ ] Callbacks that may receive rich content (links, JSX) are typed `React.ReactNode`, not `string`.
- [ ] No local constant restates a server-side config default; placeholders and fallbacks come from `form.effectiveConfig`.
- [ ] Every config-backed input passes `parentJSONPointer` + `fieldName` so schema errors land on the field.
- [ ] No individual control calls `saveWith`; nothing is written until Save.
- [ ] Copy naming who a permission covers was checked against the enforcing code, and sibling strings making the same claim were grepped and fixed together.
- [ ] A setting with create and edit surfaces shares one message id rather than near-duplicate keys.
- [ ] Run `cd portal && npm run typecheck` — must pass clean.
