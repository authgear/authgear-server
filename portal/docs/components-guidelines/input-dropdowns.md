# Dropdown inputs

Use a Radix select when the user picks one option from a known list. There is no shared dropdown wrapper.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Single choice from a short list | `Select` | `@radix-ui/themes` | One option is selected from a predefined list in a form |
| Row or toolbar action menu | `DropdownMenu` | `@radix-ui/themes` | The control opens actions, not a form value |

## Rules

- Use `Select` for a predefined option set. Use `TextField` when the value is free-form.
- Keep option labels in i18n (`renderToString` / `FormattedMessage`) and avoid hard-coded text.
- Control `Select` with `value` and `onValueChange`.
- For a required field, give it a label and a placeholder that does not look like a real selection.
- Do not put unrelated actions in `Select` options. Use `DropdownMenu` or a button when the behavior is an action.

## Existing references

- `portal/src/graphql/portal/CustomTextConfigurationScreen.tsx` uses `Select.Root` for a single choice.
- `portal/src/UserProfileAttributesList.tsx` uses `DropdownMenu` for a row action menu.
