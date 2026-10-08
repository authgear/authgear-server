# Message bars

Use a callout for inline status, warning, and error feedback that must stay on the page.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Inline status | `Callout` | `portal/src/components/v2/Callout/Callout.tsx` | Success, info, warning, or error feedback in the page |
| Error-styled wrapper kept for existing imports | `RedMessageBar` | `portal/src/RedMessageBar.tsx` | An existing screen already imports it. It renders `Callout` with `type="error"` |
| Parsed API / form errors | `ErrorMessageBar` | `portal/src/ErrorMessageBar.tsx` | One or more parsed backend or form errors |
| Plan / feature disabled | `FeatureDisabledCallout` | `portal/src/components/v2/FeatureDisabledCallout/FeatureDisabledCallout.tsx` | The plan or feature flag blocks the action |

`Callout` `type` is `"error" | "success" | "warning" | "info"`.

New code imports `FeatureDisabledCallout` directly. `portal/src/graphql/portal/FeatureDisabledMessageBar.tsx` only re-renders that callout for older import sites.

## Rules

- Use callouts for actionable system feedback, not for decorative copy.
- Match `type` to the message: `error` for a failed or blocking state, `info` for guidance.
- Keep message text and links in i18n (`FormattedMessage`) and avoid hard-coded user-facing strings.
- For error collections, use `ErrorMessageBar` so parsed API and form errors render the same way.
- An action inside a callout is a `Link`, `ExternalLink`, or v2 button passed as `text`. There is no message-bar button wrapper.
- Keep messages concise: what happened, and what the user can do next.

## Existing references

- `portal/src/components/v2/Callout/Callout.tsx` is the inline status component.
- `portal/src/ErrorMessageBar.tsx` aggregates parsed errors.
- `portal/src/RedMessageBar.tsx` is the error wrapper around `Callout`.
- `portal/src/components/v2/FeatureDisabledCallout/FeatureDisabledCallout.tsx` is the plan/feature-disabled notice.
