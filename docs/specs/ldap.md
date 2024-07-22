- [LDAP](#ldap)
  * [Supported LDAP protocol](#supported-ldap-protocol)
  * [Authenticate Authgear as a LDAP client](#authenticate-authgear-as-a-ldap-client)
  * [Configuration of LDAP servers](#configuration-of-ldap-servers)
  * [The database schema of a LDAP identity](#the-database-schema-of-a-ldap-identity)
  * [Handling of a LDAP identity](#handling-of-a-ldap-identity)
  * [The UX of LDAP in Auth UI](#the-ux-of-ldap-in-auth-ui)

# LDAP

This document describes the LDAP support in Authgear. Authgear acts as a LDAP client, and connects to one or more LDAP servers.

## Supported LDAP protocol

The supported LDAP protocol is LDAPv3, which is specified in [RFC4511](https://datatracker.ietf.org/doc/html/rfc4511).

Authgear MUST BE compatible with

1. A LDAP server without TLS (i.e. ldap://)
2. The non-standard ldaps://
3. The StartTLS operation, as defined in [Section 4.14 in RFC4511](https://datatracker.ietf.org/doc/html/rfc4511#section-4.14)

When the connection URL starts with `ldap://`, StartTLS will be tried.
If the LDAP server responds protocolError, as defined in [Section 4.14.1 in RFC4511](https://datatracker.ietf.org/doc/html/rfc4511#section-4.14.1),
then Authgear will treat it as no TLS is required.

When the connection URL starts with `ldaps://`, then Authgear will connect to the LDAP server with TLS directly, without using StartTLS.

## Authenticate Authgear as a LDAP client

It is very common that before a LDAP client can run any [Search Operation](https://datatracker.ietf.org/doc/html/rfc4511#section-4.5),
the LDAP client must perform [Bind Operation](https://datatracker.ietf.org/doc/html/rfc4511#section-4.2) first.

Authgear supports Simple Bind with username and password.

## Configuration of LDAP servers

The following example demonstrates how the developer should configure Authgear to connect to their LDAP servers.

In `authgear.yaml`

```yaml
identity:
  ldap:
    servers:
    - name: ldap1
      url: "ldap://localhost:389"
      base_distinguished_name: "dc=localhost"
      relative_distinguished_name_attribute: "uid"
    - name: ldap2
      url: "ldap://mycompany.com:389"
      base_distinguished_name: "dc=mycompany,dc=com"
      relative_distinguished_name_attribute: "uid"
```

- `identity.ldap.servers.name`: A name that only exists in `authgear.yaml` and `authgear.secrets.yaml` for associating a LDAP server. It serves no other purpose.
- `identity.ldap.servers.url`: The connection URL to the LDAP server. The scheme MUST be `ldap:` or `ldaps:`. The URL MUST contain `host`, and optionally a port. If the port is omitted, the default port of the scheme is assumed. The default port of `ldap:` is `389`, while the default port of `ldaps:` is `636`. The URL MUST NOT contain other elements, such as path, nor query.

> Why does `identity.ldap.servers.url` allow scheme, host, and port?
> The LDAP URL, defined in [Section 2 in RFC 4516](https://datatracker.ietf.org/doc/html/rfc4516#section-2), is syntactically different from the URL defined in [RFC3986](https://datatracker.ietf.org/doc/html/rfc3986).
> In particular, a LDAP URL can contain multiple question mark characters.
> To ease implementation, we do not support the LDAP URL, and require a RFC3986 URL (which is implemented by the standard library net/url package).

- `identity.ldap.servers.base_distinguished_name`: The base distinguished name to construct a Search Request, as defined in [Section 4.5.1 in RFC4511](https://datatracker.ietf.org/doc/html/rfc4511#section-4.5.1).
- `identity.ldap.servers.relative_distinguished_name_attribute`: The attribute name Authgear should use to construct the search request. For example, if the value is `uid`, and the end-user gives a username of `user1`, and the base distinguished name is `dc=example,dc=com`, then the relative distinguished name is `uid=user1`, and the distinguished name is `uid=user1,dc=example,dc=com`.

> base_distinguished_name and relative_distinguished_name_attribute may not be sufficient if the developer needs to determine the DN in a more dynamic way.
> In the future, we can support a new configuration, distinguished_name_template, which is a Go template that MUST return a DN.
> It looks like
> ```
>   {{- if (hasSuffix $.Username "@mycompany.com") }}
>     {{- (ldapDN (ldapAttribute "uid" $.Username) (ldapParse "dc=mycompany,dc=com") ) }}
>   {{- else }}
>     {{- (ldapDN (ldapAttribute "uid" $.Username) (ldapParse "dc=anothercompany,dc=com") ) }}
>   {{- end }}
> ```

> TODO: The current configuration is missing an important feature. The feature is allow the developer to specify what attributes they want to retrieve from the LDAP server, and
> what attributes map to which standard attributes.

In `authgear.secrets.yaml`

```yaml
secrets:
- data:
    items:
    - name: ldap1
      username: authgear
      password: secret1
    - name: ldap2
      username: authgear
      password: secret2
  key: ldap
```

- `items.name`: To associate a LDAP server in `authgear.yaml`.
- `items.username`: Optional. The username Authgear uses to authenticate itself to the LDAP server. If it is not provided, then Authgear does not authenticates itself, and assumes the LDAP server allows anonymous requests.
- `items.password`: Optional. The password Authgear uses to authenticate itself to the LDAP server. If `username` is provided, then `password` is required.

## The database schema of a LDAP identity

```sql
CREATE TABLE _auth_identity_ldap
(
    id                 text  PRIMARY KEY REFERENCES _auth_identity (id),
    app_id             text  NOT NULL,
    server_url         text  NOT NULL,
    distinguished_name text  NOT NULL,
    claims             jsonb NOT NULL,
    raw_entry_json     jsonb NOT NULL
);

CREATE UNIQUE INDEX _auth_identity_ldap_unique ON _auth_identity_ldap (app_id, server_url, distinguished_name);
```

- `id`: The primary key of this table. This is the same as other `_auth_identity_*` tables.
- `app_id`: The app ID of this table for multi-tenant. This is the same as other `_auth_identity_*` tables.
- `server_url`: The URL to the LDAP server when this identity was created. The value is taken from the configuration at that moment. It does not change even if the URL in `authgear.yaml` changes.
- `distinguished_name`: The distinguished name of this LDAP entry.
- `claims`: The standard claims extracted from this LDAP entry.
- `raw_entry_json`: The raw LDAP entry encoded in JSON. It looks like `{ "dn": "uid=johndoe,dc=example,dc=com", "attr1": ["value1"] }`.

## Handling of a LDAP identity

- To look up a LDAP identity in Authgear, we use the tuple `(app_id, server_url, distinguished_name)`.
- Similar to OAuth identity, we update an LDAP identity when it is used in login.

## The UX of LDAP in Auth UI

In the MVP phase (that is, now), sign in with LDAP is like sign in with an OAuth provider.
Except that the enter-username-and-password page is hosted by Authgear as a integral part of Auth UI.

In the future, we may consider an option to make LDAP "replaces" Login ID in the UX.
