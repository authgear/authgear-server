# Input fields

Use the v2 field components for editable text and for copying an immutable value.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Label, hint, and error around a control | `FormField` | `portal/src/components/v2/FormField/FormField.tsx` | The control needs a label, optional hint, or a validation error |
| Editable text | `TextField` | `portal/src/components/v2/TextField/TextField.tsx` | The user types or edits a value |
| Copy an immutable value | `CopyIconButton` | `portal/src/components/v2/CopyIconButton/CopyIconButton.tsx` | The value is backend-returned and users need to copy it (Project ID, endpoint URL, client ID) |

## Rules

- Put `TextField` inside `FormField` when the field has a label, hint, or error. Add hint text only when it helps the user enter a correct value.
- `CopyIconButton` is for a stable identifier, URL, or token. Do not use it on a draft the user is still editing.
- If the value is a secret and needs reveal/mask behavior, use `PasswordField` (`portal/src/PasswordField.tsx`). `TextField` accepts `type="password"` for a simple masked field. See `input-password-fields.md`.
- Keep labels and helper text in i18n (`renderToString` / `FormattedMessage`) and avoid hard-coded text.
- `onChange` handlers should update only the corresponding field and avoid unrelated side effects.

## Existing references

- `portal/src/graphql/adminapi/UserProfileForm.tsx` uses v2 `TextField` and `FormField` for standard attribute fields.
- `portal/src/graphql/portal/EditOAuthClientForm.tsx` uses v2 `TextField` with `CopyIconButton` for read-only client values.
- `portal/src/graphql/portal/AdminAPIConfigurationScreen.tsx` uses `CopyIconButton` for the key ID and other copyable values.
