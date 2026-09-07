# CIMD Part 2 — SSRF-Safe Document Fetching & Metadata Validation

Spec: [docs/specs/cimd.md — The Client Metadata Document](../../specs/cimd.md#the-client-metadata-document), [Accepted Metadata Fields](../../specs/cimd.md#accepted-metadata-fields), [SSRF Protection](../../specs/cimd.md#ssrf-protection), [Authgear as an SSRF/Probing Oracle](../../specs/cimd.md#authgear-as-an-ssrfprobing-oracle).

Depends on [Part 1](2026-08-28-01-config-and-client-id.md) (`oauthclient.ParseCIMDClientID`, the `oauth.client_id_metadata_document` config section).

## 1. Goal / Scope

This part builds the new `pkg/lib/cimd` package's two pure-ish halves:

1. **`Fetcher`** — an HTTP GET against an attacker-chosen URL, hardened against SSRF and DNS rebinding, with fixed size/timeout/redirect limits.
2. **`ParseAndValidate`** — the [Accepted Metadata Fields](../../specs/cimd.md#accepted-metadata-fields) rules over the fetched bytes.

Both are callable and unit-testable on their own. Nothing in this part is wired into any HTTP handler; that is [Part 3](2026-08-28-03-authorize-time-resolution.md).

**The central finding:** the repo has no SSRF-safe HTTP client. `httputil.NewExternalClient` carries the comment `// SECURITY(http): prevent SSRF` but its only actual control is "don't follow redirects" — there is no address filtering and no DNS pinning anywhere in `pkg/`. `urlutil.ValidateHTTPSStrict` filters *hostnames* by blocklist, which is a different (and, against a resolved-address attack, ineffective) control: `https://attacker.example.com/x` passes it and can still resolve to `169.254.169.254`. Every existing outbound-HTTP target in the codebase is admin-configured (webhooks, SSO providers, SMS gateways), so this gap has never mattered before. A CIMD `client_id` arrives on an **unauthenticated query parameter**, so it does now.

That transport is built here as `cimd.SafeTransport`. It is left inside `pkg/lib/cimd` rather than promoted to `pkg/util/httputil` — see §7 D1.

## 2. `pkg/lib/cimd/fetcher.go`

### 2.1 Fixed limits

```go
package cimd

const (
	// MaxDocumentBytes is draft-ietf-oauth-client-id-metadata-document-02
	// §8.7's recommended maximum. Enforced progressively while reading the
	// body (§2.4), never from Content-Length, which a server may omit or
	// misstate.
	MaxDocumentBytes = 5120

	// FetchTimeout covers DNS resolution, TLS handshake and reading the
	// response body, because it is applied as http.Client.Timeout. It
	// matches the existing blocking-webhook per-call default
	// (hook.sync_hook_timeout_seconds) and every other
	// httputil.NewExternalClient call site in this repo, all of which use 5s.
	FetchTimeout = 5 * time.Second

	// RefetchInterval bounds how often a persisted CIMD row is refreshed.
	// Consumed in Part 3; declared here so every fixed CIMD limit is in one
	// place.
	RefetchInterval = 1 * time.Hour
)
```

None of these is project-configurable. Spec § Configuration says so outright, and the reason is worth keeping in the code comment: they are not tuning knobs, they are the bounds on what an unauthenticated caller can make this server do to a third party. A project owner raising `MaxDocumentBytes` or the timeout raises the cost that *other people's* infrastructure pays.

### 2.2 Error strategy — specific internally, collapsed at one boundary

Spec § Authgear as an SSRF/Probing Oracle requires that an attacker cannot distinguish "connection refused" from "timeout" from "TLS handshake failure" from "valid JSON but failed validation" — otherwise `/oauth2/authorize` is a blind network scanner pointed at Authgear's egress.

An earlier draft enforced that by having `Fetcher.Fetch` and `ParseAndValidate` each return a single `ErrFetchFailed`, wrapping the cause with `%v` rather than `%w` so `errors.As` could not reach it. That works but spreads the invariant across the package as a discipline, and it makes the two functions hard to unit-test (every failure looks the same to the test too). Replace it:

- **`Fetcher.Fetch` and `ParseAndValidate` return specific errors.** They are package internals composed by `Service`, and their tests assert exactly which rule fired.
- **`cimd.Service.EnsureClientResolved` is the one and only place that collapses.** It logs the concrete cause and returns a single error for every fetch or validation failure ([Part 3](2026-08-28-03-authorize-time-resolution.md) §3.2).

The invariant then becomes a property of one function rather than a convention, and it is reviewable in one sentence: *nothing but `Service` returns a CIMD error to a caller outside the package.* The `%v`-not-`%w` subtlety disappears entirely, because no cause is ever attached to the error that leaves.

```go
var (
	ErrResponseNotOK   = errors.New("cimd: response status is not 2xx")
	ErrResponseTooLarge = errors.New("cimd: response exceeds the maximum document size")
)
```

Timing differences remain a side channel and are not closed here; spec § Authgear as an SSRF/Probing Oracle accepts this ("address filtering is the real control, the uniform error is defense in depth against information disclosure, not connection prevention").

**On the boundary errors being `apierrors`:** `Service`'s two exported outcomes *are* `apierrors.Kind`s — see [Part 3](2026-08-28-03-authorize-time-resolution.md) §3.2. That is safe precisely because every fetch and validation failure collapses to **one** Kind with one reason and one message: rendering it leaks nothing, since every failure mode renders identically. It would not be safe to give each failure mode its own Kind, which is why the internals above stay plain sentinels.

### 2.3 `pkg/lib/cimd/addrfilter.go` — the address policy

Go's `netip.Addr` predicates already cover most of RFC 6890's special-use space correctly, including the 4-in-6 forms, so this file does not reimplement them — it adds only what they miss. Verified coverage:

| Range | Covered by | 
|---|---|
| `127.0.0.0/8`, `::1`, `::ffff:127.0.0.1` | `IsLoopback()` |
| `10.0.0.0/8`, `172.16.0.0/12`, `192.168.0.0/16` (RFC 1918) | `IsPrivate()` |
| `fc00::/7` (RFC 4193, incl. `fd00::/8`) | `IsPrivate()` |
| `169.254.0.0/16` (**cloud metadata**), `fe80::/10` | `IsLinkLocalUnicast()` |
| `224.0.0.0/4`, `ff00::/8`, `ff01::/16`, `ff02::/16` | `IsMulticast()` / `IsLinkLocal…` / `IsInterfaceLocalMulticast()` |
| `0.0.0.0`, `::` | `IsUnspecified()` |
| `100.64.0.0/10` (CGNAT) | **table below** |
| `192.0.0.0/24`, `192.0.2.0/24`, `198.18.0.0/15`, `198.51.100.0/24`, `203.0.113.0/24` | **table below** |
| `240.0.0.0/4`, `255.255.255.255/32` | **table below** |
| `2001:db8::/32`, `100::/64` | **table below** |
| `2002::/16` (6to4), `64:ff9b::/96` + `64:ff9b:1::/48` (NAT64), `2001::/32` (Teredo) | **table below** — each embeds an arbitrary IPv4 address, including a private one |

So RFC 1918 and `fc00::/7` **are** blocked; they are just blocked by `IsPrivate()` rather than by the table. Confirm with a throwaway program over `netip.Addr` before implementing, and keep the §6 test asserting the *union* rather than either mechanism — that test is what stops the split from silently drifting.

**Do not "simplify" this to `IsGlobalUnicast()`.** It is the obvious-looking shortcut and it is wrong: it returns **true** for `10.0.0.1`, `172.16.0.1`, `fc00::1`, `100.64.0.1`, `192.0.2.1` and `240.0.0.1`. It excludes only unspecified, loopback, multicast and link-local, so as an "is this public" test it admits every RFC 1918 address. Reject this suggestion if it comes up in review.

```go
// additionalBlockedPrefixes holds only what netip.Addr's own predicates do
// NOT cover; see the coverage table in the plan. RFC 6890 is the umbrella
// reference for spec §8.6.
var additionalBlockedPrefixes = []netip.Prefix{
	netip.MustParsePrefix("100.64.0.0/10"),      // RFC 6598 CGNAT
	netip.MustParsePrefix("192.0.0.0/24"),       // RFC 6890 IETF protocol assignments
	netip.MustParsePrefix("192.0.2.0/24"),       // RFC 5737 TEST-NET-1
	netip.MustParsePrefix("198.18.0.0/15"),      // RFC 2544 benchmarking
	netip.MustParsePrefix("198.51.100.0/24"),    // RFC 5737 TEST-NET-2
	netip.MustParsePrefix("203.0.113.0/24"),     // RFC 5737 TEST-NET-3
	netip.MustParsePrefix("240.0.0.0/4"),        // RFC 1112 reserved
	netip.MustParsePrefix("255.255.255.255/32"), // limited broadcast
	netip.MustParsePrefix("2001:db8::/32"),      // RFC 3849 documentation
	netip.MustParsePrefix("2002::/16"),          // RFC 3056 6to4
	netip.MustParsePrefix("64:ff9b::/96"),       // RFC 6052 NAT64
	netip.MustParsePrefix("64:ff9b:1::/48"),     // RFC 8215 local-use NAT64
	netip.MustParsePrefix("100::/64"),           // RFC 6666 discard-only
	netip.MustParsePrefix("2001::/32"),          // RFC 4380 Teredo
}

// Unmap first: the table is written in native v4 form, so Contains would
// miss ::ffff:10.0.0.1 without it.
func IsPubliclyRoutable(addr netip.Addr) bool {
	if !addr.IsValid() {
		return false
	}
	a := addr.Unmap()

	switch {
	case a.IsLoopback(),
		a.IsPrivate(),
		a.IsLinkLocalUnicast(),
		a.IsLinkLocalMulticast(),
		a.IsInterfaceLocalMulticast(),
		a.IsMulticast(),
		a.IsUnspecified():
		return false
	}

	for _, p := range additionalBlockedPrefixes {
		if p.Contains(a) {
			return false
		}
	}
	return true
}
```

### 2.4 `SafeTransport` — resolve once, validate all, connect to a validated address

`SafeDialer` implements spec § SSRF Protection's first two rules together, because neither works without the other:

- *"Resolve the hostname once per fetch attempt and connect only to an address validated in that same resolution."* A second, independent resolution at connect time is a DNS-rebinding hole — the attacker's nameserver answers with a public address for the validation and a special-use one for the connect.
- *"Check every address a hostname resolves to, not just the first."* Reject the whole hostname if **any** A/AAAA record is non-publicly-routable. Rejecting the hostname rather than skipping the bad records is deliberate: a mixed answer is an attack signature, not a misconfiguration worth accommodating.

This is why the dial cannot just be `net.Dialer` with a `Control` hook — `Control` validates whatever the OS resolver returned for *that* connect, which is a second resolution. `Control` is still installed, as a second line of defence on the concrete syscall address; by construction it should never fire, and it exists so a future refactor that reintroduces hostname-based dialling cannot silently reopen the hole.

`AllowNonPublicAddresses` carries `insecure_fetch_address_allowed` (Part 1 §2.4) as a plain field rather than a config read, so the dialer stays a pure function of its inputs and both postures are trivially testable. There is no `DevMode` field — Part 1 D12 replaced that gate outright, so exactly one mechanism can widen this policy.

```go
type SafeDialer struct {
	Resolver                *net.Resolver // nil means net.DefaultResolver
	AllowNonPublicAddresses bool
}

func (d *SafeDialer) DialContext(ctx context.Context, network, addr string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	portNum, err := strconv.Atoi(port)
	if err != nil {
		return nil, err
	}

	// An IP-literal host never goes through DNS. Part 1 D4 lets an IP-literal
	// client_id through shape validation; this is where it is refused.
	if literal, err := netip.ParseAddr(host); err == nil {
		if !d.allow(literal) {
			return nil, errBlockedAddress
		}
		return d.dial(ctx, network, netip.AddrPortFrom(literal, uint16(portNum)))
	}

	resolver := d.Resolver
	if resolver == nil {
		resolver = net.DefaultResolver
	}
	addrs, err := resolver.LookupNetIP(ctx, "ip", host)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errBlockedAddress
	}
	for _, a := range addrs {
		if !d.allow(a) {
			return nil, errBlockedAddress
		}
	}

	// Connect to an address from THAT resolution. Each is already validated;
	// trying them in order keeps a dual-stack host with an unreachable AAAA
	// working.
	var lastErr error
	for _, a := range addrs {
		conn, err := d.dial(ctx, network, netip.AddrPortFrom(a, uint16(portNum)))
		if err == nil {
			return conn, nil
		}
		lastErr = err
	}
	return nil, lastErr
}

func (d *SafeDialer) allow(addr netip.Addr) bool {
	if IsPubliclyRoutable(addr) {
		return true
	}
	return d.AllowNonPublicAddresses
}

// Control re-validates the concrete syscall address as defence in depth.
func (d *SafeDialer) dial(ctx context.Context, network string, ap netip.AddrPort) (net.Conn, error) {
	dialer := &net.Dialer{
		Timeout: FetchTimeout,
		Control: func(network, address string, c syscall.RawConn) error {
			parsed, err := netip.ParseAddrPort(address)
			if err != nil {
				return err
			}
			if !d.allow(parsed.Addr()) {
				return errBlockedAddress
			}
			return nil
		},
	}
	return dialer.DialContext(ctx, network, ap.String())
}
```

Only `network` values `"tcp"`, `"tcp4"` and `"tcp6"` occur here.

**`AllowNonPublicAddresses` widens the policy to every non-publicly-routable range, not just loopback** — including `169.254.169.254`. That is required rather than sloppy: the e2e document host is reached at a container-network RFC 1918 address, and a developer running the document host in Docker against a host-machine Authgear is in the same position, so a loopback-only exception would not work for either. The containment is everything listed in Part 1 §2.4 — Site-Admin-only, per project, `insecure_`-named, default false, never at the cluster or plan layer, `Warn`-logged on every use (§2.5.1). It is a genuine per-project SSRF kill switch, in the same way `config.RateLimitsFeatureConfig.Disabled` is a genuine per-project rate-limit kill switch.

**TLS still verifies the hostname.** `http.Transport` derives `tls.Config.ServerName` from the request URL's host, not from whatever `DialContext` connected to, so overriding `DialContext` with a pre-resolved IP does *not* weaken certificate verification or SNI. Do not set `TLSClientConfig` at all — leaving it nil keeps the platform root pool and full verification. §6 asserts this: a server presenting a certificate for the wrong name must fail.

### 2.5 The `Fetcher` itself

```go
type Fetcher struct {
	HTTPClients *CIMDHTTPClients
	// OAuthFeatureConfig supplies insecure_fetch_address_allowed (Part 1
	// §2.4). Already fanned out by wire (pkg/lib/deps/deps_config.go:74).
	OAuthFeatureConfig *config.OAuthFeatureConfig
	// AppID is read only by clientFor's warning log (§2.5.1).
	AppID config.AppID
}

// newCIMDHTTPClient builds one of the two http.Clients used for CIMD
// fetches. It reuses httputil.NewExternalClientWithOptions so the fetch is
// otel-instrumented and non-redirect-following on the same code path as
// every other outbound call in the repo, and supplies the SSRF-safe
// transport underneath.
//
// FollowRedirect: false gives CheckRedirect = ErrUseLastResponse, so a 3xx
// is returned as a response rather than followed; the 2xx check in Fetch
// then rejects it. Spec § SSRF Protection: "Follow 0 redirects -- a redirect
// target hasn't been through Client ID Format validation, and would
// otherwise let the previous two rules be bypassed."
// Two clients are built, not one, because the address policy is a property
// of the transport and the transport is process-level (connection pooling).
// Selecting between two named clients keeps the decision in ONE greppable
// place (Fetcher.clientFor, §2.5.1) instead of threading a boolean into the
// dial path on every call, and it means the permissive transport is a
// distinct object a reviewer can search for.
type CIMDHTTPClients struct {
	Strict   *http.Client
	Insecure *http.Client
}

func ProvideCIMDHTTPClients() *CIMDHTTPClients {
	return &CIMDHTTPClients{
		Strict:   newCIMDHTTPClient(false),
		Insecure: newCIMDHTTPClient(true),
	}
}

func newCIMDHTTPClient(allowNonPublicAddresses bool) *http.Client {
	dialer := &SafeDialer{AllowNonPublicAddresses: allowNonPublicAddresses}
	transport := &http.Transport{
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConnsPerHost:   2,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   FetchTimeout,
		ResponseHeaderTimeout: FetchTimeout,
		// No Proxy. http.ProxyFromEnvironment would route the request
		// through a proxy chosen by the environment, and the proxy -- not
		// this dialer -- would then do the name resolution, silently
		// bypassing every rule above.
		Proxy: nil,
	}
	return httputil.NewExternalClientWithOptions(FetchTimeout, httputil.ExternalClientOptions{
		FollowRedirect: false,
		Transport:      transport,
	})
}

// Fetch GETs the document at u and returns its raw bytes. u must already
// have passed oauthclient.ParseCIMDClientID and the caller's
// allowed_domains check (Part 3 §3). Errors are specific; Service is what
// collapses them (§2.2).
func (f *Fetcher) Fetch(ctx context.Context, u *url.URL) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "Authgear")

	resp, err := f.clientFor(ctx, u).Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// MUST be 2xx. A 3xx lands here too (redirects are not followed) and is
	// refused by this same check.
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("%w: %d", ErrResponseNotOK, resp.StatusCode)
	}

	// Progressive size enforcement: read at most MaxDocumentBytes+1 and
	// refuse if the extra byte materialised. Content-Length is never
	// consulted -- spec § SSRF Protection requires this explicitly, "since a
	// server can omit or misstate it". Reading +1 rather than exactly the
	// limit is what makes "exactly 5120 bytes" acceptable and "5121 bytes"
	// refused, without a separate length probe.
	body, err := io.ReadAll(io.LimitReader(resp.Body, MaxDocumentBytes+1))
	if err != nil {
		return nil, err
	}
	if len(body) > MaxDocumentBytes {
		return nil, ErrResponseTooLarge
	}

	return body, nil
}
```

The response `Content-Type` is deliberately **not** checked — see §7 D3.

`CIMDHTTPClients` is a named wrapper type rather than a bare `*http.Client` so wire can distinguish it from the several other `*http.Client` providers already in the graph (`pkg/portal/deps.go:47`, `pkg/siteadmin/service/deps.go:12`, …). Same pattern as `SiteAdminHTTPClient`.

### 2.5.1 `clientFor` — the one place the escape hatch is consulted

```go
// clientFor selects the strict or the permissive transport for this
// project. This is the ONLY place in the CIMD fetch path that reads
// insecure_fetch_address_allowed; everything else takes a client. A
// reviewer auditing "when can Authgear reach a private address" reads this
// function and nothing else.
func (f *Fetcher) clientFor(ctx context.Context, u *url.URL) *http.Client {
	if !f.OAuthFeatureConfig.GetClientIDMetadataDocument().IsInsecureFetchAddressAllowed() {
		return f.HTTPClients.Strict
	}
	// Required by Part 1 D16. Without this, a flag left set on a deployed
	// project is completely invisible. Volume is bounded to <= 10/min/project
	// by the Part 4 per-project fetch limit, so this cannot flood a log
	// pipeline, and it is deliberately Warn rather than Info so it surfaces
	// in default log configurations.
	logger := FetcherLogger.GetLogger(ctx)
	logger.Warn(ctx, "cimd: fetching with SSRF address protection disabled",
		slog.String("app_id", string(f.AppID)),
		slog.String("host", u.Hostname()),
		slog.String("flag", "oauth.client_id_metadata_document.insecure_fetch_address_allowed"),
	)
	return f.HTTPClients.Insecure
}
```

There is no equivalent warning for `insecure_http_allowed` here, because that flag is consumed in `oauthclient.ParseCIMDClientID` (Part 1 §3), which is a pure function with no logger. Log it instead in `cimd.Service.EnsureClientResolved` right before the fetch, where both the scheme and a logger are in hand ([Part 3](2026-08-28-03-authorize-time-resolution.md) §3.3 step 7) — one `Warn` per insecure fetch attempt, whichever relaxation is responsible.

## 3. `pkg/lib/cimd/document.go` — Accepted Metadata Fields

```go
var (
	ErrDocumentNotJSONObject      = errors.New("cimd: document is not a JSON object")
	ErrDocumentClientIDMismatch   = errors.New("cimd: client_id does not equal the request URL")
	ErrDocumentRedirectURIsMissing = errors.New("cimd: redirect_uris is required")
	ErrDocumentRedirectURIInvalid  = errors.New("cimd: invalid redirect_uri")
	ErrDocumentGrantTypeUnsupported = errors.New("cimd: unsupported grant_type")
	ErrDocumentResponseTypeInconsistent = errors.New("cimd: response_types is inconsistent with grant_types")
	ErrDocumentApplicationTypeUnsupported = errors.New("cimd: unsupported application_type")
	ErrDocumentTokenEndpointAuthMethodNotAccepted = errors.New("cimd: token_endpoint_auth_method is not accepted")
	ErrDocumentURIFieldNotHTTPS = errors.New("cimd: uri field must use https")
)

// rawDocument is the wire shape. Every field the spec says to reject or
// ignore -- client_secret, client_secret_expires_at, jwks_uri,
// software_statement, and any unknown property -- is simply absent from
// this struct, so encoding/json drops it. There is no need to name them
// and no DisallowUnknownFields: spec § Validation says "Unrecognized
// properties are ignored (the spec explicitly allows additional
// properties)", and spec §4.1 says credential material is "always ignored",
// not "rejected". A document carrying a client_secret is therefore VALID
// and its secret is discarded -- do not turn this into an error.
type rawDocument struct {
	ClientID                *string  `json:"client_id"`
	ClientName              *string  `json:"client_name"`
	RedirectURIs            []string `json:"redirect_uris"`
	GrantTypes              []string `json:"grant_types"`
	ResponseTypes           []string `json:"response_types"`
	ApplicationType         *string  `json:"application_type"`
	LogoURI                 *string  `json:"logo_uri"`
	ClientURI               *string  `json:"client_uri"`
	TOSURI                  *string  `json:"tos_uri"`
	PolicyURI               *string  `json:"policy_uri"`
	TokenEndpointAuthMethod *string  `json:"token_endpoint_auth_method"`
}

// Document is a rawDocument after defaults have been applied and every
// field has passed validation. Field-for-field it is the CIMD counterpart
// of dcr.NormalizedRegistration, and for the same reason: the caller
// (Part 3) copies it straight onto oauthclient.NewClientOptions.
type Document struct {
	ClientName    *string
	RedirectURIs  []string
	GrantTypes    []string
	ResponseTypes []string
	// ApplicationType is always "web" or "native" -- never nil, never
	// anything else.
	ApplicationType string
	LogoURI         *string
	ClientURI       *string
	TOSURI          *string
	PolicyURI       *string
}
```

```go
// ParseAndValidate implements every rule in
// docs/specs/cimd.md#accepted-metadata-fields. requestURL is the exact
// client_id string that was fetched -- NOT a re-serialized url.URL, because
// the client_id equality check below is byte-for-byte.
//
// It returns the specific sentinel for whichever rule failed. Service
// collapses these into its single unresolvable error, so spec § Error
// Handling's "a document that fails any MUST-level check is treated as if
// the fetch had failed" holds at that boundary rather than here (§2.2).
// allowInsecureHTTP comes from
// OAuthClientIDMetadataDocumentFeatureConfig.IsInsecureHTTPAllowed() (Part 1
// §2.4) and relaxes rule 8's https requirement on the document's URI fields
// only. Like ParseCIMDClientID's equivalent parameter, it is explicit rather
// than read from config in here, so this function stays pure and every call
// site states its posture.
func ParseAndValidate(requestURL string, body []byte, allowInsecureHTTP bool) (*Document, error) {
	var raw rawDocument
	// Guard against a JSON array/string/number/null top level, which would
	// unmarshal into a struct with an error, and against a valid object
	// whose every field is absent (which is caught by the redirect_uris
	// check below anyway).
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	...
}
```

Rule by rule, in the order they are checked:

| # | Rule | Detail |
|---|---|---|
| 1 | `token_endpoint_auth_method`, if present, MUST be `"none"` | Checked first, mirroring `dcr.ValidateAndNormalize`. Rejects `private_key_jwt` and every `client_secret_*` variant. Spec § Client Authentication puts confidential CIMD clients out of scope for v1, and spec §4.1 forbids shared-secret methods for CIMD clients permanently. Default when absent: `none`. |
| 2 | `client_id` MUST be present and MUST equal `requestURL` **byte-for-byte** | `raw.ClientID == nil \|\| *raw.ClientID != requestURL` → `ErrDocumentClientIDMismatch`. No normalization, no case folding, no trailing-slash tolerance. This single comparison is the whole reason one host cannot vouch for another's identity, so it must stay a plain `!=`. Part 1's `ParseCIMDClientID` deliberately does not rewrite the string precisely so that this check is meaningful. |
| 3 | `application_type`, if present, MUST be `"web"` or `"native"` | Default `"web"`. |
| 4 | `redirect_uris` MUST be present and non-empty; each entry validated | See §3.1. **`application_type` is not consulted.** |
| 5 | `grant_types` MUST be a subset of `["authorization_code", "refresh_token"]` | Default `["authorization_code", "refresh_token"]`. |
| 6 | `response_types` MUST be a subset of `["code"]` | Default `["code"]`. |
| 7 | `response_types` MUST be consistent with `grant_types` | Same rule as DCR (`dcr/validate.go:106-111`): `contains(grantTypes, "authorization_code") == contains(responseTypes, "code")`. |
| 8 | `logo_uri`, `client_uri`, `tos_uri`, `policy_uri`, if present, MUST be `https://` | Same as DCR's `validateHTTPSURI` — except that `http://` is accepted when `insecure_http_allowed` is set (Part 1 §2.4, D13a), which is why `ParseAndValidate` takes that flag as its third parameter. Nothing else about these fields is relaxed. |
| 9 | `client_name` optional, no constraint | The `Client <clientID>` fallback is **not** applied here; it is computed on read by `oauthclient.Client.DisplayName()`, which already implements exactly the same rule (`client.go:66-71`) and which spec § Accepted Metadata Fields calls "same fallback as DCR". An empty-string `client_name` is normalized to nil by `Store.NewClient`, also already existing. |

### 3.1 `redirect_uris` — the one place CIMD diverges from DCR

```go
// validateCIMDRedirectURI implements spec § Accepted Metadata Fields --
// redirect_uris. Each entry must be an absolute URI with no fragment, and
// must be one of:
//
//   - https://
//   - a loopback http:// URI -- host localhost, 127.0.0.1 or [::1], ANY port
//   - a custom (non-http, non-https) URI scheme
//
// Unlike dcr.validateRedirectURI, application_type is NOT a parameter.
// DCR gates http://localhost on application_type: native; CIMD cannot,
// because the MCP Authorization spec's own reference CIMD document uses
// http://127.0.0.1:3000/callback and http://localhost:3000/callback while
// omitting application_type entirely (so it defaults to "web"). Gating
// loopback the DCR way would reject the reference example outright. Spec §
// Accepted Metadata Fields states this divergence explicitly.
func validateCIMDRedirectURI(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || !u.IsAbs() {
		return ErrDocumentRedirectURIInvalid
	}
	if u.Fragment != "" || u.RawFragment != "" {
		return ErrDocumentRedirectURIInvalid
	}
	switch strings.ToLower(u.Scheme) {
	case "https":
		if u.Host == "" {
			return ErrDocumentRedirectURIInvalid
		}
		return nil
	case "http":
		switch u.Hostname() {
		case "localhost", "127.0.0.1", "::1":
			return nil
		default:
			return ErrDocumentRedirectURIInvalid
		}
	default:
		// Any custom scheme (com.example.app:/callback, myapp://cb, ...).
		return nil
	}
}
```

`url.URL.Hostname()` strips the brackets from `[::1]:3000`, so the `"::1"` case matches `http://[::1]:3000/callback`. Accepting `::1` alongside the two forms the spec names is a deliberate small superset (RFC 8252 §7.3 treats both IPv4 and IPv6 loopback as loopback, and an IPv6-only developer machine has no `127.0.0.1` listener) — flagged in §7 D4 and in the spec-updates list.

**A note the reviewer will want.** After validation, `application_type` has **no runtime effect at all** for a CIMD client. Spec § Accepted Metadata Fields says it "controls redirect URI rules as in DCR", but the very next paragraph removes the only rule it controlled, and `oauthclient.Client.ToClientConfig` maps `Kind == THIRD_PARTY` to `OAuthClientApplicationTypeDynamicThirdParty` *before* it ever looks at `ApplicationType` — and every CIMD client is `THIRD_PARTY`. So the value is validated, persisted, and reported through `OAuthClient.applicationType`, and that is all it does. This is not a bug to fix; it is spec text to correct (§8).

## 4. Wiring — `pkg/lib/cimd/deps.go`

```go
package cimd

var DependencySet = wire.NewSet(
	ProvideCIMDHTTPClients,
	wire.Struct(new(Fetcher), "*"),
)
```

Add `cimd.DependencySet` to `pkg/lib/deps/deps_common.go`'s set alongside `dcr.DependencySet` and `oauthclient.DependencySet`, then `make generate`.

`ProvideCIMDHTTPClients` takes no arguments — the policy is selected per request in `clientFor`, not baked into the provider. `*config.OAuthFeatureConfig` and `config.AppID` are both already in the request-scoped graph (`pkg/lib/deps/deps_config.go:74` and the app-config deps respectively), so `Fetcher` needs no new provider either.

**Note the two clients are process-level while the flag is per-request.** That is fine and intended: the clients are stateless connection pools, and no request can reach the permissive one without passing `clientFor`'s check. It does mean the permissive transport exists in memory in every deployment; a reviewer who objects should be pointed at the alternative, which is building an `http.Client` per request and losing connection reuse for no security gain, since the gate is the same either way.

Nothing consumes `Fetcher` yet; wiring it here keeps Part 3's diff to the resolution logic. If wire prunes an unreferenced provider, defer this section's edit to Part 3 commit 1 — verify with `make generate` at the end of §9 commit 4 and move it if the generated file does not change.

## 5. What this part does *not* do

- **No `logo_uri` fetching.** Spec § SSRF Protection says "the same rules apply to fetching `logo_uri`", and spec § Privacy Considerations §9.2 describes a prefetch-and-cache as though it already existed. Neither is true today: nothing in the repo fetches a client's `logo_uri`, and the consent screen does not render a client logo at all (the only `LogoURI` use is `__brand_logo.html:22`, gated on `ReplaceProjectLogo`, which is fixed `false` for every dynamic client). [Part 6](2026-08-28-06-consent-and-authorized-apps.md) renders `logo_uri` directly and [Part 7](2026-08-28-07-logo-proxy.md) replaces that with a server-side proxy through this same `Fetcher`. The spec text is corrected in Part 7.
- **No `jwks_uri` fetching.** Confidential CIMD clients are out of scope for v1 (spec § Client Authentication). `jwks_uri` is ignored, not rejected.
- **No caching.** `Fetcher.Fetch` always performs a request. The refetch interval, single-flight and persistence are Part 3's.

## 6. Test Plan

**Unit — `pkg/lib/cimd/addrfilter_test.go`**

Table-driven over `IsPubliclyRoutable`. Must-reject: `127.0.0.1`, `127.1.2.3`, `::1`, `::ffff:127.0.0.1`, `10.0.0.1`, `172.16.0.1`, `192.168.1.1`, `fc00::1`, `fd12::1`, `169.254.169.254`, `169.254.0.1`, `fe80::1`, `224.0.0.1`, `ff02::1`, `0.0.0.0`, `::`, `100.64.0.1`, `192.0.2.1`, `198.18.0.1`, `198.51.100.1`, `203.0.113.1`, `240.0.0.1`, `255.255.255.255`, `2001:db8::1`, `2002::1`, `64:ff9b::7f00:1`, `2001::1`, and the zero `netip.Addr{}`. Must-accept: `1.1.1.1`, `8.8.8.8`, `93.184.216.34`, `2606:4700::1111`, `::ffff:1.1.1.1`.

The `::ffff:127.0.0.1` case is the one that regresses if someone deletes the `Unmap()`; give it its own named test so the failure message says so.

**Unit — `pkg/lib/cimd/fetcher_test.go`**

Uses `httptest.NewTLSServer` plus a `SafeDialer` whose `Resolver` is a stub, so DNS behavior is controllable without touching the network. Cases:

| Case | Expect |
|---|---|
| 200 + valid JSON within the limit | bytes returned |
| exactly 5120 bytes | accepted |
| 5121 bytes | `ErrResponseTooLarge` |
| `Content-Length: 10` but 6000 bytes sent | `ErrResponseTooLarge` (proves Content-Length is not trusted) |
| chunked response with no `Content-Length`, 6000 bytes | `ErrResponseTooLarge` |
| 301 to a valid document | `ErrResponseNotOK` (0 redirects), and the redirect target is **never requested** — assert via a hit counter on the second server |
| 404, 500, 204 | `ErrResponseNotOK` |
| server that sleeps 10s | a timeout error, and the call returns in well under 10s |
| resolver returns `[1.1.1.1, 127.0.0.1]` | `errBlockedAddress`, and the dial is **never attempted** (the whole hostname is rejected, not just the bad record) |
| resolver returns `[127.0.0.1]`, `AllowNonPublicAddresses: false` | `errBlockedAddress` |
| resolver returns `[127.0.0.1]`, `AllowNonPublicAddresses: true` | fetch succeeds |
| resolver returns `[10.0.0.1]`, `AllowNonPublicAddresses: true` | fetch succeeds — the flag covers every non-public range, not only loopback (§2.4) |
| resolver returns `[169.254.169.254]`, `AllowNonPublicAddresses: true` | fetch succeeds. Assert this explicitly and name it in the test: it documents, in code, exactly how dangerous the flag is, so nobody enables it casually |
| resolver returns `[169.254.169.254]`, `AllowNonPublicAddresses: false` | `errBlockedAddress` |
| IP-literal host `https://127.0.0.1/x`, `AllowNonPublicAddresses: false` | `errBlockedAddress`, no DNS lookup performed |
| IP-literal host `https://10.0.0.5:2727/x`, `AllowNonPublicAddresses: true` | fetch succeeds — the e2e shape |
| TLS cert for the wrong hostname | an `x509` verification error — proves hostname verification survives the custom `DialContext` |
| every failing case above | a **specific** error, asserted per case. Uniformity is Service's job now (§2.2), and `service_test.go` asserts it there |

The "don't leak the reason" assertion moves to `service_test.go` ([Part 3](2026-08-28-03-authorize-time-resolution.md) §7), where it belongs: every distinct failure mode fed through `EnsureClientResolved` must produce a byte-identical error, and `errors.As` on the result must not reach a `*net.OpError` or any `ErrDocument*`/`ErrResponse*` sentinel.

**Unit — `clientFor` (same file)**

- feature config absent → `Strict` returned, nothing logged;
- `insecure_fetch_address_allowed: false` → `Strict`, nothing logged;
- `insecure_fetch_address_allowed: true` → `Insecure`, and a `Warn` record is emitted containing the `app_id`, the target host and the flag name. Assert on the log output (capture the `slogutil` handler) — an unlogged insecure fetch is the failure mode Part 1 D16 exists to prevent, and it is silent otherwise.

**Unit — `pkg/lib/cimd/document_test.go`**

Modelled on `pkg/lib/dcr/validate_test.go`. The MCP spec's own example document (spec § UC1 Step 2, verbatim) as the primary happy path, asserting: both loopback redirect URIs accepted with **no** `application_type` present, defaults `application_type: "web"`, `grant_types` and `response_types` as given.

Then, one case per rule: missing `client_id`; `client_id` differing only by a trailing slash, only by scheme case, only by host case (all → mismatch); `token_endpoint_auth_method` of `none` (ok), absent (ok), `client_secret_post` / `private_key_jwt` (rejected); missing/empty `redirect_uris`; a `http://evil.com/cb` redirect URI (rejected); `http://localhost:1/cb`, `http://127.0.0.1:65535/cb`, `http://[::1]:3000/cb`, `com.example.app:/cb`, `myapp://cb` (all accepted); `https://x/cb#frag` (rejected); `application_type: "spa"` (rejected); `grant_types: ["client_credentials"]` (rejected); `response_types: ["token"]` (rejected); `grant_types: ["refresh_token"]` with `response_types: ["code"]` (rejected, inconsistent); `logo_uri: "http://x/l.png"` (rejected).

Explicitly assert the **ignore** rules: a document containing `client_secret`, `client_secret_expires_at`, `jwks_uri`, `software_statement` and an arbitrary `x_whatever` property is **valid**, and none of those values appears in the returned `Document`.

Malformed bodies: `[]`, `"str"`, `null`, `123`, truncated JSON, empty body — all error (a `json` unmarshal error is fine; the specific type is not asserted).

**Commands to run**

```
go test ./pkg/lib/cimd/...
make lint
make generate && git status --porcelain   # must be empty
```

## 7. Fixed Behavioral Decisions

- **D1. `SafeDialer`/`IsPubliclyRoutable` live in `pkg/lib/cimd`, not `pkg/util/httputil`.** They are the *only* SSRF-safe transport in the repo, and promoting them to a general util invites the next author to reach for them casually and would put pressure on the address policy to relax for an admin-configured use case. Keeping them next to the one attacker-controlled fetch target keeps the policy strict. If a second attacker-controlled fetch target ever appears, promote them **then**, with both call sites in view.
- **D2. `http.Transport.Proxy` is nil.** `ProxyFromEnvironment` would hand name resolution to a proxy and bypass every rule in §2.4. A deployment that requires an egress proxy for CIMD needs an explicit design, not an env var.
- **D3. The response `Content-Type` is not checked.** Spec § Validation requires only "MUST be `2xx` and MUST parse as a JSON object within the size limit". Requiring `application/json` would reject real static-file hosts that serve `.json` as `text/plain` while adding nothing: the body still has to parse as a JSON object with a matching `client_id`.
- **D4. `http://[::1]` is accepted as a loopback redirect URI** alongside the `localhost` and `127.0.0.1` the spec names. RFC 8252 §7.3 treats it as loopback and an IPv6-only host has no alternative. Spec text to be updated (§8).
- **D5. A document containing credential material is valid; the material is ignored.** Spec §4.1 says "always ignored". Rejecting instead would be a stricter-than-spec behavior that breaks clients publishing one document for several authorization servers.
- **D6. A whole hostname is rejected if *any* resolved address is non-routable**, rather than filtering to the routable subset. A mixed answer is an attack signature, not a configuration to accommodate.
- **D7. The address policy is widened by the per-project `insecure_fetch_address_allowed` feature flag, not by `DEV_MODE`** (Part 1 §2.4, D10–D14). `SafeDialer` has no `DevMode` field. When set, the flag widens the policy to **every** non-publicly-routable range, including `169.254.169.254` — not loopback only — because the e2e and Docker-based local-dev document hosts sit on private container addresses rather than loopback, and a loopback-only exception would not work for either. The containment is that the flag is Site-Admin-only, per project, `insecure_`-named, defaults false, never set at the cluster or plan layer, and `Warn`-logged on every use.
- **D7a. Two process-level clients (`Strict`, `Insecure`), selected per request in `Fetcher.clientFor`.** One place reads the flag; the permissive transport is a distinct, greppable object. Building a client per request would lose connection reuse for no security gain.
- **D8. Errors are specific inside the package and collapsed at exactly one boundary — `Service.EnsureClientResolved`.** §2.2. `Fetcher` and `ParseAndValidate` return real errors so their tests can assert which rule fired; nothing but `Service` returns a CIMD error outside the package. Replaces an earlier `%v`-not-`%w` wrapping trick, which spread the invariant across the package as a discipline and made both functions untestable per-rule.
- **D9. `application_type` is behaviorally inert for CIMD clients.** Validated, stored and reported; it controls nothing, because the redirect-URI rules are uniform and every CIMD client is `THIRD_PARTY`. Spec text to be updated (§8).

## 8. Spec Updates

Small corrections to `docs/specs/cimd.md`, as a `doc:` commit in this part:

1. **§ Accepted Metadata Fields — `application_type`**: replace "controls redirect URI rules as in DCR" with a statement that it is accepted and reported but has no effect on redirect URI validation or on any other runtime behavior, since redirect URI rules are uniform for CIMD and every CIMD client is third-party.
2. **§ Accepted Metadata Fields — `redirect_uris`**: add `http://[::1]` to the loopback forms.
3. **§ SSRF Protection**: the closing line "The same rules apply to fetching `logo_uri` ... and, should confidential CIMD clients be added later, `jwks_uri`" is aspirational for `logo_uri` until Part 7 lands. Leave the sentence but do not let a reader infer that a `logo_uri` fetch exists in v1 — this is corrected properly in Part 7's spec commit, so Part 2 leaves it alone and only §1 and §2 above are edited here.

## 9. Atomic Commit Plan

1. `[CIMD] Add the non-publicly-routable address filter` — §2.3 + `addrfilter_test.go`. Self-contained and heavily tested; reviewable on its own.
2. `[CIMD] Add an SSRF-safe HTTP transport for metadata document fetches` — §2.4 + the dialer half of `fetcher_test.go` (resolution, `AllowNonPublicAddresses` both ways, IP literal, TLS hostname).
3. `[CIMD] Add the client metadata document fetcher` — §2.1, §2.2, §2.5, §2.5.1 + the response half of `fetcher_test.go` (status, size, redirect, timeout) and the `clientFor` tests.
4. `[CIMD] Add client metadata document validation` — §3, §3.1, §4 + `document_test.go`.
5. `doc: Correct cimd.md on application_type and IPv6 loopback redirect URIs` — §8.

Body of each: `ref DEV-XXXX`.
