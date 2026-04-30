# Modal dialogs

Use modal dialogs for blocking decisions that require explicit user action before continuing.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Generic confirm/cancel modal | `Dialog` (FluentUI) | `@fluentui/react` | A simple one-off dialog with project-specific content |
| Shared destructive confirmation | `DeleteConfirmationDialog` | `../../components/common/DeleteConfirmationDialog` | Confirm deletion/removal with consistent destructive UX |
| Navigation blocking confirmation | `BlockerDialog` / `NavigationBlockerDialog` | `../../BlockerDialog`, `../../NavigationBlockerDialog` | User may lose unsaved changes or leave an in-progress flow |
| Re-auth confirmation dialog | `ReauthDialog` | `../../components/common/ReauthDialog` | Sensitive actions require re-authentication before continuing |
| Dialog visibility/loading store | `useConfirmationDialog` | `../../hook/useConfirmationDialog` | A feature needs lightweight show/dismiss/confirm-loading state |

## Rules

- Use dialogs only for blocking interactions. Use inline message bars or field errors for non-blocking feedback.
- Provide clear title, concise body text, and explicit primary/secondary actions.
- For destructive flows, use destructive theming and action wording (`delete`, `remove`, `confirm`) that matches the consequence.
- Keep all user-facing strings in i18n (`FormattedMessage` / `renderToString`) including title, body, and actions.
- Control visibility explicitly (`hidden` + `onDismiss`) and always define what dismiss means for state cleanup.
- During async confirm flows, disable confirm/cancel actions (or show loading state) to prevent duplicate submissions.
- Reuse shared wrappers (`DeleteConfirmationDialog`, `BlockerDialog`, `ReauthDialog`) when pattern matches instead of re-building dialog skeletons.
- For route-leave protection, prefer `NavigationBlockerDialog` pattern so browser navigation and hash/pivot transitions behave consistently.

## Existing references

- `portal/src/components/common/DeleteConfirmationDialog.tsx` shared destructive confirm dialog.
- `portal/src/BlockerDialog.tsx` shared blocking confirmation layout and actions.
- `portal/src/NavigationBlockerDialog.tsx` route-change blocker with dialog.
- `portal/src/components/common/ReauthDialog.tsx` shared re-auth dialog pattern.
- `portal/src/hook/useConfirmationDialog.tsx` reusable visible/loading dialog state helper.
- `portal/src/components/saml/EditSAMLCertificateForm.tsx` feature dialog for certificate removal.
