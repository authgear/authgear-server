# Dropdown inputs

Use dropdown inputs when users select from predefined options instead of entering free text.

## Which component to use

| Pattern | Component | Import path | Use when |
|---|---|---|---|
| Standard form selection | `Dropdown` / `FormDropdown` | `@fluentui/react`, `../../FormDropdown` | A single option is selected from a known list in forms |
| Searchable large option set | `SearchableDropdown` | `../../components/common/SearchableDropdown` | Option lists are long and users need search/filter |
| Filter-style toolbar dropdown | `CommandBarDropdown` / specialized filter dropdowns | `../../CommandBarDropdown`, `../../components/**/**FilterDropdown` | Selection is used for list filtering or command-bar style controls |

## Rules

- Use dropdowns only for predefined option sets; use text inputs when free-form values are expected.
- Keep option labels in i18n (`renderToString`/`FormattedMessage`) and avoid hard-coded text.
- Provide `selectedKey` and `onChange` as a controlled input; avoid uncontrolled state drift.
- Use `SearchableDropdown` when option count is large enough that scan-only selection is hard.
- For required fields, provide clear label and default/placeholder behavior that avoids ambiguous empty state.
- Avoid packing unrelated actions into dropdown options; use action buttons when behavior is not selection.

## Existing references

- `portal/src/graphql/portal/LoginMethodConfigurationScreen.tsx` uses `Dropdown` in verification-related settings.
- `portal/src/graphql/portal/LanguagesConfigurationScreen.tsx` and `portal/src/graphql/portal/ManageLanguageWidget.tsx` use dropdown-based language selection.
- `portal/src/components/common/SearchableDropdown.tsx` provides searchable dropdown behavior for larger lists.
