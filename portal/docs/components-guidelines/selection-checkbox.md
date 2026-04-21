# Checkbox controls

Use checkbox controls for independent multi-select flags.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Standard checkbox flag | `Checkbox` | `@fluentui/react` | Multiple options can be enabled at the same time |
| Checkbox with inline help | `CheckboxWithTooltip` | `../../CheckboxWithTooltip` | The option needs tooltip explanation |
| Checkbox with subordinate content | `CheckboxWithContentLayout` | `../../CheckboxWithContentLayout` | The option controls content shown below (for example tag pickers) |

## Rules

- Use `Checkbox` for independent multi-select flags (users can check multiple options at once).
- For checkbox options that need inline explanation/help, use `CheckboxWithTooltip`.
- For checkbox options that control extra content beneath them, use `CheckboxWithContentLayout`.
- If checkbox options are mutually exclusive, enforce the relationship with `disabled`/`hidden`, or switch to radio if the intent is single-choice.
- Keep labels and tooltip messages in i18n.

## Existing references

- `portal/src/graphql/portal/LoginMethodConfigurationScreen.tsx` email/username settings use multiple checkboxes, with tooltip and content-layout variants.
