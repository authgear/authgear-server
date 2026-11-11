# SAML `audience_includes_acs_url` implementation plan

Spec: [docs/specs/saml.md](../../specs/saml.md) (`saml.service_providers[].audience_includes_acs_url`).
Linear: DEV-4017.

## 1. Goal / scope

Let a service provider opt out of receiving the ACS URL as an extra `<Audience>` in the
`<AudienceRestriction>` of the issued assertion. Notion rejects a response with more than one
`<Audience>`, and today there is no config that avoids the extra one.

In scope: project config field and schema, response builder, unit tests, e2e test, portal SAML form.

Out of scope: changing the default behaviour, SP metadata parsing (it does not fill `audience` today).

## 2. Config model and schema

File: `pkg/lib/config/saml.go`.

`SAMLServiceProviderConfig` gets one field, placed after `Audience`:

```go
AudienceIncludesACSURL *bool `json:"audience_includes_acs_url,omitempty"`
```

`*bool` because unset means `true`; an explicit `false` must be distinguishable from unset.

Method, following `OAuthDynamicClientRegistrationConfig.IsInitialAccessTokenRequired`:

```go
func (c *SAMLServiceProviderConfig) ShouldIncludeACSURLInAudience() bool {
	return c.AudienceIncludesACSURL == nil || *c.AudienceIncludesACSURL
}
```

JSON schema `SAMLServiceProviderConfig`:

- `properties`: add `"audience_includes_acs_url": { "type": "boolean" }`.
- `allOf`: add a second rule:

  ```json
  {
    "if": {
      "properties": { "audience_includes_acs_url": { "const": false } },
      "required": ["audience_includes_acs_url"]
    },
    "then": { "required": ["audience"] }
  }
  ```

No `SetDefaults` change: the default lives in `ShouldIncludeACSURLInAudience`, so the field stays
absent from configs that never set it. No migration: existing configs have no such key and keep
today's behaviour.

## 3. Runtime flow

Entry point: `pkg/lib/saml/service.go`, `(*Service).IssueLoginSuccessResponse`. The inline block
that builds `audiences` (currently at ~line 386–400) is replaced by a call to a new package-level
function in the same file:

```go
func resolveAudiences(
	sp *config.SAMLServiceProviderConfig,
	callbackURL string,
	inResponseToAuthnRequest *samlprotocol.AuthnRequest,
) []string
```

`resolveAudiences`:

1. `audiences := setutil.Set[string]{}`
2. If `sp.Audience != ""`: add `sp.Audience`.
3. If `sp.Audience == "" || sp.ShouldIncludeACSURLInAudience()`: add `callbackURL`.
4. If `inResponseToAuthnRequest != nil`: add every value of `inResponseToAuthnRequest.CollectAudiences()`.
5. Return `audiences.Keys()` (sorted, as today).

`IssueLoginSuccessResponse` loops over the returned slice to build
`samlprotocol.AudienceRestriction.Audience`, exactly as it does today over `audiences.Keys()`.
No error path is added.

Resulting audiences:

| `audience` | `audience_includes_acs_url` | Audiences (before requested ones) |
|---|---|---|
| unset | unset | ACS URL |
| set | unset / `true` | `audience`, ACS URL |
| set | `false` | `audience` |
| unset | `false` | rejected by schema |

`ValidateAuthnRequest` is unchanged: ACS URLs stay allowed as requested audiences, so an SP that
requests its ACS URL still receives it in step 4 even with `false`.

## 4. Compatibility and deployment

- Unset behaves exactly as today. No `docs/BREAKING-CHANGES.md` entry.
- Mixed-version rollout: an older server reading a config with `audience_includes_acs_url` fails
  schema validation (`additionalProperties: false`). Only configs written after the upgrade contain
  the key, so this matters only on rollback; mention it in the PR description.
- No storage, Redis, or API payload changes.

## 5. Portal

### Config types

`portal/src/types.ts`, `SAMLServiceProviderConfig`: add `audience_includes_acs_url?: boolean;` after `audience`.

### Form state (`portal/src/hook/useOAuthClientForm.ts`)

- `FormStateSAMLServiceProviderConfig`: add `audienceIncludesACSURL?: boolean;` after `audience`.
- Config → form (`samlServiceProviders.push`): `audienceIncludesACSURL: sp.audience_includes_acs_url,`
- Form → config (`updateSAMLServiceProviders`), after `updatedSP.audience = ...`:

  ```ts
  updatedSP.audience_includes_acs_url =
    updatedSP.audience != null && editedSP.audienceIncludesACSURL === false
      ? false
      : undefined;
  ```

  Only `false` is written, and only with an audience, so the config never fails the schema rule and
  untouched SPs stay unchanged on save.

### Screen mapping (`portal/src/graphql/portal/EditOAuthClientScreen.tsx`, `OAuthClientSAML2Content`)

- Config → form state: `audienceIncludesACSURL: samlConfig.audienceIncludesACSURL ?? defaults.audienceIncludesACSURL,`
- Form state → config: `audienceIncludesACSURL: newState.audienceIncludesACSURL,`

### Form (`portal/src/components/applications/OAuthClientSAMLForm.tsx`)

