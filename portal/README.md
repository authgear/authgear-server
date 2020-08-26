# Portal

# Setup locally

We need to setup the environment variable `AUTHGEAR_CLIENT_ID=portal` and `AUTHGEAR_ENDPOINT=http://localhost:3000`.

We need the following `authgear.yaml` to setup authgear for the portal.

```yaml
http:
  allowed_origins:
  # The SDK uses XHR to fetch the OAuth/OIDC configuration,
  # So we have to allow the origin of the portal.
  - localhost:8000
oauth:
  clients:
  # Create a client for the portal.
  # Since we assume the cookie is shared, there is no grant nor response.
  - client_id: portal
    # Note that the trailing slash is very important here
    # URIs are compared byte by byte.
    redirect_uris:
    - "http://localhost:8000/"
    post_logout_redirect_uris:
    - "http://localhost:8000/"
    grant_types: []
    response_types: ["none"]
```

# Two graphql schemas

We have two graphql schemas.
We take advantage of [Babel 7 File-relative configuration](https://babeljs.io/docs/en/config-files#file-relative-configuration) to configure `babel-plugin-relay` differently.
In this setup `relay-config` is useless to us.
