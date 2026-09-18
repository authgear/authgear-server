# Session resolver API

Popular reverse proxy server supports delegating request authentication by initiating sub-request.

In nginx, it is the `auth_request` directive while in Traefik, it is `ForwardAuth`.

The resolve endpoint `/resolve` looks at `Cookie:` and `Authentication:` to authenticate the request. `Cookie:` has higher precedence.

A bearer access token presented this way is accepted only when it belongs to a **first-party client**. **Any** access token belonging to a **third-party client** (static `third_party_app`, DCR-registered or CIMD-resolved) is rejected — `x-authgear-session-valid` below is `false` — whether the token is opaque or a resource-bound JWT.

This endpoint is meant for a client deployed on the same site as Authgear, behind the same reverse proxy. A first-party client can be deployed that way whether it was declared in `authgear.yaml` or registered dynamically, so how it was registered does not matter here. A third-party client is by definition operated by someone the project does not trust and is not same-site, and this endpoint has no notion of "resource" and never exposes the token's `aud` in its response, so accepting one would invite audience confusion (see [access-token-audience-binding.md](./access-token-audience-binding.md)).

Resource binding does not affect the decision either way: a first-party client's token is accepted whether or not it names a resource. See [client.md — Access Token Behavior by Client Kind](./client.md#access-token-behavior-by-client-kind) for the full breakdown by client kind.

The resolve endpoint does not write body. Instead, it adds the following headers in the response.

- [x-authgear-session-valid](#x-authgear-session-valid)
- [x-authgear-user-id](#x-authgear-user-id)
- [x-authgear-user-anonymous](#x-authgear-user-anonymous)
- [x-authgear-user-verified](#x-authgear-user-verified)
- [x-authgear-user-roles](#x-authgear-user-roles)
- [x-authgear-session-amr](#x-authgear-session-amr)
- [x-authgear-session-authenticated-at](#x-authgear-session-authenticated-at)
- [x-authgear-user-can-reauthenticate](#x-authgear-user-can-reauthenticate)

## x-authgear-session-valid

Tell whether the session of the original request is valid.

If this header is absent, it means the original request is not associated with any session.

If the value is `true`, it indicates the original request has a valid session. More headers will be included.

If the value is `false`, it indicates the original request has invalid session.

## x-authgear-user-id

The user id.

## x-authgear-user-anonymous

The value `true` means the user is anonymous. Otherwise, it is a normal user.

## x-authgear-user-verified

The value `true` means the user is verified.

## x-authgear-user-roles

A comma-separated list of the effective roles of the user.

The order is unspecified.

If the user does not have any roles, this header is absent.

For example, `x-authgear-user-roles: stock.view,stock.edit`

## x-authgear-session-amr

See [the amr claim](./oidc.md#amr). It is comma-separated.

## x-authgear-session-authenticated-at

See [the auth_time claim](./oidc.md#auth_time). It is an integer.

## x-authgear-user-can-reauthenticate

The value `true` means the user can possibly reauthenticate.