- `OAuthClientSAMLFormState`: add `audienceIncludesACSURL: boolean;` after `audience`.
- `getDefaultOAuthClientSAMLFormState`: `audienceIncludesACSURL: true`.
- Handler, following `onIsSLOEnabledChange`:

  ```ts
  const onAudienceIncludesACSURLChange = useCallback(
    (checked: boolean) => {
      onFormStateChange({ ...formState, audienceIncludesACSURL: checked });
    },
    [formState, onFormStateChange]
  );
  ```

- Directly after the `audience` `TextField`, a `grid gap-y-2 grid-cols-1` wrapper (same as the
  "Check signature" toggle) containing:
  - `Toggle` with `text` = `OAuthClientSAMLForm.sso.audienceIncludesACSURL.label`,
    `textWeight="medium"`, `disabled={formState.audience === ""}`,
    `checked={formState.audience === "" || formState.audienceIncludesACSURL}`,
    `onCheckedChange={onAudienceIncludesACSURLChange}`.
  - `Text as="p" size="2" className={styles.toggleDescription}` with
    `OAuthClientSAMLForm.sso.audienceIncludesACSURL.description`.

### Translations (`portal/src/locale-data/en.json`)

```json
"OAuthClientSAMLForm.sso.audienceIncludesACSURL.label": "Also include ACS URL as Assertion Audience",
"OAuthClientSAMLForm.sso.audienceIncludesACSURL.description": "When on, the ACS URL is sent as an extra Audience. Turn this off for service providers that accept only one Audience.",
```

## 6. Test plan

### Config fixtures — `pkg/lib/config/testdata/config_tests.yaml`

Next to the existing `saml-service-provider*` cases:

- `saml-service-provider-audience-includes-acs-url-false`: `audience` set,
  `audience_includes_acs_url: false` → `error: null`.
- `saml-service-provider-audience-includes-acs-url-false-without-audience`: no `audience`,
  `audience_includes_acs_url: false` → the real schema error, captured by running the case, not guessed.

### Unit test — `pkg/lib/saml/service_audience_test.go`

New internal test file (`package saml`), because `service_test.go` is the external `saml_test`
package and cannot call the unexported `resolveAudiences`. Convey style, `TestResolveAudiences`, with
`callbackURL = "https://sp.example.com/acs"`:

| Case | `audience` | flag | AuthnRequest audiences | Expected |
|---|---|---|---|---|
| no audience | unset | unset | none | `[acs]` |
| audience, flag unset | `https://sp.example.com` | unset | none | `[audience, acs]` |
| audience, flag true | `https://sp.example.com` | `true` | none | `[audience, acs]` |
| audience, flag false | `https://sp.example.com` | `false` | none | `[audience]` |
| flag false, ACS requested | `https://sp.example.com` | `false` | `[acs]` | `[audience, acs]` |
| requested audience added | unset | unset | `[https://other.example.com]` | `[acs, other]` (sorted) |

### E2E — `e2e/tests/saml/sso.test.yaml`

The XML matcher checks only the listed child elements in order, so it cannot detect an extra
trailing `<Audience>`. New case "SAML SSO - IdP Initiated without ACS URL as Audience", copied from
the IdP-initiated case, with `audience: https://samlsp.localhost/sp` and
`audience_includes_acs_url: false`. Because audiences are sorted and `.../acs` < `.../sp`, the
expected `<saml:Conditions>` lists `<saml:Audience>https://samlsp.localhost/sp</saml:Audience>` as
the only (first) audience; the old behaviour would put `.../acs` first and fail the match.

Existing cases are unchanged: they do not set the flag, so they still get two audiences.

### Portal

`npm run typecheck`, `npm run lint` (`make -C portal lint`) in `portal/`.

## 7. Fixed behavioural decisions

- Default is `true` (current behaviour).
- `false` without `audience` is rejected by the schema, not ignored.
- Audiences requested in the AuthnRequest are always added, regardless of the flag.
- Portal writes only `false`; it never writes `true`.
- Portal toggle is disabled and shown on when Assertion Audience is empty.

## 8. Atomic commit plan

1. **`doc: Add implementation plan for saml audience_includes_acs_url`**
   - `docs/plans/saml-audience-includes-acs-url/2026-11-11-implementation.md`.
2. **`[SAML] Add audience_includes_acs_url config`**
   - `pkg/lib/config/saml.go` (field, schema, `ShouldIncludeACSURLInAudience`).
   - `pkg/lib/config/testdata/config_tests.yaml` (two cases).
   - Run `make export-schemas` if it changes checked-in schema files; include them.
   - Config only; runtime ignores it until commit 3.
3. **`[SAML] Respect audience_includes_acs_url in SAML response`**
   - `pkg/lib/saml/service.go` (`resolveAudiences`, call site in `IssueLoginSuccessResponse`).
   - `pkg/lib/saml/service_audience_test.go`.
   - `.vettedpositions` via `make update-vettedpositions` if line numbers move.
4. **`[SAML] Add e2e test for audience_includes_acs_url`**
   - `e2e/tests/saml/sso.test.yaml`.
5. **`[Portal] Add toggle for including ACS URL as SAML audience`**
   - `portal/src/types.ts`, `portal/src/hook/useOAuthClientForm.ts`,
     `portal/src/graphql/portal/EditOAuthClientScreen.tsx`,
     `portal/src/components/applications/OAuthClientSAMLForm.tsx`,
     `portal/src/locale-data/en.json`.
