# Portal MCP Server

The portal exposes an [MCP](https://modelcontextprotocol.io/specification/2025-11-25) server so that an AI agent can manage, on behalf of a portal user, the projects that user collaborates on.

## Table of Contents

- [Scope](#scope)
- [Endpoint](#endpoint)
- [Authorization](#authorization)
  - [Protected Resource Metadata](#protected-resource-metadata)
  - [MCP access tokens](#mcp-access-tokens)
  - [Scopes](#scopes)
  - [Registering the MCP resource](#registering-the-mcp-resource)
- [Tools](#tools)
  - [list_projects](#list_projects)
  - [get_project_config](#get_project_config)
  - [update_project_config](#update_project_config)
- [Errors](#errors)
- [Future work](#future-work)

## Scope

The MCP server supports exactly these operations:

| Operation | Tool |
|---|---|
| List the projects the user collaborates on | [`list_projects`](#list_projects) |
| Read a project's `authgear.yaml` | [`get_project_config`](#get_project_config) |
| Replace a project's `authgear.yaml` | [`update_project_config`](#update_project_config) |

Everything else is out of scope, including `authgear.secrets.yaml`, `authgear.features.yaml`, other project resources (templates, translations, images, hooks), the Admin API, collaborators, domains and billing.

## Endpoint

| | |
|---|---|
| URL | `<portal origin>/api/mcp` |
| Transport | [Streamable HTTP](https://modelcontextprotocol.io/specification/2025-11-25/basic/transports#streamable-http), stateless: no `Mcp-Session-Id` is issued, and no server-initiated messages are sent |
| Capabilities | `tools` only |

The URL is the MCP server's **resource URI** (RFC 8707 / RFC 9728). The endpoint is always served; clients can obtain a token for it only once it is [registered](#registering-the-mcp-resource).

## Authorization

The MCP server is an OAuth 2.0 resource server, following the [MCP Authorization specification](https://modelcontextprotocol.io/specification/2025-11-25/basic/authorization). The authorization server is the Authgear project the portal already signs users in with.

### Protected Resource Metadata

Served at `<portal origin>/.well-known/oauth-protected-resource/api/mcp` ([RFC 9728](https://www.rfc-editor.org/rfc/rfc9728)):

```json
{
  "resource": "https://portal.example.com/api/mcp",
  "authorization_servers": ["https://accounts.example.com"],
  "scopes_supported": ["projects:read", "projects:write"],
  "bearer_methods_supported": ["header"]
}
```

A request without a valid access token gets `401` with:

```
WWW-Authenticate: Bearer resource_metadata="https://portal.example.com/.well-known/oauth-protected-resource/api/mcp"
```

### MCP access tokens

An MCP access token is the JWT access token (`typ: at+jwt`) the authorization server issues when the client requests the MCP resource URI as `resource` (see [Access Token Audience Binding](./access-token-audience-binding.md#with-resource-indicator)):

```json
{
  "iss": "https://accounts.example.com",
  "sub": "<user ID>",
  "aud": ["https://portal.example.com/api/mcp"],
  "client_id": "https://mcp-client.example.com/client-metadata.json",
  "scope": "projects:read projects:write",
  "exp": 1700003600
}
```

The MCP server accepts a request only when all of these hold, otherwise it gets `401` as above:

1. The token is a JWT access token signed by the authorization server's keys.
2. `iss` is the authorization server.
3. `aud` contains the MCP resource URI, `<portal origin>/api/mcp` for the origin the request was served on.
4. The token has not expired.

The user is `sub`. Per-project access is then checked per tool: the user must be a collaborator (owner or editor) of the project, the same rule the portal applies.

The portal GraphQL API accepts a bearer access token only when its `aud` contains the authorization server's endpoint, the default audience of a first-party client's JWT access token. It therefore rejects an MCP access token, and a token bound to any other Resource. A portal deployed to authenticate through the [session resolver](./api-resolver.md) instead already rejects every third-party token there; see [Access Token Behavior by Client Kind](./client.md#access-token-behavior-by-client-kind).

### Scopes

| Scope | Grants |
|---|---|
| `projects:read` | `list_projects`, `get_project_config` |
| `projects:write` | `update_project_config` |

`projects:write` does not imply `projects:read`; an agent that edits config needs both, since it must read the checksum first. Calling a tool without its scope fails with `403` and:

```
WWW-Authenticate: Bearer error="insufficient_scope", scope="<required scope>", resource_metadata="..."
```

### Registering the MCP resource

The MCP resource URI is an ordinary [API Resource](./api-resource.md) of the portal's Authgear project; the Authgear server has no knowledge of the MCP server. Until it is registered, the authorization server rejects a request for it with `invalid_target`.

It is registered by a portal command, which is safe to re-run:

```sh
authgear-portal internal mcp setup --app-id <portal's project ID> --portal-origin <portal origin>
```

The command makes the project match this state, creating or updating as needed:

| | |
|---|---|
| Resource URI | `<portal origin>/api/mcp` |
| Scopes | `projects:read`, `projects:write` |
| Access policy | Every client category is allowed, on the Resource and on both Scopes |
| `authgear.yaml` | [CIMD](./cimd.md) enabled |

[DCR](./dcr.md) open registration is not enabled by the command; an operator may enable it as a fallback for clients without CIMD support. The scripts that set up a new portal (Authgear ONCE and local development) run the command.

The user sees a consent screen naming the client and the scopes before any token is issued.

## Tools

A project is identified by its project ID, the same ID shown in the portal URL (`/project/<project_id>`).

### list_projects

Lists the projects the user collaborates on, sorted by project ID ascending.

| | |
|---|---|
| Scope | `projects:read` |
| Annotations | `readOnlyHint: true` |
| Input | none |

Output (`structuredContent`):

```json
{
  "projects": [
    { "project_id": "myapp", "public_origin": "https://myapp.authgear.cloud" }
  ]
}
```

### get_project_config

Returns the project's `authgear.yaml` as stored, not the effective config with defaults filled in.

| | |
|---|---|
| Scope | `projects:read` |
| Annotations | `readOnlyHint: true` |

Input:

| Field | Type | Required |
|---|---|---|
| `project_id` | string | yes |

Output:

```json
{
  "config": "id: myapp\nhttp:\n  public_origin: https://myapp.authgear.cloud\n...",
  "checksum": "8d3c..."
}
```

- `config` — the file as YAML text.
- `checksum` — an opaque string identifying this version of the file; pass it to `update_project_config`.

### update_project_config

Replaces the project's `authgear.yaml` with the given content.

| | |
|---|---|
| Scope | `projects:write` |
| Annotations | `readOnlyHint: false`, `destructiveHint: true`, `idempotentHint: true` |

Input:

| Field | Type | Required | Description |
|---|---|---|---|
| `project_id` | string | yes | |
| `config` | string | yes | The complete new `authgear.yaml`, as YAML text |
| `checksum` | string | yes | The `checksum` from the `get_project_config` result this edit is based on |

The update succeeds only when:

- `checksum` matches the current file, so an edit based on a stale read never overwrites a concurrent change.
- The new config is valid, and permitted by the project's plan (`authgear.features.yaml`) — the same validation the portal applies.
- The new config is satisfiable with the project's current `authgear.secrets.yaml`. A change that needs a new secret (for example, adding an OAuth provider that needs a client secret) is rejected, since secrets cannot be set through MCP.
- `id` is unchanged.

On success it has the same effects as saving `authgear.yaml` in the portal, including the `project.app.updated` audit log entry, and returns the new state:

```json
{ "checksum": "a91f..." }
```

## Errors

Tool failures are returned as a tool result with `isError: true` (not a JSON-RPC error), so the agent can read and act on them. The text content states what went wrong and what to do next.

| Situation | Message tells the agent to |
|---|---|
| Project not found, or user is not a collaborator | Call `list_projects`; the two cases are indistinguishable |
| `checksum` does not match | Call `get_project_config` again and reapply the edit |
| Config is not valid YAML | Fix the syntax; includes the parser's location |
| Config fails validation | Fix each listed violation; each one names the JSON pointer and the reason |
| Config needs a secret that is not set | Set the secret in the portal first |
| `id` changed | Keep `id` as it was |

Authentication and scope failures are HTTP errors, as described under [Authorization](#authorization).

## Future work

- Partial updates (JSON Patch or merge patch) instead of whole-file replacement.
- Reading and updating other project resources, and `authgear.secrets.yaml`.
