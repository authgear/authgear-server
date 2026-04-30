# Message bars

Use message bars for inline status, warning, and error feedback that must be visible in-page.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| API/form error aggregation | `ErrorMessageBar` | `../../ErrorMessageBar` | Show one or more parsed backend/form errors consistently |
| Informational notice (blue style) | `BlueMessageBar` | `../../BlueMessageBar` | Non-blocking informational messages or guidance |
| Error/warning notice (red style) | `RedMessageBar` and specialized wrappers | `../../RedMessageBar` | Configuration blockers, warnings, or error-level reminders |
| Message bar action button | `MessageBarButton` | `../../MessageBarButton` | In-message action links/buttons with consistent FluentUI behavior |
| Feature gating notice | `FeatureDisabledMessageBar` | `./FeatureDisabledMessageBar` | Plan/feature-disabled messaging with built-in routing/link values |

## Rules

- Use message bars for actionable system feedback; avoid using them for decorative copy.
- Match severity to message intent (`error` for blocking/failed states, info-style for guidance).
- Keep message text and links in i18n (`FormattedMessage`) and avoid hard-coded user-facing strings.
- For error collections, prefer `ErrorMessageBar` so parsed API/form errors render consistently.
- Use `MessageBarButton` for in-bar actions instead of custom button/link variants.
- When links are included in message text, use project wrappers (`ReactRouterLink`/`ExternalLink`) to preserve style and routing behavior.
- Keep messages concise and task-oriented (what happened + what user can do next).

## Existing references

- `portal/src/ErrorMessageBar.tsx` central error aggregation and rich i18n link rendering.
- `portal/src/BlueMessageBar.tsx` informational message bar styling wrapper.
- `portal/src/RedMessageBar.tsx` error/warning message bar styling wrapper and reminder variants.
- `portal/src/MessageBarButton.tsx` standardized message bar action button wrapper.
- `portal/src/graphql/portal/FeatureDisabledMessageBar.tsx` plan/feature-disabled informational message bar.
