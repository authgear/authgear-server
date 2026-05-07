# Input fields

Use dedicated text-input components for editable/read-only text values.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Basic editable input (no helper text) | `FormTextField` / `TextField` | `../../FormTextField`, `../../TextField` | Standard editable text input with label only |
| Editable input with helper text | `FormTextField` / `TextField` + `WidgetDescription` (preferred) or `description` prop | `../../FormTextField`, `../../TextField`, `../../Widget` | Input needs guidance/constraints/examples |
| Read-only input with built-in copy action | `TextFieldWithCopyButton` | `../../TextFieldWithCopyButton` | Value is backend-returned, immutable in this screen, and expected to be copied (for example Project ID, endpoint URL, client ID) |

## Rules

- For editable text inputs, use `FormTextField`/`TextField`; add helper text only when it helps users make a correct input.
- Prefer helper text rendered as `WidgetDescription` under the field. Use `description` prop when the helper text must stay tightly coupled with the field itself.
- For read-only copyable values, always use `TextFieldWithCopyButton` with `readOnly={true}`.
- `TextFieldWithCopyButton` is only for backend-returned, non-editable values in current flow. Do not use it for draft/user-entered/editable inputs.
- If the value is secret and needs reveal/mask behavior, follow existing secret-field patterns instead of a plain read-only copy field.
- Keep labels and helper text in i18n (`renderToString`/`FormattedMessage`) and avoid hard-coded text.
- `onChange` handlers should update only the corresponding field and avoid unrelated side effects.

## Existing references

- Basic editable input without helper text:
  - `portal/src/graphql/adminapi/UserProfileForm.tsx` standard attribute text fields (for example `given_name`).
- Editable input with helper text below field:
  - `portal/src/graphql/adminapi/AddRoleScreen.tsx` role name and role key with `WidgetDescription`.
- Read-only copyable input:
  - `portal/src/graphql/portal/AdminAPIConfigurationScreen.tsx` Admin API endpoint and Project ID.
  - `portal/src/graphql/portal/EditOAuthClientForm.tsx` read-only OAuth client values.
