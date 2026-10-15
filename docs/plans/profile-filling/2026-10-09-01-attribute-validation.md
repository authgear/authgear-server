# Profile Filling — Part 1: Attribute types, constraints, and labels

Spec: [docs/specs/user-profile/design.md](../../specs/user-profile/design.md) (`boolean`, String constraints, Standard attribute labels), [docs/specs/user-profile/profile-filling.md](../../specs/user-profile/profile-filling.md) (Validation).

Parts:

1. **Attribute types, constraints, and labels** (this file)
2. [Authflow config and runtime](./2026-10-09-02-authflow.md)
3. [Auth UI](./2026-10-09-03-auth-ui.md)
4. [E2E tests](./2026-10-09-04-e2e.md)

Parts land in order. Part 2 depends on the config types, the constraint checker, and the label keys added here.

## 1. Goal / scope

- New custom attribute type `boolean`.
- String constraints (`min_length`, `max_length`, `allowed_characters`, `pattern`) on custom attributes of type `string`, and on string standard attributes via `user_profile.standard_attributes.validation`.
- Constraints checked on every write path that goes through `stdattrs.ServiceNoEvent.UpdateStandardAttributes` or `customattrs.ServiceNoEvent.updateCustomAttributes`: settings page, Admin API / portal, hook mutations, user import, and authflow.
- Built-in translation keys `standard-attribute-label-{pointer}`.
- Settings page: boolean checkbox, constraint hints, and constraint error messages.
- Portal user detail screen: edit `boolean` attributes, and send back any custom attribute whose type the portal does not understand, unchanged.

Out of scope: portal config editors for the boolean type, string constraints, and `standard_attributes.validation`. Developers set these in `authgear.yaml`.

## 2. Config model and schema

All in `pkg/lib/config/user_profile.go`.

### 2.1 Shared constraint struct (config layer)

```go
type AllowedCharacters string

const (
	AllowedCharactersLetters AllowedCharacters = "letters"
	AllowedCharactersDigits  AllowedCharacters = "digits"
	AllowedCharactersSpaces  AllowedCharacters = "spaces"
	AllowedCharactersSymbols AllowedCharacters = "symbols"
)

// StringConstraints is the resolved view shared by custom and standard attributes.
type StringConstraints struct {
	MinLength         *int
	MaxLength         *int
	AllowedCharacters []AllowedCharacters
	Pattern           string
}
```

`StringConstraints` has no JSON tags. It is a runtime view built from the two config structs below. The YAML keys stay flat on each config item, as in the spec.

### 2.2 Custom attributes

`CustomAttributesAttributeConfig` gets:

```go
MinLength         *int                `json:"min_length,omitempty"`
MaxLength         *int                `json:"max_length,omitempty"`
AllowedCharacters []AllowedCharacters `json:"allowed_characters,omitempty"`
Pattern           string              `json:"pattern,omitempty"`
```

and:

```go
// StringConstraints returns nil unless Type is string and at least one constraint is set.
func (c *CustomAttributesAttributeConfig) StringConstraints() *StringConstraints
```

`CustomAttributeTypeBoolean CustomAttributeType = "boolean"` is added to the constants.

Schema `CustomAttributesAttributeConfig`:

- Add `"boolean"` to the `type` enum.
- New `allOf` item: `if type const "string" then properties { min_length: {type: integer, minimum: 0}, max_length: {type: integer, minimum: 1}, allowed_characters: {type: array, minItems: 1, uniqueItems: true, items: {enum: [letters, digits, spaces, symbols]}}, pattern: {type: string, format: x_re2_regex} }`.
- New `allOf` item: `if type not const "string" then not anyOf [required min_length, required max_length, required allowed_characters, required pattern]`. This rejects constraints on other types.
- The "not number/integer/enum" item stays as is. `boolean` takes no extra keys.

`ToSchemaBuilder()`: new case `CustomAttributeTypeBoolean: builder.Type(validation.TypeBoolean)`. The string case is unchanged. Constraints are checked in Go (see 3.1), because `SchemaBuilder` has no `pattern`, and `allowed_characters` is not a JSON Schema keyword.

