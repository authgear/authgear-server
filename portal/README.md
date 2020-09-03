# Portal

# Prerequisite

Follow [authgear setup guide](../README.md) to start Authgear

# Setup locally

## Setup proxy

We need a proxy server to delegate the authentication, which is explained [here](https://docs.authgear.com/getting-started/auth-nginx)

To start proxy container

```
# in /portal
docker-compose up -d
```

This container listen to port 8000. For redirection configuration, please refer to nginx config `/portal/nginx.conf`

## Setup environment variable

We need to setup the environment variable `AUTHGEAR_CLIENT_ID=portal` and `AUTHGEAR_ENDPOINT=http://localhost:3000`, and start graphQL server.

## Setup graphQL server

```sh
go run ./cmd/portal start
```

## Setup portal development server

1. Install dependencies

```
npm install
```

2. Run development server

```
npm run start
```

This command should start a web development server on port 1234.

3. Configure authgear.yaml

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

## Sign-up

When you try to access the portal through proxy `http://localhost:8000` (port 8000), you will be redirected to authgear sign up / sign in page in port 3000 (hosted by authgear main server)

NOTE: make sure the authgear server is in development mode, which skips sending actual email

Follow the instruction on the webpage,

1. Click `Create One` link
2. Input an email address
3. The website will ask for verification code, to get verified, inspect the log from authgear main server

   - Find the line with `WARN`, you will find a link for verifying the account in the log
   - For example: `http://localhost:3000/verify_identity?code=SJDN080N&state=W3X8GNP6N1ZBCQKD0J582MAS369FYKC6`
   - Using this link, the verification code will be automatically filled in, then we can proceed to creating password

4. Enter new password to complete the signup flow.

## Visit authgear portal page

Make sure you go to authgear portal page, use `http://localhost:8000` (access through proxy), the webpage needs to call graphQL server with the same domain and port. The api call fails if we access through port 1234 directly.

# Two graphql schemas

We have two graphql schemas.
We take advantage of [Babel 7 File-relative configuration](https://babeljs.io/docs/en/config-files#file-relative-configuration) to configure `babel-plugin-relay` differently.
In this setup `relay-config` is useless to us.
