# Checkbox controls

Use checkbox controls for independent multi-select flags.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Standard checkbox flag | `Checkbox` | `@radix-ui/themes` | Multiple options can be enabled at the same time |
| Checkbox with inline help | `CheckboxWithTooltip` | `portal/src/CheckboxWithTooltip.tsx` | The option needs a tooltip explanation |
| Checkbox with subordinate content | `CheckboxWithContentLayout` | `portal/src/CheckboxWithContentLayout.tsx` | The option controls content shown below (for example tag pickers) |

## Rules

- Use `Checkbox` for independent multi-select flags (users can check multiple options at once).
- For checkbox options that need inline explanation, use `CheckboxWithTooltip`. It composes Radix `Checkbox` and v2 `Tooltip`.
- For checkbox options that control extra content beneath them, use `CheckboxWithContentLayout`.
- If checkbox options are mutually exclusive, disable or hide the ones that cannot be combined, or switch to `RadioCards`, `IconRadioCards`, or Radix `RadioGroup`.
- Keep labels and tooltip messages in i18n.

## Existing references

- `portal/src/graphql/portal/LoginMethodConfigurationScreen.tsx` uses `CheckboxWithTooltip` and `CheckboxWithContentLayout` for email and username settings.
