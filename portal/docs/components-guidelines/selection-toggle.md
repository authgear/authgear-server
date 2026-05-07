# Toggle controls

Use `Toggle` for boolean feature switches (enable/disable a behavior).

## Rules

- `Toggle` is for boolean state only. If users choose one option from multiple choices, use radio button components (`ChoiceGroup` / `RadioCards`) instead.
- Multiple toggles can appear in one section when each maps to a different boolean config.
- For dependent toggles, encode parent-child behavior explicitly:
  - If child setting is not meaningful when parent is off, hide or disable the child toggle.
  - Prefer `disabled` with clear label/description when users should still see the option.
- Keep labels in i18n (`renderToString`/`FormattedMessage`) and avoid hard-coded text.
- `onChange` handlers should update only the corresponding boolean field and avoid unrelated side effects.

## Existing references

- `portal/src/graphql/portal/LoginMethodConfigurationScreen.tsx`:
  - Email verification uses related toggles (`required` and `allowed`), where `allowed` is disabled when `required` is true.
  - Passkey section uses two independent toggles (`passkeyChecked` and `passkeyShowDoNotAskAgain`).
- `portal/src/graphql/portal/BotProtectionConfigurationScreen.tsx` uses one toggle (`enabled`) and conditionally renders dependent settings when enabled.
