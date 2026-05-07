# Expandable sections (Advanced)

Use existing expandable patterns for settings that reveal extra controls on demand. Do not introduce one-off expand/collapse triggers when an existing pattern fits.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Parent-controlled expand/collapse state | `FoldableDiv` | `../../FoldableDiv` (or relative path to `portal/src/FoldableDiv.tsx`) | The expanded state needs to be controlled by parent state, reset logic, or cross-field interactions |
| Self-contained expand/collapse state | `Accordion` | `../../components/common/Accordion` | The expanded state is local to one section and does not need parent coordination |

## UX and content rules

- Use this pattern for "Advanced", "Optional", or secondary settings that should be hidden by default.
- The toggle label must come from i18n (for example via `FormattedMessage`), not hard-coded text.
- If the expanded content includes links inside `WidgetDescription`/FluentUI `Text`, follow the link rules above (`Link`/`ExternalLink` from `portal/src`).

## Existing references

- `portal/src/graphql/adminapi/AddUserScreen.tsx` uses `FoldableDiv` for the "Advanced" section.
- `portal/src/graphql/portal/SingleSignOnConfigurationWidget.tsx` uses `FoldableDiv` for advanced settings.
- `portal/src/graphql/portal/EditOAuthClientForm.tsx` uses `Accordion` for local, self-contained expandable content.