`ParseString()`: new case `CustomAttributeTypeBoolean`, which returns `strconv.ParseBool(strRepr)`. On a parse error it returns the error, matching the integer and number cases.

### 2.3 Standard attributes

```go
type StandardAttributesConfig struct {
	Population    *StandardAttributesPopulationConfig     `json:"population,omitempty"`
	AccessControl []*StandardAttributesAccessControlConfig `json:"access_control,omitempty"`
	Validation    []*StandardAttributesValidationConfig    `json:"validation,omitempty"`
}

type StandardAttributesValidationConfig struct {
	Pointer           string              `json:"pointer,omitempty"`
	MinLength         *int                `json:"min_length,omitempty"`
	MaxLength         *int                `json:"max_length,omitempty"`
	AllowedCharacters []AllowedCharacters `json:"allowed_characters,omitempty"`
	Pattern           string              `json:"pattern,omitempty"`
}

func (c *StandardAttributesValidationConfig) StringConstraints() *StringConstraints

// GetStringConstraints returns nil when the pointer has no entry.
func (c *StandardAttributesConfig) GetStringConstraints(pointer string) *StringConstraints
```

`SetDefaults` is unchanged, because `Validation` has no defaults.

New schema `StandardAttributesValidationConfig`: object, `additionalProperties: false`, `required: [pointer]`. Its `pointer` is an enum of exactly these 10 pointers:

- Single-line strings: `/name`, `/given_name`, `/family_name`, `/middle_name`, `/nickname`, `/address/locality`, `/address/region`, `/address/postal_code`.
- Multi-line strings: `/address/formatted`, `/address/street_address`.

The other properties are the same as the custom attribute string constraints. `StandardAttributesConfig` schema gets `"validation": { "type": "array", "items": { "$ref": "#/$defs/StandardAttributesValidationConfig" } }`.

### 2.4 Cross-field validation

In `pkg/lib/config/config.go`:

- `validateCustomAttribute` (existing): when `min_length > max_length`, emit `EmitError("maxLength", {"expected": max_length, "actual": min_length})` at `…/min_length`. This mirrors the existing `minimum > maximum` check, which emits `maximum` with `{maximum, actual}` at `…/minimum` (`config.go:339-351`).
- New `validateStandardAttributesValidation(ctx *validation.Context, c *UserProfileConfig)`, called from `AppConfig.Validate` next to `validateCustomAttribute`:
  - A duplicate `pointer` emits `EmitError("duplicated", nil)` at `/user_profile/standard_attributes/validation/{i}/pointer`, the format the existing custom-attribute duplicate check uses (`config.go:316-333`).
  - `min_length > max_length` gets the same error as for custom attributes.

No migration. All keys are new and optional.

One strictness note: `CustomAttributesAttributeConfig` has no `additionalProperties: false` (`user_profile.go:182-313`), so a stray `min_length` or `pattern` key is silently accepted today. After this change it is rejected on non-string types and enforced on string types. Such keys had no meaning before and are not documented, so this needs no BREAKING-CHANGES entry.

## 3. Runtime

### 3.1 Constraint checker

New file `pkg/lib/authn/attrs/string_constraints.go`. The package already imports `config`.

```go
// ValidateStringConstraints emits one error per failed constraint on vctx.
// An empty value is not checked.
func ValidateStringConstraints(vctx *validation.Context, c *config.StringConstraints, value string, multiline bool)
```

Errors emitted, in this order. Each is checked independently, and all failures are emitted:

| Check | Keyword | Info |
|---|---|---|
| `utf8.RuneCountInString(value) < *MinLength` | `minLength` | `{"expected": min, "actual": n}` |
| `> *MaxLength` | `maxLength` | `{"expected": max, "actual": n}` |
| a rune outside the allowed kinds | `allowedCharacters` | `{"expected": ["letters","digits"]}` (config order) |
| `!regexp.MustCompile(Pattern).MatchString(value)` | `pattern` | `{"expected": Pattern}` |

`minLength`, `maxLength`, and `pattern` reuse the JSON Schema keyword names and info keys (`expected`, `actual`, as in `jsonschema/keyword_pattern.go`), so existing renderers keep working. The portal's `errors.validation.pattern` already uses `{expected}`. `allowedCharacters` is new, and Part 1 commit 7 adds it to the portal error mapping.

