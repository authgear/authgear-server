# Button components

Use the shared button wrappers in `portal/src` so action semantics and styling stay consistent across screens.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Primary page/form action | `PrimaryButton` | `../../PrimaryButton` | Main action on a screen or section (`save`, `create`, `continue`) |
| Secondary action | `DefaultButton` | `../../DefaultButton` | Non-primary action beside primary action (`cancel`, `back`, optional actions) |
| Inline lightweight action | `ActionButton` | `../../ActionButton` | Text-like actions in lists/widgets (`add`, `remove`, `edit`) |
| Link-styled button behavior | `LinkButton` | `../../LinkButton` | Action behaves like button but should look like a link |
| Async primary action with spinner | `ButtonWithLoading` | `../../ButtonWithLoading` | Confirm/save action needs loading state and duplicate-submit protection |
| Message bar action | `MessageBarButton` | `../../MessageBarButton` | Action button appears inside a message bar |
| Command bar action | `CommandBarButton` | `../../CommandBarButton` | Action is rendered in command bar / toolbar context |
| Menu-trigger button action | `PrimaryButton` / `DefaultButton` with `menuProps` | `../../PrimaryButton`, `../../DefaultButton` | One trigger needs multiple related actions (for example add options, row action menu) |
| Outlined emphasis action | `OutlinedActionButton` | `../../components/common/OutlinedActionButton` | Needs outlined visual emphasis while remaining non-primary |

## Rules

- Prefer project wrappers (`PrimaryButton`, `DefaultButton`, `ActionButton`, etc.) over importing FluentUI button primitives directly in portal screens.
- Keep button labels in i18n (`FormattedMessage` / `renderToString`) and avoid hard-coded user-facing text.
- Use exactly one primary action per section/dialog footer whenever possible; other actions should be secondary/link-style.
- For async submit/confirm flows, use `ButtonWithLoading` or explicitly disable actions during loading to prevent duplicate submissions.
- Use destructive theme only for destructive intent (delete/remove/revoke), and pair it with clear wording.
- Use `ActionButton` for collection/list row actions (`+ Add`, `Generate`, row-level delete/edit), not for the main page submit action.
- Use `LinkButton` when the interaction is an action callback, not route navigation. For navigation, use `Link` / `ExternalLink`.
- In message bars and command bars, use their dedicated wrappers (`MessageBarButton`, `CommandBarButton`) for consistent rendering.
- Use `menuProps` when one button needs to expose multiple related actions:
- Use `PrimaryButton` + `menuProps` for additive/create actions with variants (for example add different authenticator types).
- Use `DefaultButton` + `menuProps` for row/item action menus (`action`, `edit`, `remove`, `verify`) where a neutral trigger is preferred.
- Use `OutlinedActionButton` for high-visibility but non-primary state-management actions (for example account status operations), and use `iconProps` when icon + label improves quick recognition (`disable`, `enable`, `set/edit period`).

## Existing references

- `portal/src/graphql/portal/EndpointDirectAccessScreen.tsx` uses `PrimaryButton` for save flow.
- `portal/src/components/common/DeleteConfirmationDialog.tsx` uses `PrimaryButton` + `DefaultButton` as standard dialog footer actions.
- `portal/src/components/saml/EditSAMLCertificateForm.tsx` uses `ActionButton` (`+ Generate certificate`) and `ButtonWithLoading` for confirm.
- `portal/src/FieldList.tsx` uses `ActionButton` for list add action pattern.
- `portal/src/ShowError.tsx` and `portal/src/ModifiedIndicator.tsx` use `MessageBarButton` in message-bar actions.
- `portal/src/CommandBarPrimaryButton.tsx` and command bar screens demonstrate command-bar button rendering.
- `portal/src/graphql/adminapi/UserDetailsAccountSecurity.tsx` uses `PrimaryButton` for `Add Password` and `PrimaryButton` + `menuProps` for `Add 2FA`.
- `portal/src/graphql/adminapi/UserDetailsConnectedIdentities.tsx` uses `DefaultButton` + `menuProps` for identity row `Action` menus.
- `portal/src/graphql/adminapi/UserDetailsAccountStatus.tsx` uses `OutlinedActionButton` with `iconProps` for `disable user`, `set/edit valid period`, `anonymize`, and `remove` account-status actions.
