# Name passkeys by provider

Spec: [docs/specs/webauthn.md](../../specs/webauthn.md#passkey-provider).

## 1. Goal and scope

- Read the AAGUID of every passkey identity and passkey authenticator.
- Name the provider from an embedded list of known AAGUIDs.
- Expose both as claims, label each passkey by provider on the AuthUI settings page and in the portal.

Unchanged: the database schema, the registration and assertion ceremonies, and OIDC userinfo. Userinfo lists identities with type-specific fields for `login_id` and `oauth` only, and a third-party client has no use for the user's passkey provider.

## 2. Storage and compatibility

`_auth_identity_passkey.attestation_response` and `_auth_authenticator_passkey.attestation_response` already hold the whole registration, so the AAGUID is read from it on every load. No column, migration or backfill. Every existing passkey gets a provider name as soon as this is deployed.

`identity.Passkey` and `authenticator.Passkey` gain `AAGUID string` with JSON key `aaguid,omitempty`. They are serialized into authflow and interaction state; state written before the deploy lacks the key and decodes to an empty AAGUID, which only omits the `aaguid` claim from that one flow's event payload.

## 3. Lookup — `pkg/util/passkeyprovider`

New package, in the style of `pkg/util/deviceinfo` and `pkg/util/geoip`.

- `aaguid.json`, embedded with `//go:embed`: `{"<aaguid>": {"name": "..."}}`. A names-only copy of the community list, see [Passkey provider](../../specs/webauthn.md#passkey-provider). Only `name` is read.
- `func AAGUIDFromAttestationResponse(attestationResponse []byte) (string, error)`: `protocol.ParseCredentialCreationResponseBytes`, then `uuid.FromBytes(...AuthData.AttData.AAGUID).String()`.
- `func Name(aaguid string) (name string, ok bool)`: case-insensitive lookup. The all-zero AAGUID is never listed.

## 4. Runtime flow

Load: `identity/passkey.Store.scan` and `authenticator/passkey.Store.scan` call `passkeyprovider.AAGUIDFromAttestationResponse` after decoding `creation_options`, and return its error. A stored registration that does not parse already fails every assertion (`passkey.Service.PeekAssertionResponse` parses it too).

Registration: `identity/passkey.Provider.New` and `authenticator/passkey.Provider.New` call it after `PeekAttestationResponse` succeeds.

Display: `identity.Passkey.ProviderName()` and `authenticator.Passkey.ProviderName()` return `passkeyprovider.Name(AAGUID)`, or `""`.

## 5. API

Claims, constants beside the existing passkey ones:

- `identity.IdentityClaimPasskeyAAGUID`, `identity.IdentityClaimPasskeyProviderName` in `pkg/lib/authn/identity/claim_key.go`, set by `identity.Info.ToModel`.
- `authenticator.AuthenticatorClaimPasskeyAAGUID`, `authenticator.AuthenticatorClaimPasskeyProviderName` in `pkg/lib/authn/authenticator/keys.go`, set by `authenticator.Info.ToPublicClaims`.

`aaguid` is set when the AAGUID is non-empty; `provider_name` when `ProviderName()` is non-empty. This reaches Admin API `Identity.claims` and `Authenticator.claims`, and every event payload carrying `model.Identity`. `model.Authenticator` has no claims, so authenticators in event payloads are unchanged. The GraphQL schema does not change.

## 6. AuthUI settings page

`resources/authgear/templates/en/web/authflowv2/settings_passkey.html`: the item label is `.Passkey.ProviderName`, or the new key `v2.page.settings-passkey.default.item-label-unknown-provider` ("Passkey"). The supplementary note keeps the display name and the "Added at" line. The new key is added to every locale with the value of that locale's `v2.page.settings-passkey.default.title`, which is already "Passkey" translated. The handler is unchanged: it already passes `[]*identity.Info` from `Store.List`.

## 7. Portal

`portal/src/graphql/adminapi/UserDetailsAccountSecurity.tsx`: `constructPasskeyIdentityData` reads `https://authgear.com/claims/passkey/provider_name`. `PasskeyIdentityCell` labels the cell with it, or `UserDetails.account-security.passkey.unknown-provider` ("Passkey"), and shows the display name above "Added". The query already fetches `claims`.

## 8. Tests

- `pkg/util/passkeyprovider/passkeyprovider_test.go` (Convey): AAGUID from a crafted attestation response, all-zero AAGUID, malformed input; `Name` for known, upper-case, all-zero and unlisted AAGUIDs; every key in `aaguid.json` is a lowercase UUID with a name.
- `pkg/lib/authn/identity/info_test.go`, `pkg/lib/authn/authenticator/info_test.go`: the claims with a known, unknown and missing AAGUID.
- e2e: the software authenticator in `e2e/pkg/testrunner/passkey.go` takes an AAGUID through a new template function `generatePasskeyAttestationWithAAGUID`. `e2e/tests/passkey/passkey_provider_name.test.yaml` registers a passkey with the Apple Passwords AAGUID and reads the claims of the identity and the authenticator from the Admin API; `passkey_provider_name_unknown.test.yaml` does the same with the all-zero AAGUID and expects no `provider_name`.

## 9. Commits

1. Spec and this plan.
2. `pkg/util/passkeyprovider` with tests.
3. Read the AAGUID in the passkey stores and providers.
4. Claims, with `info_test.go` updates.
5. Settings page template and translations.
6. Portal.
7. e2e.