Rune classes, in a private helper `runeKind(r rune) (AllowedCharacters, bool)`:

- letters: `unicode.IsLetter(r) || unicode.Is(unicode.M, r)`. Marks count as letters, so decomposed accented letters pass.
- digits: `'0' <= r && r <= '9'`.
- spaces: `r == ' '`.
- symbols: `unicode.IsGraphic(r)`, when none of the above match.
- anything else, such as control characters and tabs: no kind, so the rune is rejected.

The character check runs only when `AllowedCharacters` is non-empty. Without it, every kind is allowed, as the spec says.
- `'\n'` and `'\r'` are accepted when `multiline` is true, whatever the allowed kinds.

`regexp.MustCompile` is safe because the config format `x_re2_regex` already rejected invalid patterns. Patterns are compiled on each check, with no cache. A process-wide cache would grow without bound across tenants and config edits, and the values checked are short.

```go
// IsMultilineStandardAttribute reports whether pointer is /address/formatted or /address/street_address.
func IsMultilineStandardAttribute(pointer string) bool
```

### 3.2 Standard attributes write path

`pkg/lib/feature/stdattrs/service_noevent.go`, `ServiceNoEvent.UpdateStandardAttributes`. After `rawUser` is loaded, and before `CheckWrite`:

```go
err = s.validateStringConstraints(rawUser.StandardAttributes, stdAttrs)
```

New private method:

```go
func (s *ServiceNoEvent) validateStringConstraints(original map[string]any, incoming map[string]any) error
```

Call flow:

1. For each `v` in `s.UserProfileConfig.StandardAttributes.Validation`:
   1. Resolve `v.Pointer` against `incoming` and `original` with `jsonpointer.MustParse(v.Pointer).Traverse`. A missing path counts as an empty string.
   2. Skip when the incoming value is not a non-empty string, or equals the original value. Constraints apply to changed values only, so narrowing a constraint affects future writes only.
   3. Call `attrs.ValidateStringConstraints(vctx.Child(ptr tokens...), v.StringConstraints(), value, attrs.IsMultilineStandardAttribute(v.Pointer))`.
2. Return `vctx.Error("invalid standard attributes")`, which is a `*validation.AggregatedError` and becomes `ValidationFailed`.

Locations are `/given_name` or `/address/locality`, the same shape as `stdattrs.Validate`.

### 3.3 Custom attributes write path

`pkg/lib/feature/customattrs/service_noevent.go`, `updateCustomAttributes`. After `validate` and `GetRaw`:

```go
func (s *ServiceNoEvent) validateStringConstraints(pointers []string, original customattrs.T, incoming customattrs.T) error
```

For each pointer in `pointers`:

1. Find its config. Skip unless `c.StringConstraints() != nil`.
2. Skip unless the incoming value is a non-empty string that differs from the original repr value.
3. Call `attrs.ValidateStringConstraints(vctx.Child(ptr[0]), cons, value, false)`.

It returns `vctx.Error("invalid custom attributes")`. The original value comes from `fromStorageForm(rawUser.CustomAttributes)`, which `updateCustomAttributes` already computes.

Because both checks sit in `ServiceNoEvent`, they cover:

- the Admin API facade
- the settings page
- hook mutations (`PerformEffectsOnUser`)
- user import
- authflow `NodeDoUpdateUserProfile`, and the new form node from Part 2

Hook mutations and user import are new enforcement points. Once a project configures a constraint, a blocking-hook mutation that writes a violating value fails the operation with `ValidationFailed`, and user import rejects that record. The spec says constraints are checked on every write, and only projects that add the new keys are affected.

Population (`PopulateStandardAttributes`, `PopulateIdentityAwareStandardAttributes`) writes the store directly and stays unchecked. Populated values are re-checked when a form submits them (Part 2).

### 3.4 Boolean values

- `customattrs.ServiceNoEvent.UpdateCustomAttributesWithForm` already calls `ParseString`. `"true"` and `"false"` parse to booleans, and `""` still deletes.
- CSV export needs no change: `pkg/lib/userexport/model.go` already formats `bool` values.

