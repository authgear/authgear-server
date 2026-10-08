# Radio button

Use a single-choice control when the options are mutually exclusive.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Card with title and subtitle | `RadioCards` | `portal/src/components/v2/RadioCards/RadioCards.tsx` | Exactly one option is selected and a card presentation fits |
| Card with an icon | `IconRadioCards` | `portal/src/components/v2/IconRadioCards/IconRadioCards.tsx` | Exactly one option is selected and each option has an icon |
| Compact single choice | `RadioGroup` | `@radix-ui/themes` | Exactly one option is selected and a compact list fits |

## Rules

- For single-choice options, use `RadioCards`, `IconRadioCards`, or `RadioGroup`. Do not fake single-choice with a checkbox group.
- Keep option labels in i18n (`renderToString` / `FormattedMessage`) and avoid hard-coded text.
- Keep `onChange` / `onValueChange` focused on the selected option.

## Existing references

- `portal/src/graphql/portal/LoginMethodConfigurationScreen.tsx` uses `IconRadioCards`.
- `portal/src/graphql/portal/AnonymousUsersConfigurationScreen.tsx` uses `RadioGroup` for the promotion-conflict choice.
- `portal/src/components/v2/RadioCards/RadioCards.stories.tsx` demonstrates card-style single-choice selection.
