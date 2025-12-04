# Network Protection

This document specifies project-level network protection configurations.

- [IP Blocklist](#ip-blocklist)
- [authgear.yaml](#authgearyaml)
- [Error response](#error-response)

## IP Blocklist

### authgear.yaml

```yaml
network_protection:
  ip_blocklist:
    cidrs:
      - 2.2.2.2/32
      - 2001:db8::/32
    country_codes:
      - HK
      - GB
```

- `network_protection.ip_blocklist` (object): Top-level configuration for IP-based blocking. When configured, requests that match any entry in the blocklist will be denied.

- `network_protection.ip_blocklist.cidrs` (array of strings): A list of IPv4 or IPv6 addresses or CIDR ranges to block. Each entry must be in CIDR notation (e.g. `2.2.2.2/32`, `2001:db8::/32`).

- `network_protection.ip_blocklist.country_codes` (array of strings): A list of ISO 3166-1 alpha-2 country codes (uppercase, e.g. `US`, `GB`, `HK`) whose traffic should be blocked.

### Error response

The server returns HTTP status code 403 with the plain-text response body:

"Your IP is not allowed to access this resource"
