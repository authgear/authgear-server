# Copyable text

Use `CopyIconButton` when the user needs to copy a stable value.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Copy an identifier, URL, or token | `CopyIconButton` | `portal/src/components/v2/CopyIconButton/CopyIconButton.tsx` | The value is shown for reuse (client ID, key ID, endpoint, resource URI) |

## Rules

- Pass the string to copy as `textToCopy`.
- Do not use `CopyIconButton` where the user is editing a draft. Use `TextField` inside `FormField` instead.
- Copyable values should be stable identifiers, URLs, or tokens.
- Keep surrounding labels in i18n and avoid hard-coded user-facing strings.
- Do not show a secret in plain copyable text unless the screen already has a mask/reveal pattern.

## Existing references

- `portal/src/graphql/portal/EditOAuthClientForm.tsx` places `CopyIconButton` on read-only OAuth client values.
- `portal/src/graphql/portal/AdminAPIConfigurationScreen.tsx` uses `CopyIconButton` for key ID display.
