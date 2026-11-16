# Remove pattern_message from fill_form_data implementation plan

Spec: [fill_form_data](../../specs/authentication-flow-api-reference.md#fill_form_data) and [Profile Filling: Translations](../../specs/user-profile/profile-filling.md#translations).

Follows [Decouple translations from the Authflow API](./2026-11-12-decouple-translations.md), which
left `constraints.pattern_message` in the API. This plan removes it, so the Authflow API renders no
translation at all.

## 1. Goal / scope

Remove `fill_form_data.form.fields[].constraints.pattern_message`. `pattern` stays: it is the
constraint, not text. A Custom UI shows its own message when `pattern` fails.

Auth UI does not change. It never read `pattern_message`: `authflowv2/__form_fields.html:205`
passes `"PatternPointer" $field.Key` to `authflowv2/__error.html`, which looks up
`custom-attribute-pattern-error-{pointer}`, then `standard-attribute-pattern-error-{pointer}`, then
`v2.error.pattern`.

With `pattern_message` gone, `fill_form_data` is the last user of the Authflow translation
dependency, so that dependency is removed too.

In scope: Authflow data, Authflow dependencies and their wiring, unit tests, e2e tests.

Out of scope: Auth UI templates, the settings page, the translation keys themselves
([String constraints](../../specs/user-profile/design.md#string-constraints)).

## 2. Config model and schema

None.

## 3. Runtime flow

### 3.1 Authflow data

File: `pkg/lib/authenticationflow/declarative/data_fill_form.go`.

- `FillFormDataFieldConstraints`: delete `PatternMessage`.
- `(*FillFormDataFieldConstraints).isEmpty`: delete the `c.PatternMessage == ""` term.
- `formFieldConstraints` no longer looks up a translation, so it takes no `ctx` and returns no
  error. Delete `patternMessageKey` and the `translatedOrDefault` call:

  ```go
  func formFieldConstraints(deps *authflow.Dependencies, pointer string, custom *config.CustomAttributesAttributeConfig) *FillFormDataFieldConstraints
  ```

- Delete `translatedOrDefault`; it has no other caller.
- `NewFillFormDataForm` has nothing left that uses `ctx` or fails:

  ```go
  func NewFillFormDataForm(deps *authflow.Dependencies, form *config.AuthenticationFlowForm, user *model.User) FillFormDataForm
  ```

  The loop sets `Constraints: formFieldConstraints(deps, pointer, custom)` in the struct literal.
  Drop the `context` import.

Callers, both of which keep their own `ctx` for `deps.Users.Get`:

- `pkg/lib/authenticationflow/declarative/intent_fill_form.go`: `dataForm := NewFillFormDataForm(deps, form, user)`.
- `pkg/lib/authenticationflow/declarative/form.go` (identify step options):
  `dataForm := NewFillFormDataForm(deps, form, user)`; drop the `if err != nil` that follows it.

### 3.2 Authflow dependencies

- `pkg/lib/authenticationflow/dependencies.go`: delete the `TranslationService` interface and the
  `Translations TranslationService` field of `Dependencies`. Both were added for profile filling
  (`e7bc4faf90`), and nothing else uses them.
- `pkg/lib/deps/deps_common.go`: delete
  `wire.Bind(new(authenticationflow.TranslationService), new(*translation.Service))`.
  `*translation.Service` stays provided; other interfaces bind to it.
- Regenerate wiring with `make generate`. Expected generated change: `pkg/auth/wire_gen.go` loses
  `Translations: ...` in each `authenticationflow.Dependencies{...}` literal, and any
  `translation.Service` construction used only there.

## 4. Compatibility and deployment behavior

- Authflow API: `constraints.pattern_message` is removed. As with the label removal, no Custom UI
  uses it yet, so there is no `docs/BREAKING-CHANGES.md` entry; mention it in the PR description.
- A field whose only constraint was a `pattern` still has `constraints: { "pattern": ... }`, so no
  field loses its `constraints` object.
- Auth UI shows the same pattern message as before.
- `FillFormData` is computed per response and not persisted.

## 5. File-level change plan

| File | Change |
|---|---|
| `pkg/lib/authenticationflow/declarative/data_fill_form.go` | Per 3.1. |
| `pkg/lib/authenticationflow/declarative/data_fill_form_test.go` | Delete `fakeTranslationService`; constraint case asserts no message. |
| `pkg/lib/authenticationflow/declarative/intent_fill_form.go` | New `NewFillFormDataForm` call. |
| `pkg/lib/authenticationflow/declarative/form.go` | New `NewFillFormDataForm` call. |
| `pkg/lib/authenticationflow/dependencies.go` | Delete `TranslationService` and `Dependencies.Translations`. |
| `pkg/lib/deps/deps_common.go` | Delete the `wire.Bind`. |
| `pkg/auth/wire_gen.go` (and any other `wire_gen.go` that `make generate` changes) | Regenerated. |
| `e2e/tests/profile_filling/constraints.test.yaml` | Remove `pattern_message` and `extra_files_directory`. |
| `e2e/tests/profile_filling/boolean_terms.test.yaml` | Remove `extra_files_directory`. |
| `e2e/tests/profile_filling/var_translations/` | Deleted. |
| `e2e/tests/webapp/fill_form/pattern_message.test.yaml` | New. |
| `e2e/tests/webapp/fill_form/var_pattern_message/templates/en/translation.json` | New. |

`var_translations` goes because, after this change, no Authflow API output depends on a
translation. Its keys only fed labels and `pattern_message`. `boolean_terms` and `constraints` are
its only users.

## 6. Test plan

### Unit tests

`pkg/lib/authenticationflow/declarative/data_fill_form_test.go` (Convey):

- Delete `fakeTranslationService` and the `translations` variable; `deps` is
  `&authflow.Dependencies{Config: cfg}`.
- `build` calls `NewFillFormDataForm(deps, ...)` and no longer checks an error.
- Rename "string constraints, and pattern_message only with a translation" to "string
  constraints". It keeps the `/x_employee_id` and `/given_name` assertions and drops the
  translation leaf.

Run: `go test ./pkg/lib/authenticationflow/... ./pkg/lib/deps/... ./pkg/auth/...`.

### E2E

- `constraints.test.yaml`: the `/x_employee_id` constraints become `{ "pattern": "^E[0-9]{5}$" }`.
- New `e2e/tests/webapp/fill_form/pattern_message.test.yaml`, modelled on
  `signup_fill_form.test.yaml`, with `extra_files_directory: ./var_pattern_message`. Custom
  attribute `/x_employee_id` (string, `pattern: "^E[0-9]{5}$"`, `end_user: readwrite`); signup with
  `identify` (username) and `create_authenticator` (password), then `fill_form` with
  `/x_employee_id` required.

  `var_pattern_message/templates/en/translation.json`:

  ```json
  {
    "custom-attribute-pattern-error-/x_employee_id": "Start with E, then 5 digits"
  }
  ```

  Cases:

  - Submitting `12345` stays on `/authflow/v2/fill_form` and the page contains
    `Start with E, then 5 digits`.
  - Submitting `E12345` redirects (302) and the flow finishes.

Run per the `write-e2e-test` skill (unset `HTTP(S)_PROXY`).

## 7. Fixed behavioral decisions

- The Authflow API renders no translation. Every piece of text in a form comes from the client.
- `pattern` stays in `constraints`; only its message is removed.
- The Authflow translation dependency is removed rather than kept unused.

## 8. Atomic commit plan

1. **`[Profile Filling] Remove pattern_message from fill_form_data`**
   - `pkg/lib/authenticationflow/declarative/data_fill_form.go`, `data_fill_form_test.go`,
     `intent_fill_form.go`, `form.go`
   - `pkg/lib/authenticationflow/dependencies.go`, `pkg/lib/deps/deps_common.go`, regenerated
     `wire_gen.go` files (same commit, so every commit builds)
   - `e2e/tests/profile_filling/constraints.test.yaml`, `boolean_terms.test.yaml`; delete
     `e2e/tests/profile_filling/var_translations/`
   - API output change, with the API e2e expectation updated in the same commit.
2. **`[Profile Filling] Add Auth UI e2e test for pattern messages`**
   - `e2e/tests/webapp/fill_form/pattern_message.test.yaml`,
     `e2e/tests/webapp/fill_form/var_pattern_message/templates/en/translation.json`
   - Test only. It passes before and after commit 1, because Auth UI never read
     `pattern_message`.
