- [LDAP](#ldap)
  * [Supported LDAP protocol](#supported-ldap-protocol)
  * [Authenticate Authgear as a LDAP client](#authenticate-authgear-as-a-ldap-client)
  * [Configuration of LDAP servers](#configuration-of-ldap-servers)
  * [Validation on the configuration](#validation-on-the-configuration)
  * [Testing on the configuration](#testing-on-the-configuration)
  * [The database schema of a LDAP identity](#the-database-schema-of-a-ldap-identity)
  * [Handling of a LDAP identity](#handling-of-a-ldap-identity)
  * [The UX of LDAP in Auth UI](#the-ux-of-ldap-in-auth-ui)
  * [Errors](#errors)

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

Authgear supports Simple Bind with DN and password, according to https://datatracker.ietf.org/doc/html/rfc4513#section-5.1.3

## Configuration of LDAP servers

The following example demonstrates how the developer should configure Authgear to connect to their LDAP servers.

In `authgear.yaml`

```yaml
identity:
  ldap:
    servers:
    - name: default
      url: "ldap://localhost:389"
      base_dn: "dc=localhost"
      search_filter_template: |
        {{- if (hasSuffix $.Username "@mycompany.com") }}
          (&(objectCategory=person)(objectClass=user)(memberof=dc=mycompany,dc=com)(sAMAccountName={{ $.Username }}))
        {{- else }}
          (&(objectCategory=person)(objectClass=user)(memberof=dc=anothercompany,dc=com)(sAMAccountName={{ $.Username }}))
        {{- end }}
      # According to https://datatracker.ietf.org/doc/html/rfc4512#section-2.5
      # The unique way to identify an attribute in LDAP is oid.
      # For those who are curious, "0.9.2342.19200300.100.1.1" is "uid", according to https://datatracker.ietf.org/doc/html/rfc4519#section-2.39
      user_id_attribute_oid: "0.9.2342.19200300.100.1.1"
```

- `identity.ldap.servers.name`: A unique name to identify this LDAP server. Once set, it cannot be changed. It is stored in the database as part of the unique key to identify a LDAP identity. See [The database schema of a LDAP identity](#the-database-schema-of-a-ldap-identity) for details.
- `identity.ldap.servers.url`: The connection URL to the LDAP server. The scheme MUST be `ldap:` or `ldaps:`. The URL MUST contain `host`, and optionally a port. If the port is omitted, the default port of the scheme is assumed. The default port of `ldap:` is `389`, while the default port of `ldaps:` is `636`. The URL MUST NOT contain other elements, such as path, nor query.
- `identity.ldap.servers.base_dn`: The base DN to construct a Search Request, as defined in [Section 4.5.1 in RFC4511](https://datatracker.ietf.org/doc/html/rfc4511#section-4.5.1).
- `identity.ldap.servers.search_filter_template`: A Go template that renders to a filter to be used in the Search Request. This template can use the variable `$.Username` to render the username entered by the end-user. `$.Username` is pre-processed so that it is an escaped LDAP string. The strings function from [https://masterminds.github.io/sprig/](https://masterminds.github.io/sprig/) can be used in the template.
- `identity.ldap.servers.user_id_attribute_oid`: The attribute that is guaranteed to be unique and never change for a given user in the LDAP server. It is used to identify a user from the LDAP server. Warning: Changing this value will cause Authgear not able to look up any previous LDAP identities.

> What if I want to use a different base_dn depend on the username?
> In this case, you need to specify a very generic base_dn like "dc=com", and then
> you write your own search_filter_template to filter entries.

> Why does `identity.ldap.servers.url` allow scheme, host, and port?
> The LDAP URL, defined in [Section 2 in RFC 4516](https://datatracker.ietf.org/doc/html/rfc4516#section-2), is syntactically different from the URL defined in [RFC3986](https://datatracker.ietf.org/doc/html/rfc3986).
> In particular, a LDAP URL can contain multiple question mark characters.
> To ease implementation, we do not support the LDAP URL, and require a RFC3986 URL (which is implemented by the standard library net/url package).

> TODO: The current configuration is missing an important feature. The feature is allow the developer to specify what attributes they want to retrieve from the LDAP server, and
> what attributes map to which standard attributes.

In `authgear.secrets.yaml`

```yaml
secrets:
- data:
    items:
    - name: default
      # According to https://datatracker.ietf.org/doc/html/rfc4513#section-5.1.3,
      # Simple Bind takes a DN and a password.
      dn: cn=authgear,dc=example,dc=com
      password: secret1
  key: ldap
```

- `items.name`: To associate a LDAP server in `authgear.yaml`.
- `items.dn`: Optional. The DN of the LDAP entry Authgear uses to authenticate itself to the LDAP server. If it is not provided, then Authgear does not authenticates itself, and assumes the LDAP server allows anonymous requests.
- `items.password`: Optional. The password of the LDAP entry Authgear uses to authenticate itself to the LDAP server. If `dn` is provided, then `password` is required.

## Validation on the configuration

Here is the JSON schema for the LDAP server configuration.

```
{
  "type": "object",
  "additionalProperties": false,
  "required": ["name", "url", "base_dn", "search_filter_template", "user_id_attribute_oid"],
  "properties": {
    "name": {
      "type": "string",
      "minLength": 1
    },
    "url": {
      "type": "string",
      "format": "ldap_url"
    },
    "base_dn": {
      "type": "string",
      "format": "ldap_dn"
    },
    "search_filter_template": {
      "type": "string",
      "format": "ldap_search_filter_template"
    },
    "user_id_attribute_oid": {
      "type": "string",
      "format": "ldap_oid"
    }
  }
}
```

- `format: ldap_url`: It is a JSON schema format that implements the rules of `identity.ldap.servers.url`.
- `format: ldap_dn`: It is a JSON schema format that validates the value to be a valid DN.
- `format: ldap_search_filter_template`: It is a JSON schema format that validates the rendered string to be a valid Search Filter. It does the validation by running the template with `Username=user`, `Username=user@example.com`, and `Username=+85298765432`, and then parse the resulting Search Filter as a Search Filter.
- `format: ldap_oid`: It is a JSON schema format that validates the value to be a valid LDAP oid.

## Testing on the configuration

> TODO: Add a mutation in the Admin API to test LDAP connection.
> It should take the whole server configuration, and optionally a end-user username.
> It connects the LDAP server with the URL and the credentials.
> If the optional end-user username is given, it performs a Search request, and validates the user exists and has user_id_attribute_oid.
> It returns detailed API errors so that the portal can display relevant information for the developer to debug the configuration.
> Such detailed API errors ARE NOT returned in actual use. They are reported as internal errors.

## The database schema of a LDAP identity

```sql
CREATE TABLE _auth_identity_ldap
(
    id                      text  PRIMARY KEY REFERENCES _auth_identity (id),
    app_id                  text  NOT NULL,
    server_name             text  NOT NULL,
    user_id_attribute_oid   text  NOT NULL,
    user_id_attribute_value text  NOT NULL,
    claims                  jsonb NOT NULL,
    raw_entry_json          jsonb NOT NULL
);

CREATE UNIQUE INDEX _auth_identity_ldap_unique ON _auth_identity_ldap (app_id, server_name, user_id_attribute_oid, user_id_attribute_value);
```

- `id`: The primary key of this table. This is the same as other `_auth_identity_*` tables.
- `app_id`: The app ID of this table for multi-tenant. This is the same as other `_auth_identity_*` tables.
- `server_name`: The `name` of the LDAP server.
- `user_id_attribute_oid`: The `user_id_attribute_oid` of the LDAP server when this identity is created.
- `user_id_attribute_value`: The value of the `user_id_attribute_oid` of the user.
- `claims`: The standard claims extracted from this LDAP entry.
- `raw_entry_json`: The raw LDAP entry encoded in JSON. It looks like `{ "dn": "uid=johndoe,dc=example,dc=com", "attr1": ["value1"] }`.

## Handling of a LDAP identity

- To look up a LDAP identity in Authgear, we use the tuple `(app_id, server_name, user_id_attribute_oid, user_id_attribute_value)`.
- Similar to OAuth identity, we update an LDAP identity when it is used in login.

## The UX of LDAP in Auth UI

In the MVP phase (that is, now), sign in with LDAP is like sign in with an OAuth provider.
Except that the enter-username-and-password page is hosted by Authgear as a integral part of Auth UI.

In the future, we may consider an option to make LDAP "replaces" Login ID in the UX.

## Errors

This section documents the expected errors.

|Description|Name|Reason|Info|
|---|---|---|---|
|When the LDAP server is service unavailable, `user_id_attribute_oid` not found in a user, search filter turns out to be invalid, etc|InternalError|UnexpectedError||
|When the end-user cannot authenticate to the LDAP server|Unauthorized|InvalidCredentials||
