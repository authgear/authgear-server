# Decouple translations from the Authflow API implementation plan

Spec: [fill_form_data](../../specs/authentication-flow-api-reference.md#fill_form_data) and [Profile Filling: Translations](../../specs/user-profile/profile-filling.md#translations).

Builds on [Per-form field label](./2026-11-12-per-form-field-label.md), which is implemented first.

## 1. Goal / scope

Stop returning localized labels from the Authflow API: remove `fields[].label` and
`options[].label` from `fill_form_data`. A Custom UI uses the form's `name`, the field's `key`, and
the option's `value` to show its own text, as it already does for the title and description. Auth UI
resolves the same labels itself, so what the end-user sees does not change.

Auth UI resolves a field's label in the order the spec defines:

1. `v2.page.fill-form.{name}.field-label-{key}` — only when the form has a `name`.
2. `standard-attribute-label-{pointer}` or `custom-attribute-label-{pointer}`.
3. `labelutil.Label` of the first pointer token.

Option labels keep today's keys: `language-{value}` for `/locale`, otherwise
`custom-attribute-enum-label-{pointer}-{value}`, falling back to `value` and `labelutil.Label(value)`
respectively.

In scope: Authflow data, Auth UI view model, unit tests, e2e tests.

Out of scope: `constraints.pattern_message` (stays in the API; it is an error message, not a label),
the settings page, portal UI.

## 2. Config model and schema

No config change. One existing helper is exported so Auth UI can pick the attribute label key
without the app config:

- `pkg/lib/config/authentication_flow_form.go`: rename `isStandardAttributePointer` to
  `IsStandardAttributePointer`, and update its caller `isFillableUserProfilePointer` in the same file.

## 3. Runtime flow

### 3.1 Auth UI: resolve labels

Both Auth UI pages build forms through `NewFormViewModel` in
`pkg/auth/handler/webapp/authflowv2/form.go`:

- `AuthflowV2FillFormHandler.GetData` (`fill_form.go`) → `NewFormViewModel(screenData.Form, &baseViewModel)`
- `IdentifyFormViewModels` (`form.go`, identify step forms) → `NewFormViewModel(*option.Form, base)`

`base.Translations` (`viewmodels.TranslationsCompat`, with `HasKey(key)` and `RenderText(key, args)`)
is set by `BaseViewModeler.ViewModelForAuthFlow`, so no handler or wiring changes.

New file `pkg/auth/handler/webapp/authflowv2/form_label.go`:

```go
// formFieldLabel resolves the label of the field with key in the form named formName.
func formFieldLabel(t viewmodels.TranslationsCompat, formName string, key string) (string, error)

// formFieldOptions resolves the label of each option of the field with key.
func formFieldOptions(t viewmodels.TranslationsCompat, key string, options []declarative.FillFormDataFieldOption) ([]FormFieldOptionViewModel, error)

func translatedOrDefault(t viewmodels.TranslationsCompat, key string, defaultText string) (string, error)

func firstPointerToken(pointer string) string
```

`formFieldLabel` call flow:

1. If `formName != ""`: `k := "v2.page.fill-form." + formName + ".field-label-" + key`.
   `t.HasKey(k)`: error → return it; `true` → return `t.RenderText(k, nil)`; `false` → continue.
2. The only field kind is `user_profile`, whose key is its pointer. Prefix is
   `standard-attribute-label-` when `config.IsStandardAttributePointer(key)`, else
   `custom-attribute-label-`. Return
   `translatedOrDefault(t, prefix+key, labelutil.Label(firstPointerToken(key)))`.

`formFieldOptions` call flow, once per option:

- `key == "/locale"`: label is `translatedOrDefault(t, "language-"+value, value)`.
- otherwise: label is `translatedOrDefault(t, "custom-attribute-enum-label-"+key+"-"+value, labelutil.Label(value))`.
- Returns `nil` for `nil` options. Any error is returned.

`firstPointerToken` is copied from `data_fill_form.go` in this step and deleted there in 3.2.

Changes in `form.go`:

- New type, so `__form_fields.html` keeps reading `.Value` and `.Label` unchanged:

  ```go
  type FormFieldOptionViewModel struct {
  	Value string
  	Label string
  }
  ```

- `FormFieldViewModel.Options` becomes `[]FormFieldOptionViewModel`.
- In the `NewFormViewModel` field loop, before building `fvm`:
  `label, err := formFieldLabel(base.Translations, form.Name, f.Key)` and
  `options, err := formFieldOptions(base.Translations, f.Key, f.Options)`; return any error.
  Set `Label: label`, `Options: options`.
- Boolean: `fvm.LabelHTML = htmlutil.SanitizeLinksHTML(label)`. Sanitizing moves here from the
  Authflow API, so the `// nolint:gosec` cast goes, and the `LabelHTML` comment "The Authflow API has
  sanitized it" becomes "Links are kept; the rest is escaped."

Templates are unchanged: `__form_fields.html` already renders `$field.Label`, `$field.LabelHTML`,
and `.Label` of each option.

### 3.2 Authflow data: remove labels

File: `pkg/lib/authenticationflow/declarative/data_fill_form.go`.

- `FillFormDataField`: delete `Label`.
- `FillFormDataFieldOption`: delete `Label`; it keeps `Value string \`json:"value"\``.
- `NewFillFormDataForm`: delete the `formFieldLabel` call and the boolean `SanitizeLinksHTML` block.
  The loop no longer needs a separate field key variable for labels; `Key` is still set from
  `formFieldKey(f)`.
- Delete `formFieldLabel` (including its per-form lookup) and `firstPointerToken`.
- `formFieldOptions` builds `FillFormDataFieldOption{Value: value}` for custom `enum` values and
  `FillFormDataFieldOption{Value: tag}` for `/locale`. It no longer takes `ctx` or returns an error:

  ```go
  func formFieldOptions(deps *authflow.Dependencies, pointer string, custom *config.CustomAttributesAttributeConfig) []FillFormDataFieldOption
  ```

- `translatedOrDefault` stays; `formFieldConstraints` still uses it for `pattern_message`.
- Drop the now-unused `htmlutil` and `labelutil` imports.

## 4. Compatibility and deployment behavior

- Authflow API: `fill_form_data.form.fields[].label` and `options[].label` are removed. This shipped
  in `2026-11-04.0`, but no Custom UI uses it yet, so there is no `docs/BREAKING-CHANGES.md` entry.
  Mention the removal in the PR description.
- Auth UI shows the same labels as before, including per-form labels.
- `FillFormData` is computed per response, not persisted, so no stored state changes. Auth UI reads
  the data from the same process that builds it, so mixed versions during a deploy do not apply.

## 5. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/config/authentication_flow_form.go` | Export `IsStandardAttributePointer`. |
| `pkg/auth/handler/webapp/authflowv2/form_label.go` | New, per 3.1. |
| `pkg/auth/handler/webapp/authflowv2/form_label_test.go` | New unit tests. |
| `pkg/auth/handler/webapp/authflowv2/form.go` | `FormFieldOptionViewModel`, label and option resolution, boolean sanitizing. |
| `pkg/auth/handler/webapp/authflowv2/form_test.go` | `TestNewFormViewModel` passes a fake `Translations`; add label assertions. |
| `pkg/lib/authenticationflow/declarative/data_fill_form.go` | Remove labels, per 3.2. |
| `pkg/lib/authenticationflow/declarative/data_fill_form_test.go` | Remove label cases; options carry values only. |
| `e2e/tests/profile_filling/*.test.yaml` | Remove `"label"` from expected outputs. |
| `e2e/tests/profile_filling/per_form_field_label.test.yaml` | Deleted; replaced by the webapp test below. |
| `e2e/tests/profile_filling/var_translations/templates/en/translation.json` | Remove `v2.page.fill-form.consent.field-label-/x_newsletter`. |
| `e2e/tests/webapp/fill_form/per_form_field_label.test.yaml` | New. |
| `e2e/tests/webapp/fill_form/var_labels/templates/en/translation.json` | New. |

## 6. Test plan

### Unit tests (Convey, matching both packages)

`pkg/auth/handler/webapp/authflowv2/form_label_test.go` (new), with a hand-written fake
`TranslationsCompat` backed by a `map[string]string`. The cases are moved from
`data_fill_form_test.go`:

- `formFieldLabel`:
  - per-form key wins over the attribute label (`/given_name` in form `names`);
  - per-form key of another form name is ignored, and an unnamed form never reads it;
  - standard attribute uses `standard-attribute-label-`, custom uses `custom-attribute-label-`;
  - no translation falls back to `labelutil.Label` (`/x_employee_id` → "X Employee Id").
- `formFieldOptions`:
  - `/locale` uses `language-{value}`, falling back to the tag;
  - custom enum uses `custom-attribute-enum-label-{key}-{value}`, falling back to `labelutil.Label`;
  - `nil` options give `nil`.

`pkg/auth/handler/webapp/authflowv2/form_test.go`, `TestNewFormViewModel`:

- Every `viewmodels.BaseViewModel{}` gets `Translations: fakeTranslations{}`, or the new lookups
  call a nil interface.
- New case: a boolean field whose label contains `onclick` and `<script>`; `LabelHTML` is sanitized
  as `htmlutil.SanitizeLinksHTML` does.

`pkg/lib/authenticationflow/declarative/data_fill_form_test.go`:

- Delete every label case, including the per-form ones.
- Enum and `/locale` option cases assert `[]FillFormDataFieldOption{{Value: ...}}` only.
- `fakeTranslationService` stays for the `pattern_message` cases.

Run: `go test ./pkg/lib/config/... ./pkg/lib/authenticationflow/declarative/... ./pkg/auth/handler/webapp/authflowv2/...`.

### E2E

Delete the `"label": ...` lines from the expected outputs of every `e2e/tests/profile_filling` test
that has them (`grep -rln '"label"' e2e/tests/profile_filling`), and change
`{ "value": ..., "label": ... }` options to `{ "value": ... }`.

The per-form label can no longer be checked through the Authflow API, so
`e2e/tests/profile_filling/per_form_field_label.test.yaml` moves to Auth UI as
`e2e/tests/webapp/fill_form/per_form_field_label.test.yaml`, modelled on
`signup_fill_form.test.yaml`, with `extra_files_directory: ./var_labels`. Same config: custom
attributes `/x_terms` and `/x_newsletter` (boolean); signup with `identify` (username) and
`create_authenticator` (password), then:

1. `fill_form` with `form.name: consent`, fields `/x_terms` (required) and `/x_newsletter`.
2. `fill_form` without `name`, field `/x_newsletter`.

`var_labels/templates/en/translation.json`:

```json
{
  "custom-attribute-label-/x_newsletter": "Newsletter",
  "v2.page.fill-form.consent.field-label-/x_newsletter": "Send me news. See our <a href=\"https://example.com/privacy\">Privacy Policy</a>."
}
```

Cases (`html_text_contains`):

- Page 1 contains `Send me news. See our`, `href="https://example.com/privacy"`, and `X Terms`
  (fallback for `/x_terms`).
- Page 2 contains `Newsletter` and not the per-form text, proving the override is per form.
- Submitting both pages reaches the end of the flow.

Run per the `write-e2e-test` skill (unset `HTTP(S)_PROXY`).

## 7. Fixed behavioral decisions

- The Authflow API returns no field or option labels. `pattern_message` stays.
- Options stay objects, `{ "value" }`, so fields can be added later without changing the shape.
- Auth UI's label lookup is the one the spec defines; this change moves it, not alters it.

## 8. Atomic commit plan

1. **`[Profile Filling] Resolve form labels in Auth UI`**
   - `pkg/lib/config/authentication_flow_form.go`
   - `pkg/auth/handler/webapp/authflowv2/form_label.go`, `form_label_test.go`
   - `pkg/auth/handler/webapp/authflowv2/form.go`, `form_test.go`
   - Auth UI resolves field and option labels itself and ignores the API labels. The API still
     returns them, so no output changes. No wiring or generated files.
2. **`[Profile Filling] Remove labels from fill_form_data`**
   - `pkg/lib/authenticationflow/declarative/data_fill_form.go`, `data_fill_form_test.go`
   - `e2e/tests/profile_filling/*.test.yaml` expectations; delete
     `e2e/tests/profile_filling/per_form_field_label.test.yaml` and its translation key.
   - `e2e/tests/webapp/fill_form/per_form_field_label.test.yaml`,
     `e2e/tests/webapp/fill_form/var_labels/templates/en/translation.json`
   - API output change, with e2e moved in the same commit so the suite stays green.
