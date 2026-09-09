# SSRF Protection

Authgear fetches URLs it did not choose: webhook targets, an SSO provider's discovery endpoint, a CIMD `client_id`. Without a restriction those reach whatever the server can route to — cloud metadata, internal APIs, another tenant.

## Restricted fetches

| Fetch | URL comes from |
| --- | --- |
| Blocking / non-blocking event webhooks | `hook.blocking_handlers[].url`, `hook.non_blocking_handlers[].url` |
| Custom SMS provider | `messaging.custom_sms_provider.url` |
| Account migration hook | `account_migration.hook.url` |
| Phone number verification hook | `proof_of_phone_number_verification.hook.url` |
| OIDC discovery document | an SSO provider's `discovery_document_endpoint` |
| JWKs | the `jwks_uri` that document names |
| CIMD document and logo | the `client_id` — see [cimd.md](./cimd.md) |

Discovery is two hops, and both are restricted: the `jwks_uri` is chosen by whatever answered the first hop.

The portal's test-SMS action takes its URL from the mutation input and is restricted on the same terms.

## Unrestricted fetches

The Deno hook runner (`DENO_ENDPOINT`), object storage, and the SMS/captcha/analytics vendors. This deployment configures where they point, and they legitimately address internal hosts, so the address rules do not apply to them.

## Address rules

- Resolve the hostname **once**, and connect only to an address from that resolution. Re-resolving at connect time is open to DNS rebinding.
- Reject the hostname if **any** A/AAAA record is not publicly routable, rather than filtering to the routable subset.
- Follow no redirects; a redirect target has been through none of the above.
- Use no proxy; a proxy would resolve the name itself.

IPv4-mapped IPv6 is unmapped first, so `::ffff:127.0.0.1` is refused as loopback.

"Publicly routable" excludes loopback, private, link-local, multicast, unspecified, and the other special-use ranges of RFC 6890.

## Response limit

A restricted fetch reads at most **2 MiB** of response body, counted after decompression. Past that the fetch fails the way a malformed response does. Not configurable.

A fetch with a tighter limit of its own keeps it: a CIMD document is capped at 5120 bytes by the OAuth spec, and a CIMD logo at 256 KiB.

## Configuration

`authgear.features.yaml`, which the portal and the Admin API refuse to write. Only an operator can set either.

```yaml
http:
  insecure_fetch_address_allowed: false
  insecure_fetch_address_allowed_hosts: []
```

### `insecure_fetch_address_allowed`

Boolean, default `false`. Lifts the address rules for every fetch above. For test and local-development deployments; prefer the allowlist otherwise.

### `insecure_fetch_address_allowed_hosts`

List of hostnames, default empty. Exempts those hosts from the address rules and leaves the rules in force everywhere else.

```yaml
http:
  insecure_fetch_address_allowed_hosts:
    - hooks.internal.example.com
    - "*.svc.cluster.local"
```

An entry is a hostname — never host:port — matched case-insensitively against the host as configured, before resolution. A leading `*.` matches **exactly one** label: `*.example.com` matches `a.example.com`, not `a.b.example.com` and not `example.com`. A single-label hostname such as `localhost` is valid. Same matching rule as CIMD's `allowed_domains`.

An entry is a statement about the **destination**, not a grant to a caller: it holds however the URL was chosen, including a CIMD `client_id` chosen by an anonymous caller. **Only list a host where an arbitrary fetch to any path on it, triggered by anyone, is harmless.**

An exempt host is not re-checked after resolution, so it is not protected against rebinding either. List only hosts this deployment controls.

Which layer to set it at:

| Layer | Meaning |
| --- | --- |
| App | This project may reach the host. Use this on a multi-tenant deployment. |
| Cluster / plan | Every project, including future sign-ups, may reach the host. Only where the operator trusts every project admin. |

Omitting the field inherits the layer below; an explicit `[]` clears what that layer set.

## Logging

A refusal logs once at `ERROR`:

```
outbound fetch blocked by SSRF address policy
  blocked_by_ssrf_policy=true
  sink=hook.non_blocking_handlers
  host=hooks.internal.example.com
  flag=http.insecure_fetch_address_allowed
```

`blocked_by_ssrf_policy` appears on no other message, so it is a complete filter. `sink` names the config field that chose the URL.

A refusal is reported as its own message rather than as another delivery failure, so that "we refused to call it" and "it was unreachable" are distinguishable.

`host` is subject to the deployment's log masking, so a host containing a configured secret value is masked.