### 3.5 Label translation keys

`resources/authgear/templates/en/translation.json` gets `standard-attribute-label-{pointer}` for every standard attribute:

- `/email`, `/phone_number`, `/preferred_username`
- `/name`, `/given_name`, `/family_name`, `/middle_name`, `/nickname`
- `/profile`, `/picture`, `/website`
- `/gender`, `/birthdate`, `/zoneinfo`, `/locale`, `/address`

The values are English labels, for example `"standard-attribute-label-/given_name": "Given name"`. Other locales follow the normal translation process. These keys are read by Part 2 (Authflow API labels) and Part 3 (Auth UI).

## 4. Settings page (Auth UI)

| File | Change |
|---|---|
| `pkg/auth/handler/webapp/viewmodels/settings_profile.go` | `CustomAttribute` gets `StringConstraints *config.StringConstraints` (filled from `c.StringConstraints()`). `SettingsProfileViewModel` gets `StandardAttributeStringConstraints map[string]*config.StringConstraints` (from `StandardAttributes.Validation`). |
| `resources/authgear/templates/en/web/authflowv2/settings_profile_edit_custom.html` | New `{{ if (eq $ca.Type "boolean") }}` branch:<br>• `<input type="checkbox" name="{{ $ca.Pointer }}" value="true" data-form-state-target="input">` with `checked` when the current value is `true`, followed by `<input type="hidden" name="{{ $ca.Pointer }}" value="false">`. Only the checkbox gets `data-form-state-target`: `formState.ts` keys values by `name`, so a targeted hidden input would overwrite the checkbox state and Save would never enable.<br>• `JSONPointerFormToMap` uses `form.Get`, which returns the first value, so a checked box posts `true` and an unchecked one posts `false`. `UpdateCustomAttributesWithForm` updates only pointers present in the form (`service_noevent.go:175-206`).<br>• The label comes from `$.Translations.RenderText` on the label key, a raw string with no translation `<span>` wrapper, piped into `sanitizeLinksHTML`.<br>• The navbar `Title` (lines 6-21) uses `linksToText` on the same string.<br>The string branch renders `authflowv2/__string_constraints_hint.html`. |
| `resources/authgear/templates/en/web/authflowv2/settings_profile.html` | The boolean display value uses `v2.page.settings-profile.default.boolean-true` / `-false`. The item label uses `linksToText`, because the item is already an `<a href=edit>` (lines 139-163) and must not contain nested links. |
| `resources/authgear/templates/en/web/authflowv2/settings_profile_edit_name.html` (covers `/name`, `/given_name`, `/family_name`, `/middle_name`, `/nickname`) and `settings_profile_edit_address.html` | Render `__string_constraints_hint.html` under each input that has an entry in `StandardAttributeStringConstraints`. |
| new `resources/authgear/templates/en/web/authflowv2/__string_constraints_hint.html` | Input: `dict "Constraints" … `. Renders `v2.component.string-constraints-hint.length` (ICU with `min`/`max` selects) and `v2.component.string-constraints-hint.allowed-characters` (`kinds`, joined from `v2.component.allowed-characters.kind.{kind}` with `v2.component.allowed-characters.separator`). Registered in `pkg/lib/web/html.go` and added to `BaseComponentsHTML` only. `SettingsComponentsHTML` already includes `BaseComponentsHTML` (`html.go:7-13`). |
| `resources/authgear/templates/en/web/authflowv2/__error.html` | New `ValidationFailed` branches. `minLength` renders `v2.error.min-length` (`expected`). `allowedCharacters` renders `v2.error.allowed-characters` (`kinds` joined as above). `pattern` tries `custom-attribute-pattern-error-{location}`, then `standard-attribute-pattern-error-{location}`, with `$.Translations.HasKey`, and falls back to `v2.error.pattern`. It uses each cause's own `.location`, which is the attribute pointer on the settings page, and works on pages with several fields. The two key families cannot collide: custom pointers may not be standard pointers (`user_profile.go:193-227`). For Auth UI form errors, where the location is `/fields/{i}/value`, the caller passes `PatternPointer` (Part 3), which takes precedence over `.location`. |
| `resources/authgear/templates/en/translation.json` | The `v2.error.*`, `v2.component.*`, and `boolean-true/false` keys above. |
| new `pkg/util/htmlutil/links.go` | `func SanitizeLinksHTML(s string) template.HTML`: tokenizes with `golang.org/x/net/html`; keeps `<a>` whose `href` scheme is `http` or `https`, rewritten as `<a href="…" target="_blank" rel="noopener noreferrer">`; escapes every other tag. `func LinksToText(s string) string`: the same tokenizer, keeping only text. |
| `pkg/util/template/func.go` | Register `sanitizeLinksHTML` and `linksToText` in `makeTemplateFuncMap`. Both take `any` and accept `string` or `template.HTML`, converting the latter with `string(...)` before sanitizing. The label must still come from `RenderText`, not `translate`, because `translate` wraps the text in a `<span class="translated-message">` that the sanitizer would escape. |

