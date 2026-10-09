# Button components

Use the shared button wrappers in `portal/src/components/v2` so action semantics and styling stay consistent across screens.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Primary page/form action | `PrimaryButton` | `portal/src/components/v2/Button/PrimaryButton/PrimaryButton.tsx` | Main action on a screen or section (`save`, `create`, `continue`) |
| Secondary action | `SecondaryButton` | `portal/src/components/v2/Button/SecondaryButton/SecondaryButton.tsx` | Non-primary action beside the primary action (`cancel`, `back`) |
| Text-style action | `TextButton` | `portal/src/components/v2/Button/TextButton/TextButton.tsx` | A low-emphasis action in a list or section |
| Light fill on a dark surface | `WhiteButton` | `portal/src/components/v2/Button/WhiteButton/WhiteButton.tsx` | The button sits on a dark island |
| Icon-only action | `IconButton` | `portal/src/components/v2/IconButton/IconButton.tsx` | The action is an icon with an accessible name |
| Button that looks like a link | `LinkButton` | `portal/src/LinkButton.tsx` | A click handler that should look like a link |

`PrimaryButton` and `SecondaryButton` both take `loading?: boolean` and disable the button while `loading` is true.

## Rules

- Import these wrappers. Do not import `Button` from `@radix-ui/themes` in a screen when one of these fits.
- Keep button labels in i18n (`FormattedMessage` / `renderToString`) and avoid hard-coded user-facing text.
- Use exactly one primary action per section or dialog footer. Other actions are `SecondaryButton` or `TextButton`.
- For async submit/confirm flows, pass `loading` so the button disables and a second submit cannot start.
- Destructive confirm uses `ConfirmationDialog` with `confirmColor="red"`, not a separate destructive button component.
- Use `LinkButton` when the interaction is an action callback, not route navigation. For navigation, use `Link` / `ExternalLink`.

## Existing references

- `portal/src/graphql/portal/SMSProviderConfigurationScreen.tsx` uses `PrimaryButton` and `SecondaryButton` together.
- `portal/src/graphql/portal/CreateOAuthClientScreen.tsx` uses `PrimaryButton`.
- `portal/src/components/common/DeleteConfirmationDialog.tsx` confirms deletion through `ConfirmationDialog` with `confirmColor="red"`.
