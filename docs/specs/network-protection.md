# Network Protection

This document specifies project-level network protection configurations.

- [IP Blocklist](#ip-blocklist)
  - [authgear.yaml](#authgearyaml)
  - [Error response](#error-response)
  - [Use cases](#use-cases)

## IP Blocklist

### authgear.yaml

```yaml
network_protection:
  ip_blocklist:
    cidrs:
      - 2.2.2.2/16
      - 2001:db8::/32
    country_codes:
      - GB
    exceptions:
      cidrs:
        - 2.2.3.4/32
      country_codes:
        - HK
```

- `network_protection.ip_blocklist` (object): Top-level configuration for IP-based blocking. When configured, requests that match any entry in the blocklist will be denied.

- `network_protection.ip_blocklist.cidrs` (array of strings): A list of IPv4 or IPv6 addresses or CIDR ranges to block. Each entry must be in CIDR notation (e.g. `2.2.2.2/32`, `2001:db8::/32`).

- `network_protection.ip_blocklist.country_codes` (array of strings): A list of ISO 3166-1 alpha-2 country codes (uppercase, e.g. `US`, `GB`, `HK`) whose traffic should be blocked.
 
- `network_protection.ip_blocklist.exceptions` (object, optional): A set of exemptions from the blocklist. If a request matches an exception entry, it is allowed even if it also matches a blocklist entry. Use exceptions to whitelist specific IPs/ranges or entire countries that should be exempted from blocking.

- `network_protection.ip_blocklist.exceptions.cidrs` (array of strings): CIDR ranges or single IP addresses (in CIDR notation) that are exempted from blocking (e.g. `2.2.3.4/32`).

- `network_protection.ip_blocklist.exceptions.country_codes` (array of strings): ISO 3166-1 alpha-2 country codes (uppercase) whose traffic is exempted from blocking (e.g. `HK`).

Matching semantics:

1. If the client's IP matches any entry in `ip_blocklist.exceptions.cidrs` or the client's country matches any entry in `ip_blocklist.exceptions.country_codes`, the request is allowed.
2. Otherwise, if the client's IP or country matches any entry in `ip_blocklist.cidrs` or `ip_blocklist.country_codes`, the request is denied.
3. If neither exceptions nor blocklist entries match, the request is allowed.

### Error response

The server returns HTTP status code 403 with the plain-text response body:

"Your IP is not allowed to access this resource"

### Use cases

1. Block only a specific IP address

```yaml
network_protection:
  ip_blocklist:
    cidrs:
      - 203.0.113.5/32
```

This configuration blocks requests coming from the single IPv4 address `203.0.113.5` while allowing other traffic.

2. Block only a specific country

```yaml
network_protection:
  ip_blocklist:
    country_codes:
      - CN
```

This configuration blocks traffic originating from China (`CN`) only.

3. Block all requests, except a specific IP range

```yaml
network_protection:
  ip_blocklist:
    cidrs:
      - 0.0.0.0/0
      - ::/0
    exceptions:
      cidrs:
        - 198.51.100.0/24
```

This blocks all IPv4 and IPv6 addresses except the IPv4 range `198.51.100.0/24`, which is allowed via `exceptions.cidrs`.

4. Block all requests, except a specific country

```yaml
network_protection:
  ip_blocklist:
    cidrs:
      - 0.0.0.0/0
      - ::/0
    exceptions:
      country_codes:
        - US
```

This blocks all traffic except requests from the United States (`US`), which are exempted.
