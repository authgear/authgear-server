# Breaking changes

Changes that can break a **working deployment** when you upgrade. Newest first.

This is not a changelog. A change is listed here only if an upgrade can stop something that used to work, and only with what to do about it.

## How to use this file

Releases are tagged `YYYY-MM-DD.N` — for example `2026-01-01.0`, and `2026-01-01.1` for a second release the same day.

Upgrading from tag **A** to tag **B**: read every section below that is newer than A and no newer than B, oldest first, and act on it before or during the upgrade.

For the full commit list between two releases, which includes everything not listed here:

```sh
make logs-summary A=2026-01-01.0 B=2026-02-01.0
```

## [Unreleased]

### Outbound requests to private and loopback addresses are refused

Authgear now refuses to connect to an address that is not publicly routable when the URL came from a project's own configuration. Loopback, private (`10/8`, `172.16/12`, `192.168/16`), link-local (including `169.254.169.254`), and the other special-use ranges of RFC 6890 are all refused. A hostname is refused if **any** address it resolves to is in one of those ranges.

**Affects** any deployment that points one of these at an address inside its own network:

| Setting | What stops working |
| --- | --- |
| `hook.blocking_handlers[].url` | the auth flow the hook gates fails, so logins or signups break |
| `hook.non_blocking_handlers[].url` | the event is not delivered |
| `messaging.custom_sms_provider.url` | SMS, so OTP delivery breaks |
| `account_migration.hook.url` | account migration |
| `proof_of_phone_number_verification.hook.url` | phone number verification |
| an SSO provider's `discovery_document_endpoint`, and the `jwks_uri` it names | login with that provider |

Addresses like `http://10.0.0.5:8080/hook`, `http://host.docker.internal:3000/hook` and `http://my-service.svc.cluster.local/hook` are all refused by default after this upgrade.

**Symptom** — one `ERROR` per refused fetch, from logger `ssrf-address-policy`:

```
outbound fetch blocked by SSRF address policy
  blocked_by_ssrf_policy=true
  sink=hook.blocking_handlers
  host=my-service.svc.cluster.local
  flag=http.insecure_fetch_address_allowed
```

`blocked_by_ssrf_policy` appears on no other message, so searching your logs for it after upgrading finds every affected fetch.

**Action** — name the hosts you intend to reach, in `authgear.features.yaml`:

```yaml
http:
  insecure_fetch_address_allowed_hosts:
    - hooks.internal.example.com
    - "*.svc.cluster.local"
    - 10.0.0.5
```

An entry is a hostname or a literal IP, never `host:port`, matched case-insensitively against the host as written in the URL, before DNS resolution. A leading `*.` matches exactly one label: `*.example.com` matches `a.example.com` but not `a.b.example.com` and not `example.com`.

To lift the restriction entirely instead — appropriate for a development or test deployment, not for one serving real users:

```yaml
http:
  insecure_fetch_address_allowed: true
```

Both settings live in `authgear.features.yaml`, which the portal and the Admin API cannot write, so only whoever operates the deployment can set them. Neither is a per-project setting a tenant can turn on.

An allowlisted host is exempt from the rebinding protection as well as the address rules, so list only hosts you control. See [specs/ssrf-protection.md](./specs/ssrf-protection.md).

### Response bodies are capped at 2 MiB

The same fetches now read at most 2 MiB of response body, counted after decompression. A larger response fails the way a malformed one does.

**Affects** a deployment whose webhook, custom SMS provider or OIDC discovery endpoint replies with more than 2 MiB. No legitimate response on these paths is near that.

**Action** — none, unless you have such an endpoint, in which case make it reply with less. The cap is not configurable.

---

## Maintaining this file

Add an entry in the same commit as the breaking change, under `## [Unreleased]`.

On release, rename `## [Unreleased]` to the new tag and open a fresh empty `## [Unreleased]` above it.
