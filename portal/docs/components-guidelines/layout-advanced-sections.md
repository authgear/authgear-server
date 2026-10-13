# Expandable sections (Advanced)

Use `Accordion` for settings that reveal extra controls on demand. Do not introduce a one-off expand/collapse trigger when it fits.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Self-contained expand/collapse | `Accordion` | `portal/src/components/common/Accordion.tsx` | A section can show or hide its own extra controls |

## UX and content rules

- Use this pattern for "Advanced", "Optional", or secondary settings that should be hidden by default.
- The toggle label must come from i18n (for example via `FormattedMessage`), not hard-coded text. Pass it as `Accordion`'s `text`.
- If the expanded content includes links inside `WidgetDescription` or Radix `Text`, use `Link` or `ExternalLink` from `portal/src`. See `nav-link-components.md`.

## Existing references

- `portal/src/graphql/portal/EditOAuthClientForm.tsx` uses `Accordion` for local, self-contained expandable content.
