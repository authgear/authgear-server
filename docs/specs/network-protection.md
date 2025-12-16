# Network Protection

This document specifies project-level network protection configurations.

- [IP Blocklist](#ip-blocklist)
  - [authgear.yaml](#authgearyaml)
  - [Error response](#error-response)
  - [Use cases](#use-cases)

## IP Filter

### authgear.yaml

```yaml
network_protection:
  ip_filter:
    default_action: deny  # Default is "allow" when omitted
    rules: # Match from top to bottom, the first match takes effect
      # Allow office subnet in US
      - name: allow-office-us
        action: allow
        source:
          cidrs: ["192.168.1.0/24"]
          geo_location_codes: ["US"] # Named to match event context

      # Block known abusive IP range
      - name: block-abuse-range
        action: deny
        source:
          cidrs:
            - "203.0.113.0/24"
            - "198.51.100.0/24"

      # Block traffic from specific countries
      - name: block-countries
        action: deny
        source:
          geo_location_codes: ["KP", "IR", "SY", "VE", "CU"]

      # Optional catch-all deny rule (redundant with default_action)
      - name: deny-all
        action: deny
        source:
          cidrs: ["0.0.0.0/0"]
```

`network_protection.ip_filter` (object): Top-level configuration for rule-based IP filtering. Rules are evaluated from top to bottom; the first matching rule's action is applied. If no rule matches, `default_action` is used.

- `network_protection.ip_filter.default_action` (string): Global fallback action when no rule matches. Allowed values: `allow`, `deny`. Default is `allow` if omitted.

- `network_protection.ip_filter.rules` (array of objects): Ordered list of matching rules. The first rule that matches a request is applied.

- Rule object:
  - `name` (string, optional): Human-readable identifier for the rule.
  - `action` (string): Action to take when the rule matches. Allowed values: `allow`, `deny`.
  - `source` (object): Criteria used to match requests. At least one of the subfields should be present.
    - `cidrs` (array of strings, optional): IPv4/IPv6 addresses or CIDR ranges to match (e.g. `198.51.100.0/24`, `2001:db8::/32`).
    - `geo_location_codes` (array of strings, optional): ISO 3166-1 alpha-2 country codes (uppercase, e.g. `US`, `GB`).

Matching semantics:

1. For each incoming request, evaluate rules in `network_protection.ip_filter.rules` from top to bottom.
2. A rule matches if the request's source IP matches any CIDR in `source.cidrs` (if present) or the request's geo country matches any entry in `source.geo_location_codes` (if present). If both are present, the rule matches when either dimension matches.
3. When the first matching rule is found, apply that rule's `action` (`allow` or `deny`) and stop evaluating further rules.
4. If no rule matches, apply `network_protection.ip_filter.default_action`.

### Error response

The server returns HTTP status code 403 with the plain-text response body:

"Your IP is not allowed to access this resource"

### Use cases

1. Block only a specific IP address

```yaml
network_protection:
  ip_filter:
    rules:
      - name: block-single-ip
        action: deny
        source:
          cidrs: ["203.0.113.5/32"]
```

This configuration denies requests from the single IPv4 address `203.0.113.5` while allowing other traffic.

2. Block only a specific country

```yaml
network_protection:
  ip_filter:
    default_action: allow
    rules:
      - name: block-china
        action: deny
        source:
          geo_location_codes: ["CN"]
```

This configuration denies traffic originating from China (`CN`) only.

3. Block all requests, except a specific IP range

```yaml
network_protection:
  ip_filter:
    default_action: deny
    rules:
      - name: allow-exception-range
        action: allow
        source:
          cidrs: ["198.51.100.0/24"]
```

This configuration denies all traffic except the IPv4 range `198.51.100.0/24`, which is allowed by the earlier rule.

4. Block all requests, except a specific country

```yaml
network_protection:
  ip_filter:
    default_action: deny
    rules:
      - name: allow-us
        action: allow
        source:
          geo_location_codes: ["US"]
```

This denies all traffic except requests from the United States (`US`), which are allowed by the first rule.

5. Block requests from data center IPs, except the one we owned

```yaml
network_protection:
  ip_filter:
    rules:
      - name: allow-my-node
        action: allow
        source:
          cidrs: ["8.8.4.33/32"]
      # See https://www.gstatic.com/ipranges/goog.json
      - name: block-datacenter
        action: block
        source:
          cidrs: ["8.8.4.0/24"]
```

This denies requests from Google, except `8.8.4.33` which allowed by the first rule.
