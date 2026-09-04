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

### Always prefer a static message id over a computed one

Write the id as a literal that appears verbatim in the source:

```tsx
// Good — greppable
<FormattedMessage id="UserDetails.connected-identities.email" />

// Avoid — invisible to a search for the key
<FormattedMessage id={"UserDetails.connected-identities." + kind} />
{renderToString(`standard-attribute.${fieldName}`)}
```

A computed id makes the key unfindable in both directions: you cannot grep from
`en.json` to the code that renders it, and you cannot grep from the code to see which
strings a screen can actually produce. That costs real correctness, not just
convenience — a computed id silently survives every dead-key sweep, so orphans
accumulate forever, and renaming or deleting the wrong one ships a raw key id to
users because nothing fails at build time.

Prefer an explicit lookup table over string concatenation when a key genuinely varies —
each id stays a literal, so both greps work and TypeScript catches a missing case. This
is the established pattern here; copy it rather than concatenating. See
`AddUserScreen.tsx:65` (`loginIdTypeNameIds`) and `MFAConfigurationScreen.tsx:61`
(`secondaryAuthenticatorNameIds`):

```tsx
const loginIdTypeNameIds: Record<LoginIDKeyType, string> = {
  username: "login-id-key.username",
  email: "login-id-key.email",
  phone: "login-id-key.phone",
};
<FormattedMessage id={loginIdTypeNameIds[loginIdType]} />
```

Typing the table as `Record<SomeUnion, string>` is what buys the exhaustiveness check —
a bare object literal will not fail when a union member is added later.

Passing an id through a variable or prop (`id={labelKey}`, `id={messageID}`) is fine —
the call site still holds a literal. What to avoid is *synthesising* the id from
fragments. Concatenation is justified only for a genuinely open-ended set enumerated
elsewhere (country codes, locale tags); when you do it, note the prefix in the
dynamic-key list under "Clean up what your change orphans" below so the next dead-key
sweep does not delete live strings.

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

## Clean up what your change orphans

A change that stops using a message id, a CSS class, or a component must delete it in
the same commit. Nothing in CI catches this — `typecheck`, `eslint`, `stylelint` and
`build` all pass with a locale file full of unreachable keys and `.module.css` files
full of unreachable rules. A migration that replaces a screen leaves both behind by
default, and "we'll clean it up later" means the next reader cannot tell which keys
are live.

**Find orphaned i18n keys by diffing the orphan set, not by grepping once.** A plain
"key not found in `src`" sweep over `en.json` reports hundreds of false positives,
because many ids are built at runtime. Instead compute the orphan set at your merge
base and at `HEAD`, and report only keys that *became* orphans — dynamic keys are
orphans in both and cancel out:

```bash
git archive "$(git merge-base HEAD main)" portal/src | tar -x -C /tmp/base
# for each key in portal/src/locale-data/en.json: is it referenced in
# /tmp/base/portal/src but no longer in portal/src? that set is what you orphaned.
```

**Before deleting a key, confirm it is not built dynamically.** Grep for the key's
prefix used in concatenation or a template literal. Known dynamic families include
`standard-attribute.` + field, `AuditLogActivityType.` + type, `Territory.` + alpha2,
`Locales.` + tag, `custom-attribute-type.` + key, `MFAConfigurationScreen.policy.mode.`
+ option. Ids passed through a variable (`id={labelKey}`, `id={messageID}`) are safe —
the call site still holds a literal — but a key reached only via `"prefix." + name` is
live even though nothing greps for it. Check whether the *specific* value is reachable
(e.g. `standard-attribute.updated_at` is dead because `updated_at` appears in no
section pointer list), not merely whether the family is dynamic.

This list is the running cost of every computed id, which is why new code should use a
static one — see "Always prefer a static message id over a computed one" above. Do not
grow the list without cause; if your change makes a family static again, delete its
entry here.

**Dead CSS: compare declared classes against the paired `.tsx`.** For every
`.module.css` your change touches, list its declared classes and confirm each is still
referenced as `styles.foo` (or `styles["foo"]`) somewhere in the tree. Two important
exclusions — these are *not* dead:

- `:global(...)` selectors targeting third-party DOM: Radix internals (`rt-*`),
  intl-tel-input (`iti`, `iti__*`), cropperjs (`cropper-*`). They are styled by class
  name emitted by the library, never via `styles.foo`.
- A custom property (`--foo`) read by a nested library component rather than by a rule
  in the same file.

Conversely, when you delete the `:global(...)` rules for a library you just removed,
check whether a custom property declared alongside them (e.g. `--text-field-height`)
had no other consumer — if so it dies with them.

**Dead code.** Removing the last usage of a component, hook, or util means deleting the
file, its `.module.css`, and its stories. `npm run build` succeeding proves nothing —
an unimported module is simply dropped from the bundle, silently.

## Verification checklist

Before submitting a portal UI change:

- [ ] No hardcoded user-facing string literals anywhere in the diff, including non-JSX config objects (chart library `label`/legend/tooltip config, form option lists, etc.) — all go through `renderToString`/`FormattedMessage`.
- [ ] Every `Intl.DisplayNames`/`Intl.NumberFormat`/`Intl.DateTimeFormat`/luxon `toFormat`/`toLocaleString` call that produces user-visible text uses the active portal locale, not a hardcoded or default locale.
- [ ] Every message id is a static literal, greppable from `en.json` to its render site — no id synthesised by concatenation or template literal. Where a varying key was unavoidable, it uses a lookup table of literals, or the new prefix is recorded in the dynamic-key list.
- [ ] Links inside `WidgetDescription` or FluentUI `Text` use `Link` or `ExternalLink` from `portal/src` — not `react-router-dom`'s `Link` and not `portal/src/ReactRouterLink`, both of which render a bare `<a>`.
- [ ] Inline links in `FormattedMessage` `values` use `Link` or `ExternalLink` from `portal/src`.
- [ ] Callbacks that may receive rich content (links, JSX) are typed `React.ReactNode`, not `string`.
- [ ] No local constant restates a server-side config default; placeholders and fallbacks come from `form.effectiveConfig`.
- [ ] Every config-backed input passes `parentJSONPointer` + `fieldName` so schema errors land on the field.
- [ ] No individual control calls `saveWith`; nothing is written until Save.
- [ ] Copy naming who a permission covers was checked against the enforcing code, and sibling strings making the same claim were grepped and fixed together.
- [ ] A setting with create and edit surfaces shares one message id rather than near-duplicate keys.
- [ ] Every i18n key whose last reference this change removed is deleted from `locale-data/en.json` — found by diffing the orphan set at the merge base against `HEAD`, and each one checked against the dynamic-key families before deleting.
- [ ] Every CSS class this change stopped using is deleted from its `.module.css`, excluding `:global(...)` selectors for third-party DOM (`rt-*`, `iti*`, `cropper-*`), and any custom property left with no consumer went with them.
- [ ] Every component/hook/util this change stopped importing is deleted along with its `.module.css` and stories — a green `build` does not prove otherwise.
- [ ] Run `cd portal && npm run typecheck` — must pass clean.
