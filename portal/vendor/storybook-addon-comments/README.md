# @oursky/storybook-addon-comments

**Private Oursky addon.** Storybook 10 **Comments** panel (per-story threads, resolve/delete, sidebar counts).

**Default path at Oursky:** ship Storybook on [Skyrocket](https://github.com/oursky/skyrocket) (`https://<project>.oursky.dev`) with this panel plus a comments API. Local `npm run storybook` is only for UI work; comments need the Skyrocket API (or a local proxy to one).

This package is **only the panel**. It does not store comments. Storage is a small HTTP API and Postgres on Skyrocket (see below).

## Install for Skyrocket (required)

Skyrocket builds your Docker image on the server. That builder runs `npm ci` **with no GitHub login**. A `github:` / `git+ssh` dependency on this **private** repo fails with **git exit 128**.

**Copy this repo into your Storybook project** and depend on the copy:

```bash
gh repo clone oursky/storybook-addon-comments vendor/storybook-addon-comments
# commit vendor/ — it must be in the Skyrocket tarball
```

`package.json`:

```json
{
  "devDependencies": {
    "@oursky/storybook-addon-comments": "file:vendor/storybook-addon-comments"
  }
}
```

Storybook Dockerfile: copy the vendor folder **before** `npm ci`:

```dockerfile
COPY package.json package-lock.json ./
COPY vendor/storybook-addon-comments ./vendor/storybook-addon-comments
RUN npm ci
COPY . .
```

When you upgrade the addon, refresh `vendor/`, commit, then deploy (Skyrocket only rolls pods when `git-<HEAD>` changes).

Laptop-only (not enough for Skyrocket):

```bash
npm install git+ssh://git@github.com/oursky/storybook-addon-comments.git
```

You still need org access to this repo.

## Use the panel

`.storybook/main.ts`:

```ts
const config = {
  addons: ["@oursky/storybook-addon-comments"],
};
export default config;
```

Optional API prefix in `.storybook/manager.ts` (default `/api`):

```ts
import { addons } from "storybook/manager-api";

addons.setConfig({
  commentsApiPath: "/api",
});
```

The panel calls:

| Method | Path |
| --- | --- |
| `GET` | `{commentsApiPath}/comments?storyId=` |
| `GET` | `{commentsApiPath}/comments/counts` → `{ counts: { [storyId]: unresolvedCount } }` |
| `POST` | `{commentsApiPath}/comments` body `{ storyId, author, body }` |
| `POST` | `{commentsApiPath}/comments/:id/resolve` body `{ resolved }` |
| `POST` | `{commentsApiPath}/comments/:id/delete` |

Unresolved counts also appear as teal badges on **sidebar stories, group headers, and root nav items** (sum of descendant stories). Resolved comments are not counted. There is no per-user “read” state yet.

Local `npm run storybook` has no API unless you proxy one. The panel will error until Storybook is served in front of that API (Skyrocket nginx `/api/` is the usual setup).

## Deploy comments on Skyrocket

Match the official [Skyrocket two-service example](https://github.com/oursky/skyrocket/blob/main/example/skyrocket.yaml): **default** is nginx serving static Storybook; **api** is Node on `8080`; Postgres is injected.

### 1. Project layout

Run `skyrocket deploy` from the directory that contains `skyrocket.yaml` (not a parent repo root unless that is the Skyrocket project).

```yaml
name: your-project-storybook
services:
  - name: default
    port: 80
    public: true
    dockerfile: Dockerfile
  - name: api
    port: 8080
    public: true
    dockerfile: comments-api/Dockerfile
    context: comments-api
resources:
  postgresql:
    enable: true
```

`dockerfile` paths are relative to the **tarball root** (the Skyrocket project directory), not to `context`. If `context` is `comments-api`, set `dockerfile: comments-api/Dockerfile`, not `dockerfile: Dockerfile`. Using the Storybook Dockerfile with the API folder as context will fail (`COPY` paths missing).

### 2. nginx (`docker/default.conf.template`)

Copy [example/default.conf.template](https://github.com/oursky/skyrocket/blob/main/example/default.conf.template). Proxy `/api/` to the sibling service. **Do not use `$host` or `$uri`** — the nginx image runs `envsubst` and will blank them.

```nginx
server {
  listen 80;

  location /api/ {
    proxy_pass ${SKYROCKET_API_URL}/;
  }

  location / {
    root /usr/share/nginx/html;
    index index.html;
  }
}
```

Set `api.public: true` like the example. In-cluster the Storybook container still reaches `api:8080` via `SKYROCKET_API_URL`.

### 3. Comments API

Write a small Node (or any) server on `PORT` (default `8080`) that implements the routes in the table above and uses `SKYROCKET_POSTGRES_URL`.

JSON the panel expects:

- `GET /comments?storyId=` → `{ comments: [{ id, storyId, author, body, createdAt, resolved }] }`
- `GET /comments/counts` → `{ counts: { "<storyId>": 2 } }` (unresolved only)
- `POST /comments` → `201` `{ comment: { ...same fields } }`
- `POST /comments/:id/resolve` → `{ comment: { ...same fields } }`
- `POST /comments/:id/delete` → `{ ok: true }`
- Unknown route → `{ error: "not found" }`; missing row → `{ error: "comment not found" }`

A matching Postgres table:

```sql
CREATE TABLE IF NOT EXISTS story_comments (
  id TEXT PRIMARY KEY,
  story_id TEXT NOT NULL,
  author TEXT NOT NULL,
  body TEXT NOT NULL,
  resolved BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

Skyrocket’s URL often has `sslmode=require`. **`pg` v8 treats `require` as verify-full**; Skyrocket’s cert is self-signed. Rewrite to `sslmode=no-verify` in the API.

Listen on `0.0.0.0` **before** waiting on schema, or the rollout health check times out. Serve `GET /` and `GET /healthz` immediately (JSON `{ ok, postgres, schemaReady }` is enough).

### 4. Storybook image

Build static Storybook in Docker and copy `storybook-static` into nginx (same shape as the [Skyrocket example Dockerfile](https://github.com/oursky/skyrocket/blob/main/example/Dockerfile)). Use the **vendored `file:` addon** from the install section. Do not `npm install github:oursky/storybook-addon-comments` in that image.

### 5. Deploy

Install the Skyrocket CLI (repo is private; `gh` must see `oursky/skyrocket`):

```bash
gh api -H "Accept: application/vnd.github.raw" /repos/oursky/skyrocket/contents/install.sh | bash
skyrocket auth login
skyrocket deploy
```

There is no extra deploy flag besides `--env`.

Skyrocket tags images `git-<HEAD sha>`. Helm only restarts pods when that **tag** changes. Uncommitted deploys rebuild and push, then leave the old pods running. **Commit, then deploy**, or the live site stays on the first digest for that commit.

### 6. Check the rollout

Use one `curl` per line (zsh will try to run pasted `#` comments unless `interactivecomments` is on):

```bash
curl -sS https://<project>.oursky.dev/api/healthz
```

Expect JSON, not Storybook `index.html`. Then open Storybook → **Comments** and post / resolve / delete.

`POST .../resolve` returning `{ "error": "not found" }` (not `comment not found`) means the **api** image is still an old Node server without those routes.

## Pitfalls

- **Skyrocket + `github:` this repo → `npm ci` exit 128.** Vendor with `file:` (see install). That is the default Oursky setup, not an edge case.
- Installing this addon does not deploy an API. Without `/api/comments`, the panel shows an error.
- `dockerfile: Dockerfile` plus `context: comments-api` builds the **Storybook** Dockerfile against the API folder.
- PATCH/DELETE may be dropped by proxies; this addon uses **POST** for resolve and delete.
- Changing files without a new git commit does not roll Skyrocket pods.
