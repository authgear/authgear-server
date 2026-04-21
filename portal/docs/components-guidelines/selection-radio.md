# Radio button

Use radio-button-style controls for single-choice settings so users can clearly see mutually exclusive options.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Single choice in FluentUI-based forms | `ChoiceGroup` | `@fluentui/react` | Exactly one option can be selected at a time |
| Single choice in v2 card-style UI | `RadioCards` | `../../components/v2/RadioCards/RadioCards` | Exactly one option can be selected at a time and card presentation fits the screen |

## Rules

- For single-choice options, use a radio button component (`ChoiceGroup` / `RadioCards`) instead of checkbox groups, custom toggles, or ad-hoc clickable rows.
- Keep option labels in i18n (`renderToString`/`FormattedMessage`) and avoid hard-coded text.
- Use `ChoiceGroup` for FluentUI-based form screens; use `RadioCards` when card-style visual selection better matches the UI.
- Keep `onChange` focused on updating the selected option only, and avoid unrelated side effects.

## Existing references

- `portal/src/graphql/portal/EndpointDirectAccessScreen.tsx` uses `ChoiceGroup` for mutually exclusive direct-access behavior options.
- `portal/src/graphql/portal/CreateOAuthClientScreen.tsx` uses `ChoiceGroup` for single-choice configuration in FluentUI-based forms.
- `portal/src/components/v2/RadioCards/RadioCards.stories.tsx` demonstrates card-style single-choice selection via `RadioCards`.
