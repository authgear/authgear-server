# Copyable text

Use copyable text patterns for immutable text values shown in lists or read-only display rows.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Text display with copy action | `TextWithCopyButton` | `../../components/common/TextWithCopyButton` | The value is plain display text (not an input field), but users frequently need to copy it |
| Read-only field with copy action | `TextFieldWithCopyButton` | `../../TextFieldWithCopyButton` | The value should be presented in input-field form and copied as a field value |

## Rules

- Use `TextWithCopyButton` for display-only text in tables/lists (for example Group Key, Role Key, resource URI).
- Do not use `TextWithCopyButton` where user editing is expected; use input components instead.
- Use `TextFieldWithCopyButton` when the UI is semantically a form field (read-only input) rather than plain text.
- Copyable values should be stable identifiers/URLs/tokens that users commonly need to reuse.
- Keep surrounding labels/column headers in i18n and avoid hard-coded user-facing strings.
- Avoid showing sensitive secrets in plain copyable text without existing mask/reveal patterns.

## Existing references

- `portal/src/components/common/TextWithCopyButton.tsx` shared text-with-copy component.
- `portal/src/components/roles-and-groups/list/GroupsList.tsx` uses `TextWithCopyButton` for Group Key display.
- `portal/src/components/roles-and-groups/list/RolesList.tsx` uses `TextWithCopyButton` for Role Key display.
- `portal/src/components/api-resources/ResourceList.tsx` uses `TextWithCopyButton` for resource URI display.
- `portal/src/graphql/portal/AdminAPIConfigurationScreen.tsx` uses `TextWithCopyButton` for key ID display.
