# Client ID Metadata Documents (CIMD)

Authgear supports Client ID Metadata Documents as defined by:

- [draft-ietf-oauth-client-id-metadata-document-02](https://datatracker.ietf.org/doc/draft-ietf-oauth-client-id-metadata-document/02/) — OAuth Client ID Metadata Document.

This document assumes familiarity with the [Client Model](./client.md), which CIMD feeds as a new client source alongside static config and [DCR](./dcr.md).

## Table of Contents

- [Glossary](#glossary)
- [Use Cases](#use-cases)
  - [UC1. MCP client publishes its identity as a CIMD document](#uc1-mcp-client-publishes-its-identity-as-a-cimd-document)
- [Configuration](#configuration)
  - [Client Limit](#client-limit)
- [OIDC Discovery Metadata](#oidc-discovery-metadata)
- [Client ID Format](#client-id-format)
- [The Client Metadata Document](#the-client-metadata-document)
  - [Fetching](#fetching)
  - [Validation](#validation)
  - [Error Handling](#error-handling)
- [Accepted Metadata Fields](#accepted-metadata-fields)
- [Client Resolution](#client-resolution)
  - [Where resolution happens](#where-resolution-happens)
- [Mapping to the Unified Client Model](#mapping-to-the-unified-client-model)
- [Redirect URI Validation](#redirect-uri-validation)
- [Client Authentication](#client-authentication)
- [Security Considerations](#security-considerations)
  - [SSRF Protection](#ssrf-protection)
  - [Authgear as an SSRF/Probing Oracle](#authgear-as-an-ssrfprobing-oracle)
  - [Denial of Service via Attacker-Chosen Fetch Targets](#denial-of-service-via-attacker-chosen-fetch-targets)
  - [Domain Trust](#domain-trust)
  - [Phishing Mitigation](#phishing-mitigation)
  - [Privacy Considerations](#privacy-considerations)
  - [Access Token Audience Binding](#access-token-audience-binding)
- [Reading a CIMD Client's Stored Config](#reading-a-cimd-clients-stored-config)

## Glossary

**Client ID Metadata Document (CIMD)** — a JSON document, hosted by the client developer at an HTTPS URL, containing OAuth/OIDC client metadata (`redirect_uris`, `client_name`, etc.). The URL itself **is** the `client_id`.

Not covered by this design:

- **CIMD Services** ([§8.10](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.10)) — a pattern where the authorization server itself hosts documents on behalf of developers who don't want to run their own server. Authgear does not offer this; a client always hosts its own document.
- **`software_statement`** ([§4.3](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-4.3)) — a signed RFC 7591 assertion that MAY accompany a Client Identifier URL. Authgear does not process it (see [Rejected or ignored fields](#rejected-or-ignored-fields)).

## Use Cases

### UC1. MCP client publishes its identity as a CIMD document

Per the [MCP Authorization specification](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization), MCP clients and authorization servers SHOULD support CIMD, and a client that supports more than one client-identification mechanism follows a fixed priority order: use a pre-registered `client_id` if it has one; otherwise use CIMD if the authorization server advertises `client_id_metadata_document_supported: true`; otherwise fall back to DCR; otherwise prompt the user. Once CIMD is enabled, it is therefore the mechanism a compliant MCP client will actually use against Authgear — DCR is only reached as a fallback.

**Required configuration:**

```yaml
oauth:
  client_id_metadata_document:
    enabled: true
```

**Admin setup (once)**

Register the MCP server's own canonical URI (e.g. `https://mcp-server.example.com`) as an API Resource, and mark it and its scopes `access_policy.allow_dynamic_third_party_client_access: true` (see [Access Token Audience Binding](#access-token-audience-binding)). This is the same URI the MCP client sends as `resource=` in Steps 3 and 4 below — **not** the client's own CIMD URL, which is unrelated to Resource registration.

This step is **required**, not optional: MCP clients always send `resource=` on both the authorization and token requests ("MCP clients **MUST** send this parameter regardless of whether authorization servers support it"). If the Resource isn't registered with the access policy enabled, that `resource=` value doesn't match any known Resource and the authorization request fails with `invalid_target` — the MCP client cannot complete the flow at all, regardless of whether its CIMD document is otherwise valid. See [Resource URI Requirements](./api-resource.md#resource-uri-requirements) for the URI's format constraints (`https://`, no query, no fragment, unique per project).

No further per-client admin action is required beyond this one-time Resource setup — any client presenting a valid CIMD `client_id` can self-identify and immediately use it. (The MCP server's own RFC 9728 Protected Resource Metadata, pointing back to this Authgear project as its `authorization_servers` entry, is configured on the MCP server side and is outside the scope of this document.)

**Step 1 — MCP client discovers Authgear as the authorization server**

Before ever contacting Authgear, the MCP client makes an unauthenticated request to the MCP server:

```
HTTP/1.1 401 Unauthorized
WWW-Authenticate: Bearer resource_metadata="https://mcp-server.example.com/.well-known/oauth-protected-resource"
```

The client fetches that Protected Resource Metadata document (RFC 9728), which names `https://myapp.authgear.cloud` in `authorization_servers`, then fetches Authgear's own discovery document and finds `client_id_metadata_document_supported: true` (see [OIDC Discovery Metadata](#oidc-discovery-metadata)) — this is what tells the client it may use its CIMD URL as `client_id` instead of falling back to DCR.

**Step 2 — Client publishes its metadata document**

```
GET /oauth/client-metadata.json HTTP/1.1
Host: mcp-client.example.com
```

```json
{
  "client_id": "https://mcp-client.example.com/oauth/client-metadata.json",
  "client_name": "Example MCP Client",
  "redirect_uris": [
    "http://127.0.0.1:3000/callback",
    "http://localhost:3000/callback"
  ],
  "grant_types": ["authorization_code", "refresh_token"],
  "response_types": ["code"],
  "token_endpoint_auth_method": "none"
}
```

This is the same shape as the [MCP spec's own example CIMD document](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization#example-metadata-document): a desktop/CLI MCP client with a local loopback callback listener, no `application_type` declared. See the loopback note in [Accepted Metadata Fields — `redirect_uris`](#accepted-metadata-fields).

**Step 3 — Authorization request with PKCE and the resource parameter**

```
GET /oauth2/authorize
  ?client_id=https://mcp-client.example.com/oauth/client-metadata.json
  &response_type=code
  &scope=openid+read:tools
  &redirect_uri=http://127.0.0.1:3000/callback
  &code_challenge=<challenge>
  &code_challenge_method=S256
  &resource=https://mcp-server.example.com HTTP/1.1
Host: myapp.authgear.cloud
```

PKCE with `S256` is mandatory per the MCP spec; Authgear already requires it for every public client regardless of source. `resource` identifies the MCP server per RFC 8707 and is required by MCP on both the authorization request and the token request.

**Step 4 — Token exchange, resource parameter repeated**

```
POST /oauth2/token
Content-Type: application/x-www-form-urlencoded

grant_type=authorization_code
&code=<code>
&code_verifier=<verifier>
&client_id=https://mcp-client.example.com/oauth/client-metadata.json
&redirect_uri=http://127.0.0.1:3000/callback
&resource=https://mcp-server.example.com
```

Authgear resolves the document (see [Fetching](#fetching)) the first time it sees this `client_id`, validates `redirect_uri` against the document's `redirect_uris`, and shows the consent screen labelled with the hostname `mcp-client.example.com` (see [Phishing Mitigation](#phishing-mitigation)). Token issuance, resource-indicator handling, and audience binding proceed exactly as for any other third-party client.

## Configuration

```yaml
oauth:
  client_id_metadata_document:
    enabled: true
    allowed_domains: []
    client_config:
      access_token_lifetime_seconds: 1800
      refresh_token_lifetime_seconds: 2592000
      refresh_token_idle_timeout_enabled: true
      refresh_token_idle_timeout_seconds: 1209600
```

- `oauth.client_id_metadata_document.enabled`: Optional. Boolean. Default `false`. When `true`, any `client_id` that parses as a valid CIMD URL (see [Client ID Format](#client-id-format)) is resolved by fetching the document.
- `oauth.client_id_metadata_document.allowed_domains`: Optional. List of glob patterns (e.g. `["*.example.com"]`). Default empty, meaning any domain is allowed (subject to the protections in [SSRF Protection](#ssrf-protection)). See [Domain Trust](#domain-trust).
- `oauth.client_id_metadata_document.client_config`: Optional. Token lifetimes applied to every CIMD-resolved client, since the document itself cannot set them. Same fields as `access_token_lifetime_seconds`, `refresh_token_lifetime_seconds`, `refresh_token_idle_timeout_enabled`, `refresh_token_idle_timeout_seconds`. Unlike DCR's `default_client_config`, this is not named "default": the persisted record a CIMD client resolves to (see [Where resolution happens](#where-resolution-happens)) is a system-maintained mirror of externally-hosted data, refreshed automatically on refetch — not something an admin creates or edits — so it isn't a natural place to attach an admin-side override either. There is no override path, planned or otherwise; this is simply _the_ config for every CIMD client.

Fetch timeout, maximum document size, and cache lifetime are fixed values, not project-configurable — see [Fetching](#fetching).

### Client Limit

**authgear.features.yaml**

```yaml
oauth:
  client_id_metadata_document:
    maximum_clients: null
```

- `oauth.client_id_metadata_document.maximum_clients`: Optional. Integer. Default `null` (no limit). The maximum number of CIMD-resolved clients the project may have persisted at once, checked against the current count of `OAuthClient` records with `source: CIMD`. Once at `maximum_clients`, resolving a `client_id` with no existing persisted record fails the authorization request with `access_denied` (see [Error Handling](#error-handling)). A `client_id` that already has a persisted record is unaffected regardless of the limit, since resolving it again doesn't create a new one.

An admin can free a slot with [`deleteDynamicClient`](./dcr.md#new-mutation), but for a CIMD client this only evicts the current persisted record — the same `client_id` can produce a new one on its next successful resolution. It is not a durable ban; see [Domain Trust](#domain-trust) for the closest thing to one.

## OIDC Discovery Metadata

When CIMD is enabled, `client_id_metadata_document_supported: true` is added to the discovery documents at `<endpoint>/.well-known/openid-configuration` and `<endpoint>/.well-known/oauth-authorization-server`, per [spec §6](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-6) (Authorization Server Metadata). The spec itself marks this property OPTIONAL for an authorization server to include; Authgear includes it whenever CIMD is enabled.

## Client ID Format

Per [spec §3](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-3) (Client Identifier URL), a CIMD `client_id` MUST:

- Use the `https` scheme.
- Contain a path component (i.e. not just `https://host`).
- Not contain a fragment component.
- Not contain a username or password (userinfo) component.
- Not contain single-dot (`.`) or double-dot (`..`) path segments.

Authgear rejects a `client_id` that violates any of these **before** any network access is attempted — this is what keeps an obviously-malformed `client_id` from ever reaching the fetch path.

This validation also disambiguates CIMD candidates from the other client*id shapes already in use: static client IDs are admin-chosen opaque strings, and DCR client IDs always start with `dcrc*`. This isn't accidental: [spec §7.1](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-7.1) recommends that an authorization server supporting both CIMD and self-generated client IDs "SHOULD ensure that the `client_id`strings it generates do not start with`https://`" — exactly the property both existing Authgear client_id shapes already have. Neither shape can accidentally parse as a `https://.../path` URL, so there is no realistic collision, but resolution order ([Client Resolution](#client-resolution)) makes static and DCR clients take precedence regardless.

Per [spec §8.3](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.3), the Client Identifier URL _is_ the client's identity — a client that changes its URL is, to Authgear, a brand new and entirely unrelated client, with no carryover of prior consent or authorization history. This is ordinary OAuth `client_id` semantics rather than new CIMD-specific behavior, but is worth calling out since a CIMD client's identity is more casually mutable (just move the file to a new URL) than a static or DCR client's.

Admins who want a specific external client pinned without relying on request-time fetching at all can already do so today: put the `https://...` string directly in a static client's `client_id` in `authgear.yaml`. Authgear never treats a statically-configured client as a CIMD candidate (see [Client Resolution](#client-resolution)), so this is simply an ordinary static client whose `client_id` happens to look like a URL. [Spec §7.2](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-7.2) documents this as a supported deployment pattern — "pre-registering Client Identifier URLs" — for admins who want the namespacing benefit of a URL-shaped identifier without ever depending on live fetching.

## The Client Metadata Document

### Fetching

Authgear fetches the document with an HTTP GET request to the `client_id` URL, requesting a JSON response, with the following fixed limits — none of them project-configurable:

| Limit                   | Value                                                                | Source                                                                                                                            |
| ----------------------- | -------------------------------------------------------------------- | --------------------------------------------------------------------------------------------------------------------------------- |
| Response size           | 5120 bytes (5 KB)                                                    | [Spec §8.7](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.7) recommended maximum |
| Request timeout         | 5 seconds, covering DNS resolution, TLS handshake, and response body | Authgear default, matching the existing blocking-webhook per-call timeout default (`sync_hook_timeout_seconds`)                   |
| HTTP redirects followed | 0                                                                    | Authgear decision — the spec does not address redirects; see [SSRF Protection](#ssrf-protection)                                  |
| Refetch interval        | 1 hour                                                               | Authgear decision — see [Where resolution happens](#where-resolution-happens)                                                     |

See [SSRF Protection](#ssrf-protection) for why the size/timeout/redirect limits are fixed rather than left to project owners to tune. The refetch interval bounds how often the persisted record described in [Where resolution happens](#where-resolution-happens) is refreshed: an hour meaningfully cuts fetch volume (fewer signals leaked per [Privacy Considerations](#privacy-considerations) §9.1, more headroom under the [DoS rate limits](#denial-of-service-via-attacker-chosen-fetch-targets) for a project with many distinct legitimate CIMD clients) at the cost of a developer waiting up to an hour to see an intentional document change take effect.

### Validation

- The response MUST be `2xx` and MUST parse as a JSON object within the size limit described above.
- Each field is checked against the rules in [Accepted Metadata Fields](#accepted-metadata-fields).
- Unrecognized properties are ignored (the spec explicitly allows additional properties).

A document that fails any MUST-level check is treated as if the fetch had failed (see [Error Handling](#error-handling)); it is never reused for a later request.

### Error Handling

If fetching or validating the document fails for any reason, Authgear treats the `client_id` as unresolvable — the authorization request fails the same way it does today for a `client_id` that matches no known client (`invalid_client`-shaped error, since redirect URI validation cannot proceed without a resolved client). Per [spec §5.2](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-5.2) (Metadata Caching), a failed or invalid fetch is never reused, so a document author who fixes their document sees the fix take effect on the very next authorization request that needs it. See [Authgear as an SSRF/Probing Oracle](#authgear-as-an-ssrfprobing-oracle) for why the error surface deliberately does not distinguish _why_ the fetch failed.

A brand-new `client_id` that would otherwise resolve successfully but finds the project already at its [client limit](#client-limit) is a distinct case: the authorization request fails with `access_denied`, not the generic `invalid_client`-shaped error above. This is not a fetch-outcome signal — it doesn't vary by target host or reveal anything about network reachability — so it falls outside the uniform-error rule in [Authgear as an SSRF/Probing Oracle](#authgear-as-an-ssrfprobing-oracle), which governs only errors arising from the fetch itself.

## Accepted Metadata Fields

### `client_id` (required)

Must be present and must equal the request URL byte-for-byte. This is the binding that prevents one host from vouching for another's identity.

### `redirect_uris` (required)

The spec itself does not define a format for `redirect_uris` — [§4.2](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-4.2) (Redirect URL Registration) only requires (via RFC 9700) that redirect URIs be pre-registered and matched exactly, which Authgear already does for every client regardless of source (see [Redirect URI Validation](#redirect-uri-validation)). The format constraints below are an Authgear choice, not a spec requirement, and deliberately diverge from [DCR's `redirect_uris` rules](./dcr.md#redirect_uris-required) in one respect: each entry must be `https://`, a loopback address (`http://localhost` or `http://127.0.0.1`, any port), or a custom URI scheme; absolute URI; no fragment component.

Unlike DCR, **loopback redirect URIs are accepted regardless of `application_type`** — DCR only allows `http://localhost` when `application_type: native`. CIMD can't reuse that gate: the [MCP Authorization spec's own example CIMD document](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization#example-metadata-document) uses `http://127.0.0.1:3000/callback` and `http://localhost:3000/callback` as redirect URIs while omitting `application_type` entirely (which defaults to `web` — see below). MCP clients are typically desktop/CLI tools that open a local callback listener without declaring themselves `native`; gating loopback on `application_type` the way DCR does would reject that reference example outright.

### `client_name` (optional)

Human-readable name shown on the consent screen and in the portal. Default when omitted: `Client <clientID>`, same fallback as DCR.

### `grant_types` (optional)

Must be a subset of `["authorization_code", "refresh_token"]`. Default when absent: `["authorization_code", "refresh_token"]`.

### `response_types` (optional)

Must be a subset of `["code"]`, and consistent with `grant_types` (same consistency rule as [DCR](./dcr.md#response_types-optional)). Default when absent: `["code"]`.

### `application_type` (optional)

Must be `web` or `native`; controls redirect URI rules as in DCR. Default: `web`.

### `token_endpoint_auth_method` (optional)

Must be `none` if present. Any other value — including `private_key_jwt` and any `client_secret_*` variant — is out of scope for this v1 proposal; see [Client Authentication](#client-authentication). Default when absent: `none`.

### `logo_uri`, `client_uri`, `tos_uri`, `policy_uri` (all optional)

Each must be `https://` if present. Same meaning as the equivalent [DCR fields](./dcr.md#accepted-client-metadata).

### Rejected or ignored fields

- `client_secret`, `client_secret_expires_at`, and any raw private key material — always ignored, per [spec §4.1](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-4.1) (Credential and Key Material Restrictions).
- `jwks_uri` — not used; confidential CIMD clients are out of scope for this proposal (see [Client Authentication](#client-authentication)).
- `software_statement` — not processed (see [Glossary](#glossary)).
- Any other property — ignored (the spec explicitly allows additional properties).

## Client Resolution

When Authgear receives a `client_id` that matches neither a static client nor a DCR client, and CIMD is enabled, it checks whether the string is a valid CIMD URL ([Client ID Format](#client-id-format)) and, if so, resolves it by fetching (or reusing a recent fetch of) the metadata document. Static and DCR clients are always checked first — a string is only ever treated as a CIMD candidate once neither matches.

Once resolved, a CIMD client is indistinguishable from any other client for the rest of the flow: consent screen, redirect URI validation, resource-indicator handling, and audience binding behave identically regardless of which of the three sources the client came from. A CIMD client has no update mechanism through Authgear at all — its only metadata is whatever is currently published at its own URL, and editing that document is how the client changes its own metadata.

### Where resolution happens

A CIMD fetch — and therefore any outbound network call — happens **only** as a side effect of `/oauth2/authorize`. The result is stored in a **persisted record keyed by `client_id`**, shared across every user and every grant for that client, not scoped to any one authorization. A fresh fetch overwrites that same record in place; it is created on first resolution and updated on every refetch thereafter (per spec §5's "SHOULD ... periodically re-fetch"), at most once per [refetch interval](#fetching).

Every other step reads that persisted record — a plain lookup, never a live fetch:

- **`/oauth2/token`** (both `authorization_code` and `refresh_token` grants) reads it to validate the client and load its config.
- **The Authorized Apps settings page** reads it to render `client_name`/`logo_uri`. There is nothing to fail: the page has no network dependency at all.

Because the record is shared rather than frozen per grant, both endpoints always see the *current* known state of the client, not what was true when any particular user originally authorized it — this is what lets Authgear implement spec §8.4/§8.4.1's "notice metadata changed compared to the last time it fetched" in the future (there is a real "last fetch" to diff against), which a per-grant snapshot could never support.

## Mapping to the Unified Client Model

CIMD fields map onto `OAuthClient` (see [client.md](./client.md)) as follows:

| CIMD field                                                | `OAuthClient` field                                                          |
| --------------------------------------------------------- | ---------------------------------------------------------------------------- |
| `client_id` (the URL)                                     | `clientID`                                                                   |
| — (always)                                                | `source`: `CIMD`                                                             |
| — (always)                                                | `kind`: `THIRD_PARTY`                                                        |
| — (always, v1)                                            | `isConfidential`: `false`                                                    |
| — (always)                                                | `isServiceClient`: `false`                                                   |
| `application_type` (default `web`)                        | `applicationType`                                                            |
| `client_name` (default: `Client <clientID>` when omitted) | `name`, `clientName`                                                         |
| `client_uri`, `logo_uri`, `tos_uri`, `policy_uri`         | `clientURI`, `logoURI`, `tosURI`, `policyURI`                                |
| `redirect_uris`                                           | `redirectURIs`                                                               |
| `grant_types`, `response_types`                           | `grantTypes`, `responseTypes`                                                |
| —                                                         | `postLogoutRedirectURIs`: always `[]`                                        |
| —                                                         | token lifetimes from `client_config`                                         |
| —                                                         | `registeredAt`: always `null` — there is no registration event, only a fetch |
| —                                                         | `lastFetchedAt`: timestamp of the most recent successful fetch (see [Where resolution happens](#where-resolution-happens)) |

All Authgear-only extension fields (`app2appEnabled`, `dpopDisabled`, `customUIURI`, `authenticationFlowAllowlist`, `preAuthenticatedURLEnabled`, ...) are fixed at their zero values.

## Redirect URI Validation

Unchanged from today's rule for any client: exact match against the resolved client's `redirect_uris` — read from the persisted, shared record described in [Where resolution happens](#where-resolution-happens), which reflects the client's current known state, not a copy frozen at any particular authorization. This matches [spec §8.4](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.4) (Changes in Client Metadata), which frames an authorization server as expected to track a document's current state rather than isolate individual grants from it.

For the `authorization_code` grant specifically, the `redirect_uri` presented at token exchange is validated against the URI bound to that authorization code at issuance — ordinary OAuth code-binding behavior (RFC 6749 §4.1.3), unrelated to CIMD and unaffected by anything discussed here.

[Spec §8.1](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.1) allows an authorization server to additionally require the redirect URI to share an origin with `client_id`/`client_uri`; Authgear does **not** enforce this by default — a client's callback domain can legitimately differ from its identity/hosting domain (e.g. a CLI tool whose CIMD document is hosted on a docs site but whose redirect URI is `http://localhost:*`).

## Client Authentication

CIMD clients in v1 are always **public**: `token_endpoint_auth_method` must be absent or `none`, PKCE is required exactly as it is for any other public client today, and no `client_secret` is ever issued or accepted. This keeps v1 scoped to the change that has no new cryptographic surface. Confidential CIMD clients via `private_key_jwt` + `jwks_uri` are out of scope for this proposal — the spec explicitly forbids shared-secret auth methods for CIMD clients regardless ([§4.1](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-4.1)), so any future addition would be key-based only, never `client_secret_post`/`client_secret_basic`.

## Security Considerations

### SSRF Protection

`client_id` is attacker-controlled — unlike webhook/SSO targets, which are admin-configured. [Spec §8.6](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.6) says the authorization server MUST NOT fetch a URL that resolves to a special-use IP address ([RFC 6890](https://www.rfc-editor.org/rfc/rfc6890)). Applied validations:

- **Resolve the hostname once per fetch attempt and connect only to an address validated in that same resolution** — a second, independent resolution at connect time is vulnerable to DNS rebinding (the attacker's own DNS server returns a public address for validation, then a special-use one for the actual connection).
- **Check every address a hostname resolves to, not just the first** — reject the whole hostname if any A/AAAA record is private (e.g. `10.0.0.0/8`), loopback (`127.0.0.0/8`, `::1`), link-local (`169.254.0.0/16`, `fe80::/10`), or otherwise non-publicly-routable.
- **Follow 0 redirects** — a redirect target hasn't been through [Client ID Format](#client-id-format) validation, and would otherwise let the previous two rules be bypassed. The spec doesn't address redirects; this is an Authgear decision.
- **Enforce the 5120-byte limit ([§8.7](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.7)) progressively while reading the response**, not via `Content-Length` alone, since a server can omit or misstate it.

Spec §8.6 permits a dev/test-only exception: an AS running on loopback may fetch loopback addresses. Authgear implements this gated on `DEV_MODE`, the existing process-wide environment variable (default `false`) already used elsewhere to distinguish a local development run from a deployed instance — not a per-project `authgear.yaml` setting, so no project admin can enable it. When `DEV_MODE` is `false` (always the case in production), the rules above apply unconditionally.

The same rules apply to fetching `logo_uri` ([§8.8](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.8)) and, should confidential CIMD clients be added later, `jwks_uri`.

### Authgear as an SSRF/Probing Oracle

Because the target of the fetch is attacker-chosen, an attacker can use `/oauth2/authorize?client_id=https://internal-or-arbitrary-host/path` as a blind probe of what Authgear's network can reach, even with address filtering in place (e.g. distinguishing "connection refused" vs "timeout" vs "TLS handshake failure" vs "valid JSON but failed validation" tells an attacker something about a target that isn't Authgear itself). Authgear should:

- Return the **same** generic unresolvable-client error for every failure mode (blocked address, timeout, non-2xx, invalid JSON, failed validation) — do not let `error_description` vary by failure reason.
- Accept that timing differences are a harder side channel to fully close; address filtering is the real control, the uniform error is defense in depth against information disclosure, not connection prevention.

### Denial of Service via Attacker-Chosen Fetch Targets

A `client_id` can point at any internet host, and caching is keyed on the exact URL string, not the hostname, so an attacker can force a fresh fetch on every request while still targeting one host, just by varying the query string. Since resolution only happens at `/oauth2/authorize` (see [Where resolution happens](#where-resolution-happens)), that's the only endpoint these limits need to apply to. Fixed limits, not project-configurable:

| Limit | Value | Prevents |
| --- | --- | --- |
| Per project (`app_id`) | 10 per minute | One project (or an attacker abusing it) exceeding a reasonable total, regardless of target |
| Per caller IP | 20 per minute | One caller spreading abuse across many different projects |
| Per (project, host) pair | 5 per minute | One project's traffic concentrating entirely on a single external host; scoped per-project so it never throttles an unrelated project sharing that host |

### Domain Trust

Per [spec §8.9](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.9), an authorization server MAY apply its own policy on which domains it trusts as `client_id`s. Authgear exposes this as the optional `allowed_domains` allowlist (empty = no restriction beyond the protections above). This is likely to matter more for higher-trust deployments than for the MCP use case, where the whole point is accepting arbitrary client domains.

### Phishing Mitigation

Per [spec §8.5](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-8.5), the consent screen displays the **hostname** of the `client_id` URL prominently, in addition to `client_name`/`logo_uri` from the fetched document — the hostname is the one piece of information that isn't self-asserted by the document itself (it's implied by the URL that was actually reachable and validated).

### Privacy Considerations

Two considerations from [spec §9](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-9) that this design already addresses through choices made for other reasons:

- **Authorization Server Fetch Side Channel** ([§9.1](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-9.1)): fetching a CIMD document tells the document's hosting operator that _someone_ is going through an authorization flow at Authgear, at the moment of the fetch. Caching a resolved document for reuse — already required for other reasons (spec §5.2) — reduces how often this signal is generated: a document fetched once per cache lifetime leaks far less than one fetched on every authorization request.
- **URLs Referenced in Client ID Metadata Documents** ([§9.2](https://www.ietf.org/archive/id/draft-ietf-oauth-client-id-metadata-document-02.html#section-9.2)): if Authgear rendered `logo_uri` directly in the end user's browser, that browser would fetch it straight from the client's own server, leaking the user's IP and browser fingerprint to that third party as a side effect of merely viewing a consent screen — a tracking-pixel-shaped risk. This is exactly what the prefetch-and-cache handling of `logo_uri` already described in [SSRF Protection](#ssrf-protection) (citing spec §8.8) avoids: the end user's browser only ever talks to Authgear, never directly to the client's server.

### Access Token Audience Binding

CIMD clients use the existing `allow_dynamic_third_party_client_access` policy to gain access to a Resource/Scope's audience. See [api-resource.md — Access Policy](./api-resource.md#access-policy) for the config. No new flag is introduced.

## Reading a CIMD Client's Stored Config

CIMD clients are returned by [DCR's `dynamicClients` query](./dcr.md#new-query) alongside DCR-registered clients — no separate query is needed, since both are now backed by a real, deduplicated, per-`client_id` record (see [Where resolution happens](#where-resolution-happens)). A CIMD client is distinguished from a DCR client via `source: CIMD` on the unified `OAuthClient` model (see [client.md](./client.md#graphql-type)); `registeredAt` stays `null` (there is no registration event, only a resolution) and `lastFetchedAt` carries the freshness signal DCR clients don't have.

The Authorized Apps settings page reads the same persisted record for display (`client_name`, `logo_uri`), so no separate end-user mechanism is needed either.
