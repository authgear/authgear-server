# List add actions

Use `TextFieldList` for an editable list of strings. Use a v2 button when the user appends some other kind of item.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Editable list of strings | `TextFieldList` | `portal/src/components/v2/TextFieldList/TextFieldList.tsx` | The list items are strings the user adds, edits, and deletes |
| Any other append action | `SecondaryButton` or `TextButton` | `portal/src/components/v2/Button/SecondaryButton/SecondaryButton.tsx`, `portal/src/components/v2/Button/TextButton/TextButton.tsx` | The user appends an item that is not a plain string field |

`TextFieldList` already renders a `SecondaryButton` for add. It takes `onListItemAdd`, `onListItemChange`, `onListItemDelete`, and `addButtonLabelMessageID`.

## Rules

- Keep add labels in i18n and avoid hard-coded user-facing strings.
- Disable add when a limit is reached or an async operation is in progress.
- Apply domain limits in UI state (for example max item count by feature config or plan), not only in backend validation.
- Use `Add ...` for user-provided entries and `Generate ...` for system-generated entries (keys, certificates, secrets).
- Keep the click handler scoped to the collection update.

## Existing references

- `portal/src/graphql/portal/EditOAuthClientForm.tsx` uses `TextFieldList` for redirect URI lists, including one inside an `Accordion`.
