# Modal dialogs

Use modal dialogs for blocking decisions that require explicit user action before continuing.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Confirm / cancel | `ConfirmationDialog` | `portal/src/components/v2/ConfirmationDialog/ConfirmationDialog.tsx` | The user must confirm or cancel before continuing |
| Shared destructive confirmation | `DeleteConfirmationDialog` | `portal/src/components/common/DeleteConfirmationDialog.tsx` | Confirm deletion or removal. It renders `ConfirmationDialog` |
| Navigation blocking confirmation | `BlockerDialog` | `portal/src/BlockerDialog.tsx` | The user may lose unsaved changes |
| Route-leave confirmation | `NavigationBlockerDialog` | `portal/src/NavigationBlockerDialog.tsx` | Browser or in-app navigation would leave an in-progress flow |
| Dialog visibility/loading store | `useConfirmationDialog` | `portal/src/hook/useConfirmationDialog.tsx` | A feature needs lightweight show/dismiss/confirm-loading state |
| One-off dialog that is not confirm/cancel | `Dialog` | `@radix-ui/themes` | The content does not fit `ConfirmationDialog` |

## Rules

- Use dialogs only for blocking interactions. Use `Callout` or field errors for non-blocking feedback.
- Provide a clear title, concise body text, and explicit primary and secondary actions.
- For destructive flows, set `confirmColor="red"` and use action wording (`delete`, `remove`, `confirm`) that matches the consequence.
- Keep all user-facing strings in i18n (`FormattedMessage` / `renderToString`), including title, body, and actions.
- `ConfirmationDialog` is controlled with `open` and `onOpenChange`. Cancel side effects go in `onCancel` or `onOpenChange(false)`. Escape and overlay dismissal call `onOpenChange(false)` and do not call `onCancel`.
- During async confirm flows, pass `loading` so confirm and cancel cannot submit twice.
- Reuse `DeleteConfirmationDialog`, `BlockerDialog`, and `NavigationBlockerDialog` when the pattern matches.
- For route-leave protection, use `NavigationBlockerDialog` so browser navigation and in-app transitions share one dialog.

## Existing references

- `portal/src/components/common/DeleteConfirmationDialog.tsx` wraps `ConfirmationDialog` with `confirmColor="red"`.
- `portal/src/BlockerDialog.tsx` is the unsaved-changes dialog.
- `portal/src/NavigationBlockerDialog.tsx` blocks route changes.
- `portal/src/hook/useConfirmationDialog.tsx` holds visible/loading dialog state.