## 5. Portal user detail screen

`portal/src/graphql/adminapi/UserDetailsScreen.tsx` and `portal/src/graphql/adminapi/UserProfileForm.tsx`.

- `portal/src/types.ts`: add `"boolean"` to `customAttributeTypes`. The `custom-attribute-type.boolean` label goes in `portal/src/locale-data/en.json`.
- `portal/src/EditCustomAttributeForm.tsx`: the type dropdown (line 315) lists every `customAttributeTypes` entry, so adding `boolean` would offer it in the editor. In create mode, filter `boolean` out. In edit mode, keep the current value as an option, because the select is disabled there and an existing boolean attribute would otherwise show a blank type. This is the only editor change.
- `makeCustomAttributesState`: a boolean is stored as `"true"`/`"false"`. Values of a type the portal does not know are not put into the editable state.
- `makeCustomAttributesFromState(state, config, original)`: new `original: CustomAttributes` parameter, the user's custom attributes as loaded.
  - New case `"boolean"` → `out[fieldName] = value === "true"`.
  - In `default:`, and for every config entry whose type is not in `customAttributeTypes`, copy `original[fieldName]` unchanged when it is defined. Every caller passes the loaded user's `customAttributes`.
- `portal/src/error/validation.ts`, `portal/src/error/parse.ts` (`errorCauseMessageIDs`), and `portal/src/locale-data/en.json`: add the `allowedCharacters` cause with message `errors.validation.allowedCharacters`.
- `UserProfileForm.tsx` `CustomAttributeControl`: new `case "boolean"` renders a FluentUI `Checkbox`, disabled when the attribute's `portal_ui` is `readonly`. Follow the `update-portal-ui` skill for the label.

## 6. Compatibility and deployment

- Config: additive only. No migration. No BREAKING-CHANGES entry, because existing configs and requests behave the same.
- Stored data: booleans are stored as JSON `true`/`false` in `_auth_user.custom_attributes`. Readers already copy values unchanged.
- Rolling deploy: an old pod loading a config with `boolean` or constraints fails schema validation. The new keys appear only after a developer writes them through the new version, so a mixed-version window only matters if config is edited during the deploy. That is the same rule as any new config key.

## 7. Test plan

