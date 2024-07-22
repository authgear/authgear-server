# Demonstrating Proof of Possession (DPoP)

Authgear supports DPoP which was specified in [rfc9449](https://datatracker.ietf.org/doc/html/rfc9449).

The following tokens will be bound to a DPoP private key if the `DPoP` header was provided in the request which issued the token:

- Refresh Token
- Device Secret

The following tokens will be bound to a DPoP private key if the `dpop_jkt` query parameter was provided in the authorization request which issued th token:

- Authorization Code
