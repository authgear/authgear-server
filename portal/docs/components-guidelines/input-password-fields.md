# Password inputs

Use dedicated password input patterns for secrets and credential-like values.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Password entry with reveal behavior | `PasswordField` | `../../PasswordField` | Users enter or update passwords/secrets in forms |
| Basic masked field in existing forms | `TextField` with `type="password"` | `../../TextField` | A password-like field is required in a simple form flow |

## Rules

- Use masked input (`PasswordField` or `type="password"`) for secrets; never render secret input as plain text by default.
- Prefer `PasswordField` when reveal/hide behavior and password UX are needed.
- Keep labels, helper text, and validation messages in i18n.
- Do not auto-fill password-like fields with sensitive values from backend unless product behavior explicitly requires it.
- Separate "set/reset password" actions from generic profile editing to reduce accidental secret exposure.
- Pair password fields with clear validation/error feedback and required-state handling.

## Existing references

- `portal/src/PasswordField.tsx` is the dedicated password input component.
- `portal/src/components/users/ResetPasswordForm.tsx` uses password-specific form behavior.
- `portal/src/graphql/portal/SMTPConfigurationScreen.tsx` uses masked `type="password"` fields for provider credentials.
- `portal/src/graphql/adminapi/AddUserScreen.tsx` and `portal/src/graphql/adminapi/Add2FAScreen.tsx` include password/secret input flows.