| Test | Style | Cases |
|---|---|---|
| `pkg/lib/config/testdata/config_tests.yaml` | fixture | invalid `pattern` rejected; `min_length > max_length` rejected (custom and standard); duplicate `standard_attributes.validation` pointer rejected; non-string standard pointer (`/birthdate`) rejected. |
| `pkg/lib/config/testdata/custom_attributes_tests.yaml` (schema fixture, loaded by `schema_test.go:40`) | fixture | boolean type accepted; constraints on `string` accepted; constraints on `integer` rejected; `min_length` on `boolean` rejected. |
| `pkg/lib/config/user_profile_test.go` | Convey | boolean `ToSchemaBuilder` accepts `true`/`false` and rejects `"true"`; `ParseString` for boolean; `StringConstraints()` nil for non-string. |
| new `pkg/lib/authn/attrs/string_constraints_test.go` | Convey (matches `attrs_test.go`) | code-point counting with emoji; each kind; marks as letters; tab rejected; newline only when multiline; pattern unanchored; multiple failures emitted together; empty value skipped; no character check without `allowed_characters`; `pattern` info is `{expected}`. |
| `pkg/lib/feature/stdattrs/service_noevent_test.go` | Convey | changed value checked; unchanged violating value passes; nested `/address/locality` location. |
| `pkg/lib/feature/customattrs/service_noevent_test.go` | Convey | same three cases for custom attributes; boolean write via form `"false"`. |
| `pkg/util/htmlutil/links_test.go` | Convey | `https` link kept with target and rel; `javascript:` href dropped; `<script>` escaped; `LinksToText` keeps only text. |
| `pkg/api/event/blocking` or `pkg/lib/hook/sink_test.go` | Convey | a hook mutation writing a value that violates a constraint fails with `ValidationFailed`. |
| portal | `npm run typecheck`, `npm run lint` in `portal/` | — |

## 8. Fixed behavioral decisions

- New constraints check only values that change in the write. A form submission re-checks every submitted value (Part 2).
- Error keywords: `minLength`, `maxLength`, `pattern`, `allowedCharacters`.
- The portal never drops a custom attribute it cannot render. It sends back the loaded value.
- With a constraint configured, hook mutations and user import are rejected when they write a violating value.
- `/address/formatted` can have a constraint but has no field on the settings address page. The constraint applies to every other write path.

## 9. Atomic commit plan

The portal fix lands first. Once config accepts `boolean` (commit 2), a portal save would otherwise delete boolean values: `makeCustomAttributesFromState` drops unknown types (`UserDetailsScreen.tsx:255-299`), the Admin API replaces all custom attributes, and `T.Update` removes missing pointers (`authn/customattrs/model.go:53-68`).

1. **`[Portal] Keep custom attributes of unknown types on user save`**
   - Files: `portal/src/graphql/adminapi/UserDetailsScreen.tsx`: the `original` parameter and the `default:` copy-through. This does not depend on `boolean`.
2. **`Add boolean custom attribute type`**
   - Files: `pkg/lib/config/user_profile.go`, `pkg/lib/config/testdata/custom_attributes_tests.yaml`, `pkg/lib/config/user_profile_test.go`.
   - Run `make export-schemas`.
3. **`Add string constraints to custom and standard attributes config`**
   - Files: `pkg/lib/config/user_profile.go` (structs, schema, `StringConstraints()`, `GetStringConstraints`), `pkg/lib/config/config.go` (`validateCustomAttribute`, `validateStandardAttributesValidation`), `pkg/lib/config/testdata/config_tests.yaml`, `pkg/lib/config/testdata/custom_attributes_tests.yaml`.
   - Run `make export-schemas`.
4. **`Check string constraints on attribute writes`**
   - Files: `pkg/lib/authn/attrs/string_constraints.go` + test, `pkg/lib/feature/stdattrs/service_noevent.go` + test, `pkg/lib/feature/customattrs/service_noevent.go` + test, and the hook-mutation test.
   - No wiring change.
5. **`Add standard attribute label translations`**
   - Files: `resources/authgear/templates/en/translation.json`.
6. **`[Settings] Support boolean attributes and string constraints`**
   - Files: `pkg/util/htmlutil/links.go` + test, `pkg/util/template/func.go`, `pkg/auth/handler/webapp/viewmodels/settings_profile.go`, the settings templates listed in section 4, `__string_constraints_hint.html`, `__error.html`, `pkg/lib/web/html.go`, `resources/authgear/templates/en/translation.json`.
7. **`[Portal] Edit boolean custom attributes and show allowedCharacters errors`**
   - Files: `portal/src/types.ts`, `portal/src/locale-data/en.json`, `portal/src/EditCustomAttributeForm.tsx`, `portal/src/graphql/adminapi/UserDetailsScreen.tsx` (`boolean` case), `portal/src/graphql/adminapi/UserProfileForm.tsx`, `portal/src/error/validation.ts`, `portal/src/error/parse.ts`.
8. **`chore: Update vettedpositions`**, only if `make update-vettedpositions` reports moved lines.
