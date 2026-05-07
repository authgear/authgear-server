# List add/generate actions

Use a consistent add-action pattern when users append items to a list (for example `+ Add` and `+ Generate certificate`).

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Add new list item | `ActionButton` (plus icon) | `../../ActionButton` | User appends a new item to a dynamic list |
| Structured list add/remove flow | `FieldList` | `../../FieldList` | List supports add/edit/delete with form binding and validation |
| Domain-specific generate action | `ActionButton` (plus icon) | `../../ActionButton` | User generates a new item (for example certificate/key) into a collection |

## Rules

- Use `ActionButton` with `CirclePlus` icon for collection-append actions to keep visual consistency across screens.
- Keep add/generate labels in i18n (`add`, `generate certificate`, etc.) and avoid hard-coded user-facing strings.
- Disable add/generate when limits are reached or async operation is in progress.
- If add behavior is a form-list pattern, prefer `FieldList` and configure `makeDefaultItem`, `onListItemAdd`, and `addDisabled` instead of ad-hoc list mutation code.
- Apply domain limits in UI state (for example max item count by feature config/plan), not just backend validation.
- Use clear action wording:
- `Add ...` for user-provided draft entries.
- `Generate ...` for system-generated entries (keys/certificates/secrets).
- Keep add/generate side effects scoped to the collection update; avoid unrelated state mutation in click handlers.

## Existing references

- `portal/src/FieldList.tsx` shared list component with built-in add/delete actions.
- `portal/src/graphql/portal/HookConfigurationScreen.tsx` uses `FieldList` for `+ Add` in blocking/non-blocking hooks.
- `portal/src/components/saml/EditSAMLCertificateForm.tsx` uses `ActionButton` for `+ Generate certificate`.
- `portal/src/components/applications/OAuthClientSAMLForm.tsx` uses `FieldList` for certificate list add behavior.
