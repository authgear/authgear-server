# DCR Part 1 — Initial Access Token (IAT) Management

Spec: [docs/specs/dcr.md](../../specs/dcr.md) — [Initial Access Token](../../specs/dcr.md#initial-access-token), [Admin API — IAT management](../../specs/dcr.md#iat-management).

## 1. Goal / Scope

Implement storage and Admin API management for Initial Access Tokens (IATs), independent of the registration endpoint itself.

In scope:

- `_auth_oauth_initial_access_token` DB table + migration.
- `pkg/lib/dcr` package: `InitialAccessToken` domain type, `Store`, `Commands`, `Queries`, token generation/hashing.
- Admin GraphQL: `initialAccessTokens` query, `createInitialAccessToken` / `revokeInitialAccessToken` mutations.

Out of scope (later parts):

- `POST /oauth2/register` and everything that consumes an IAT to authorize registration (Part 2).
- The persisted dynamic-client record (`pkg/lib/oauthclient`, `_auth_oauth_client`), `dynamicClients` query, `deleteDynamicClient` mutation (Part 2).
- `oauth.dynamic_client_registration.*` authgear.yaml config section (introduced in Part 2, since nothing in Part 1 reads it).
- `oauth.dynamic_client_registration.maximum_clients` (explicitly deferred by the user to be done last, after Part 4).

## 2. Spec Deviation — `token_type` column

dcr.md's SQL for `_auth_oauth_initial_access_token` (docs/specs/dcr.md lines 303-312) has no column recording whether the token is first-party or third-party. But the GraphQL `InitialAccessToken` type (docs/specs/dcr.md lines 517-522) exposes `type: InitialAccessTokenType!`, and the token is stored **hashed only** — the type cannot be recovered from `token_hash` after creation, and it cannot be derived from the prefix either, because the prefix is part of the plaintext, not the hash. The type must be persisted as its own column.

**Decision:** add `token_type text NOT NULL` to the table. Update docs/specs/dcr.md's SQL block in the same PR that adds the migration (small doc fix commit, see [Atomic Commit Plan](#8-atomic-commit-plan)).

## 3. Data Model & Migration

New file: `cmd/authgear/cmd/cmddatabase/migrations/authgear/20260817120000-add_oauth_initial_access_token.sql`

```sql
-- +migrate Up

CREATE TABLE _auth_oauth_initial_access_token (
  id text PRIMARY KEY,
  app_id text NOT NULL,
  created_at timestamp without time zone NOT NULL,
  expires_at timestamp without time zone NOT NULL,
  token_type text NOT NULL,
  token_hash text NOT NULL
);
CREATE UNIQUE INDEX _auth_oauth_initial_access_token_hash_unique ON _auth_oauth_initial_access_token USING btree (app_id, token_hash);
CREATE INDEX _auth_oauth_initial_access_token_app_id_created_at ON _auth_oauth_initial_access_token USING btree (app_id, created_at);

-- +migrate Down

DROP INDEX IF EXISTS _auth_oauth_initial_access_token_app_id_created_at;
DROP INDEX IF EXISTS _auth_oauth_initial_access_token_hash_unique;
DROP TABLE IF EXISTS _auth_oauth_initial_access_token;
```

`token_type` stores the literal string `"THIRD_PARTY"` or `"FIRST_PARTY"` (matches the GraphQL enum values verbatim, avoids a separate mapping table). `app_id` scoping and the `(app_id, token_hash)` unique index follow the same convention as `_auth_resource_uri_unique` in `cmd/authgear/cmd/cmddatabase/migrations/authgear/20250710143552-add_resource_scope.sql`.

Revocation is a hard delete (`DELETE FROM _auth_oauth_initial_access_token WHERE id = ?`) — there is no soft-delete/revoked column in the spec's schema, and none is needed: a revoked IAT must never again validate, and once deleted, `GetInitialAccessTokenByHash` behaves identically to "was never issued" (see [Errors](#6-error-handling), consumed by Part 2).

## 4. Domain Model (`pkg/api/model`)

New file: `pkg/api/model/oauth_initial_access_token.go`

```go
package model

import "time"

type OAuthInitialAccessTokenType string

const (
	OAuthInitialAccessTokenTypeThirdParty OAuthInitialAccessTokenType = "THIRD_PARTY"
	OAuthInitialAccessTokenTypeFirstParty OAuthInitialAccessTokenType = "FIRST_PARTY"
)

// OAuthInitialAccessToken is the Admin API-facing representation.
//
// model.Meta MUST be embedded even though the GraphQL InitialAccessToken type
// (docs/specs/dcr.md) exposes no updatedAt field. pkg/admin/graphql's
// entityIDField and entityCreatedAtField both do an unchecked
// obj.(EntityRef) assertion, where EntityRef is `GetMeta() model.Meta`
// (pkg/admin/graphql/entity.go:41-70) — a model without Meta panics at
// resolve time, not compile time. Embedding Meta is orthogonal to which
// fields the GraphQL object declares: §7.1 simply does not add an
// "updatedAt" field, and does not implement entityInterface. UpdatedAt is
// set equal to CreatedAt (an IAT is immutable) and never surfaced.
type OAuthInitialAccessToken struct {
	model.Meta
	ExpiresAt time.Time
	Type      OAuthInitialAccessTokenType
}
```

## 5. `pkg/lib/dcr` Package

New package, mirrors the `pkg/lib/resourcescope` layout (`Store` + `Commands` + `Queries`, internal struct with a `ToModel()` method — see `pkg/lib/resourcescope/resource.go` and `pkg/lib/resourcescope/store_resource.go`).

`pkg/lib/dcr` hosts the **IAT domain and the DCR registration protocol only**. The persisted client record lives in a separate, source-agnostic `pkg/lib/oauthclient` package, because the same table serves CIMD-resolved clients — see [Part 2 §3.1 and §4](2026-08-17-02-client-registration.md).

### 5.1 `pkg/lib/dcr/initial_access_token.go`

```go
package dcr

import "time"

type InitialAccessTokenType string

const (
	InitialAccessTokenTypeThirdParty InitialAccessTokenType = "THIRD_PARTY"
	InitialAccessTokenTypeFirstParty InitialAccessTokenType = "FIRST_PARTY"
)

const DefaultInitialAccessTokenExpiresIn = 3600 // seconds, per spec "e.g. 3600"

type InitialAccessToken struct {
	ID        string
	CreatedAt time.Time
	ExpiresAt time.Time
	Type      InitialAccessTokenType
	TokenHash string
}

func (t *InitialAccessToken) ToModel() *model.OAuthInitialAccessToken {
	return &model.OAuthInitialAccessToken{
		Meta: model.Meta{
			ID:        t.ID,
			CreatedAt: t.CreatedAt,
			// An IAT is immutable; UpdatedAt exists only to satisfy
			// model.Meta / EntityRef and is never exposed in GraphQL.
			UpdatedAt: t.CreatedAt,
		},
		ExpiresAt: t.ExpiresAt,
		Type:      model.OAuthInitialAccessTokenType(t.Type),
	}
}

type NewInitialAccessTokenOptions struct {
	ExpiresIn *int // seconds; nil means DefaultInitialAccessTokenExpiresIn
	Type      InitialAccessTokenType
}
```

### 5.2 `pkg/lib/dcr/token.go` — prefixed token generation

Reuses the existing `pkg/util/rand` + `pkg/util/crypto` primitives already used by `oauth.GenerateToken()` / `oauth.HashToken()` (`pkg/lib/oauth/token.go`), rather than inventing a new RNG helper.

```go
package dcr

import (
	"github.com/authgear/authgear-server/pkg/util/crypto"
	"github.com/authgear/authgear-server/pkg/util/rand"
)

const (
	tokenAlphabet         = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789-_"
	tokenRandomLength     = 22 // matches spec: "22 chars URL-safe base64 (16 bytes)"
	IATPrefixThirdParty   = "iat_tp_"
	IATPrefixFirstParty   = "iat_fp_"
	// NOTE: superseded. DCRClientIDPrefix moves to
	// pkg/lib/oauthclient/client_id.go in Part 2 §4, together with
	// GenerateDCRClientID(), so that pkg/lib/oauthclient never has to import
	// pkg/lib/dcr. Do not add it here.
	// DCRClientIDPrefix = "dcrc_"
)

// GenerateInitialAccessToken returns the plaintext token (returned to the
// caller once) and its SHA-256 hash (persisted).
func GenerateInitialAccessToken(t InitialAccessTokenType) (plaintext string, hash string) {
	prefix := IATPrefixThirdParty
	if t == InitialAccessTokenTypeFirstParty {
		prefix = IATPrefixFirstParty
	}
	plaintext = prefix + rand.StringWithAlphabet(tokenRandomLength, tokenAlphabet, rand.SecureRand)
	hash = crypto.SHA256String(plaintext)
	return
}

func HashInitialAccessToken(plaintext string) string {
	return crypto.SHA256String(plaintext)
}
```

**`DCRClientIDPrefix` is deliberately *not* declared here.** An earlier draft put it in this file, on the grounds that this is the natural home for all DCR prefix constants. It belongs instead with the client record and the resolver in `pkg/lib/oauthclient/client_id.go` (Part 2 §4): the resolver is its main consumer, and keeping it out of `dcr` is what lets `pkg/lib/oauthclient` avoid importing `pkg/lib/dcr` at all — which is what makes the store and the resolver live in one package without an import cycle. Part 1 declares only the two IAT prefixes; nothing in Part 1 needs a client-id prefix.

### 5.3 `pkg/lib/dcr/store.go`

```go
package dcr

import (
	"github.com/authgear/authgear-server/pkg/lib/infra/db/appdb"
	"github.com/authgear/authgear-server/pkg/util/clock"
)

type Store struct {
	SQLBuilder  *appdb.SQLBuilderApp
	SQLExecutor *appdb.SQLExecutor
	Clock       clock.Clock
}
```

These are the exact types used by `pkg/lib/resourcescope/store.go` (verified). `SQLBuilderApp` is the tenant-scoped builder, so `app_id` scoping on every query is automatic the same way it is for `_auth_resource`.

### 5.4 `pkg/lib/dcr/store_initial_access_token.go`

Methods, following `pkg/lib/resourcescope/store_resource.go`'s exact style (`uuid.NewString()` for `id`, `s.SQLBuilder.Insert(...)`/`Select(...)`/`Delete(...)`, `databaseutil.IsDuplicateKeyError` where relevant):

- `func (s *Store) NewInitialAccessToken(options *NewInitialAccessTokenOptions, tokenHash string) *InitialAccessToken` — builds the struct with `ID: uuid.NewString()`, `CreatedAt: s.Clock.NowUTC()`, `ExpiresAt` computed from `options.ExpiresIn` (or `DefaultInitialAccessTokenExpiresIn`), `Type: options.Type`, `TokenHash: tokenHash`. Does not touch the DB.
- `func (s *Store) CreateInitialAccessToken(ctx context.Context, t *InitialAccessToken) error` — `INSERT INTO _auth_oauth_initial_access_token (id, created_at, expires_at, token_type, token_hash) VALUES (...)`.
- `func (s *Store) GetInitialAccessTokenByID(ctx context.Context, id string) (*InitialAccessToken, error)` — `SELECT ... WHERE id = ?`; `sql.ErrNoRows` → `ErrInitialAccessTokenNotFound`.
- `func (s *Store) GetInitialAccessTokenByHash(ctx context.Context, tokenHash string) (*InitialAccessToken, error)` — `SELECT ... WHERE token_hash = ?`; `sql.ErrNoRows` → `ErrInitialAccessTokenNotFound`. Does **not** filter `expires_at` here — expiry is a `Queries`/`Commands`-level concern (see below) so the same row-fetch is reusable for both "does this hash exist at all" and "is it still valid" checks.
- `func (s *Store) DeleteInitialAccessToken(ctx context.Context, id string) error` — `DELETE ... WHERE id = ?`; 0 rows affected → `ErrInitialAccessTokenNotFound`.
- `func (s *Store) ListActiveInitialAccessTokens(ctx context.Context) ([]*InitialAccessToken, error)` — `SELECT ... WHERE expires_at > ? ORDER BY created_at DESC` (now passed in via `s.Clock.NowUTC()`).

### 5.5 `pkg/lib/dcr/errors.go`

```go
package dcr

import "errors"

var ErrInitialAccessTokenNotFound = errors.New("initial access token not found")
```

### 5.6 `pkg/lib/dcr/commands.go`

```go
package dcr

type Commands struct {
	Store *Store
}

func (c *Commands) CreateInitialAccessToken(ctx context.Context, options *NewInitialAccessTokenOptions) (plaintext string, iat *model.OAuthInitialAccessToken, err error) {
	plaintext, hash := GenerateInitialAccessToken(options.Type)
	t := c.Store.NewInitialAccessToken(options, hash)
	if err := c.Store.CreateInitialAccessToken(ctx, t); err != nil {
		return "", nil, err
	}
	return plaintext, t.ToModel(), nil
}

func (c *Commands) RevokeInitialAccessToken(ctx context.Context, id string) error {
	return c.Store.DeleteInitialAccessToken(ctx, id)
}
```

### 5.7 `pkg/lib/dcr/queries.go`

```go
type Queries struct {
	Store *Store
	Clock clock.Clock
}

func (q *Queries) GetInitialAccessTokenByID(ctx context.Context, id string) (*model.OAuthInitialAccessToken, error) {
	t, err := q.Store.GetInitialAccessTokenByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return t.ToModel(), nil
}

func (q *Queries) ListInitialAccessTokens(ctx context.Context) ([]*model.OAuthInitialAccessToken, error) {
	ts, err := q.Store.ListActiveInitialAccessTokens(ctx)
	if err != nil {
		return nil, err
	}
	models := make([]*model.OAuthInitialAccessToken, len(ts))
	for i, t := range ts {
		models[i] = t.ToModel()
	}
	return models, nil
}

// ValidateAndGetByToken hashes the given plaintext bearer token, looks it up,
// and returns ErrInitialAccessTokenNotFound if it does not exist OR has
// expired (both cases must be indistinguishable to the caller — Part 2's
// registration handler maps both to `invalid_initial_access_token`).
func (q *Queries) ValidateAndGetByToken(ctx context.Context, plaintext string) (*model.OAuthInitialAccessToken, error) {
	hash := HashInitialAccessToken(plaintext)
	t, err := q.Store.GetInitialAccessTokenByHash(ctx, hash)
	if err != nil {
		return nil, err
	}
	if !t.ExpiresAt.After(q.Clock.NowUTC()) {
		return nil, ErrInitialAccessTokenNotFound
	}
	return t.ToModel(), nil
}
```

`ValidateAndGetByToken` is unused until Part 2 wires up `POST /oauth2/register`, but it is IAT domain logic and belongs with the rest of this package — declared now to keep the IAT domain cohesive, per the user's framing of Part 1 as "all things about IAT."

### 5.8 `pkg/lib/dcr/deps.go`

```go
package dcr

import "github.com/google/wire"

var DependencySet = wire.NewSet(
	wire.Struct(new(Store), "*"),
	wire.Struct(new(Commands), "*"),
	wire.Struct(new(Queries), "*"),
)
```

## 6. Error Handling

No new `apierrors` kinds are needed in Part 1: `ErrInitialAccessTokenNotFound` surfaces as a plain GraphQL error from `revokeInitialAccessToken` (matches how `resourcescope.ErrResourceNotFound` propagates from `deleteResource` — no explicit wrapping into an `apierrors.Kind` was found for that path either). Part 2 is responsible for mapping `ErrInitialAccessTokenNotFound` (from `ValidateAndGetByToken`) to the RFC 7591 `invalid_initial_access_token` (401) response — not part of this file.

## 7. Admin GraphQL Layer

### 7.1 `pkg/admin/graphql/initial_access_token.go` (new file)

Pattern source: `pkg/admin/graphql/resource.go`'s `node(...)` helper and `pkg/admin/graphql/authenticator.go`'s `graphql.NewEnum` usage.

```go
const typeInitialAccessToken = "InitialAccessToken"

var initialAccessTokenType = graphql.NewEnum(graphql.EnumConfig{
	Name: "InitialAccessTokenType",
	Values: graphql.EnumValueConfigMap{
		"THIRD_PARTY": &graphql.EnumValueConfig{Value: "THIRD_PARTY"},
		"FIRST_PARTY": &graphql.EnumValueConfig{Value: "FIRST_PARTY"},
	},
})

var nodeInitialAccessToken = node(
	graphql.NewObject(graphql.ObjectConfig{
		Name: typeInitialAccessToken,
		Interfaces: []*graphql.Interface{
			nodeDefs.NodeInterface, // NOT entityInterface — no updatedAt field per spec.
			// Note this is independent of model.Meta being embedded (§4):
			// entityIDField/entityCreatedAtField below require Meta at
			// runtime regardless of which interfaces the object declares.
		},
		Fields: graphql.Fields{
			"id":        entityIDField(typeInitialAccessToken),
			"createdAt": entityCreatedAtField(nil),
			"expiresAt": &graphql.Field{Type: graphql.NewNonNull(graphql.DateTime)},
			"type":      &graphql.Field{Type: graphql.NewNonNull(initialAccessTokenType)},
		},
	}),
	&model.OAuthInitialAccessToken{},
	func(ctx context.Context, gqlCtx *Context, id string) (any, error) {
		return gqlCtx.InitialAccessTokens.Load(ctx, id).Value, nil
	},
)
```

No `connInitialAccessToken` — the query returns a plain non-null list (`[InitialAccessToken!]!`), not a relay Connection, per docs/specs/dcr.md lines 489-491.

### 7.2 `pkg/admin/graphql/query.go` — add the `initialAccessTokens` field

```go
"initialAccessTokens": &graphql.Field{
	Description: "Returns all active (non-expired) Initial Access Tokens for the project.",
	Type:        graphql.NewNonNull(graphql.NewList(graphql.NewNonNull(nodeInitialAccessToken))),
	Resolve: func(p graphql.ResolveParams) (any, error) {
		ctx := p.Context
		gqlCtx := GQLContext(ctx)
		return gqlCtx.DCRFacade.ListInitialAccessTokens(ctx)
	},
},
```

### 7.3 `pkg/admin/graphql/initial_access_token_mutation.go` (new file)

Pattern source: `pkg/admin/graphql/session_mutation.go` (`registerMutationField`, Input/Payload naming, `relay.FromGlobalID` for ID-taking mutations).

```go
var createInitialAccessTokenInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "CreateInitialAccessTokenInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"expiresIn": &graphql.InputObjectFieldConfig{Type: graphql.Int},
		"type":      &graphql.InputObjectFieldConfig{Type: initialAccessTokenType, DefaultValue: "THIRD_PARTY"},
	},
})

var createInitialAccessTokenPayload = graphql.NewObject(graphql.ObjectConfig{
	Name: "CreateInitialAccessTokenPayload",
	Fields: graphql.Fields{
		"token":              &graphql.Field{Type: graphql.NewNonNull(graphql.String)},
		"initialAccessToken": &graphql.Field{Type: graphql.NewNonNull(nodeInitialAccessToken)},
	},
})

var _ = registerMutationField(
	"createInitialAccessToken",
	&graphql.Field{
		Type: graphql.NewNonNull(createInitialAccessTokenPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(createInitialAccessTokenInput)},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)
			var expiresIn *int
			if v, ok := input["expiresIn"].(int); ok {
				expiresIn = &v
			}
			iatType := dcr.InitialAccessTokenType(input["type"].(string))

			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			token, iat, err := gqlCtx.DCRFacade.CreateInitialAccessToken(ctx, &dcr.NewInitialAccessTokenOptions{
				ExpiresIn: expiresIn,
				Type:      iatType,
			})
			if err != nil {
				return nil, err
			}
			return map[string]any{"token": token, "initialAccessToken": iat}, nil
		},
	},
)

var revokeInitialAccessTokenInput = graphql.NewInputObject(graphql.InputObjectConfig{
	Name: "RevokeInitialAccessTokenInput",
	Fields: graphql.InputObjectConfigFieldMap{
		"id": &graphql.InputObjectFieldConfig{Type: graphql.NewNonNull(graphql.ID)},
	},
})

var revokeInitialAccessTokenPayload = graphql.NewObject(graphql.ObjectConfig{
	Name:   "RevokeInitialAccessTokenPayload",
	Fields: graphql.Fields{"ok": &graphql.Field{Type: graphql.Boolean}},
})

var _ = registerMutationField(
	"revokeInitialAccessToken",
	&graphql.Field{
		Type: graphql.NewNonNull(revokeInitialAccessTokenPayload),
		Args: graphql.FieldConfigArgument{
			"input": &graphql.ArgumentConfig{Type: graphql.NewNonNull(revokeInitialAccessTokenInput)},
		},
		Resolve: func(p graphql.ResolveParams) (any, error) {
			input := p.Args["input"].(map[string]any)
			resolvedNodeID := relay.FromGlobalID(input["id"].(string))
			if resolvedNodeID == nil || resolvedNodeID.Type != typeInitialAccessToken {
				return nil, apierrors.NewInvalid("invalid initial access token ID")
			}
			ctx := p.Context
			gqlCtx := GQLContext(ctx)
			if err := gqlCtx.DCRFacade.RevokeInitialAccessToken(ctx, resolvedNodeID.ID); err != nil {
				return nil, err
			}
			return map[string]any{"ok": true}, nil
		},
	},
)
```

No event dispatch — dcr.md states "Resource and Scope CRUD operations do not generate events" for the analogous Resource/Scope Admin API; IAT create/revoke follows the same non-event convention (no use case in the spec calls for auditing IAT lifecycle via the event system).

### 7.4 `pkg/admin/graphql/context.go` — add facade + loader

```go
type InitialAccessTokenLoader interface {
	graphqlutil.DataLoaderInterface
}

type DCRFacade interface {
	CreateInitialAccessToken(ctx context.Context, options *dcr.NewInitialAccessTokenOptions) (token string, iat *apimodel.OAuthInitialAccessToken, err error)
	RevokeInitialAccessToken(ctx context.Context, id string) error
	ListInitialAccessTokens(ctx context.Context) ([]*apimodel.OAuthInitialAccessToken, error)
}
```

Add to `Context` struct:

```go
InitialAccessTokens InitialAccessTokenLoader
...
DCRFacade DCRFacade
```

### 7.5 `pkg/admin/loader/initial_access_token.go` (new file)

Mirrors `pkg/admin/loader/resource.go` exactly: a `graphqlutil.DataLoader`-backed loader keyed by ID, backed by `dcr.Queries.GetInitialAccessTokenByID`. Wired into `pkg/admin/loader/deps.go`'s `wire.NewSet(...)` alongside `NewResourceLoader`.

### 7.6 `pkg/admin/facade/dcr.go` (new file)

```go
package facade

type DCRFacade struct {
	Commands *dcr.Commands
	Queries  *dcr.Queries
}

func (f *DCRFacade) CreateInitialAccessToken(ctx context.Context, options *dcr.NewInitialAccessTokenOptions) (string, *apimodel.OAuthInitialAccessToken, error) {
	return f.Commands.CreateInitialAccessToken(ctx, options)
}

func (f *DCRFacade) RevokeInitialAccessToken(ctx context.Context, id string) error {
	return f.Commands.RevokeInitialAccessToken(ctx, id)
}

func (f *DCRFacade) ListInitialAccessTokens(ctx context.Context) ([]*apimodel.OAuthInitialAccessToken, error) {
	return f.Queries.ListInitialAccessTokens(ctx)
}
```

Add `DCRFacade` to `pkg/admin/facade/deps.go`'s `wire.NewSet(...)` (same pattern as `ResourceScopeFacade` in that file).

### 7.7 Wiring — `pkg/admin/deps.go` and generated wire

- Add `dcr.DependencySet` to the provider set consumed by `pkg/admin` (same place `resourcescope`'s dependency set is included).
- Run `make generate` to regenerate `pkg/admin/wire_gen.go` (adds `dcrStore`, `dcrCommands`, `dcrQueries`, `dcrFacade`, `initialAccessTokenLoader` construction and wires them into `graphqlContext := &graphql.Context{...}`, `Loaders`, `Facades` structs) — **must** be committed in the same commit as the wiring change, per project convention (never hand-edit `wire_gen.go`).
- Run `make export-schemas` to regenerate the checked-in Admin GraphQL SDL/schema file(s) and portal `gentype` output that reflect the new types/fields — **must** be committed together with the schema change, not hand-edited.

## 8. Test Plan

Unit tests (Convey style, matching `pkg/lib/resourcescope/formats_test.go`'s style — see `add-go-test` skill):

- `pkg/lib/dcr/token_test.go` — `GenerateInitialAccessToken` produces the correct prefix per type, correct length, `HashInitialAccessToken` is deterministic and matches `crypto.SHA256String`.
- `pkg/lib/dcr/queries_test.go` — `ValidateAndGetByToken` returns `ErrInitialAccessTokenNotFound` for both a nonexistent hash and an expired-but-present row (use a fake `Store`/gomock or an in-memory `Clock` — check whether `resourcescope` uses gomock for `Store` in any existing test; if no precedent exists, this may need to become an e2e-covered case instead — see next bullet).

e2e tests (YAML-driven, see `write-e2e-test` skill) in `e2e/tests/`, new suite e.g. `e2e/tests/dcr_iat.yaml`:

1. `createInitialAccessToken` with default input → returns a `token` starting with `iat_tp_`, and `initialAccessToken.type == THIRD_PARTY`.
2. `createInitialAccessToken` with `type: FIRST_PARTY` → token starts with `iat_fp_`.
3. `createInitialAccessToken` with explicit `expiresIn` → `initialAccessToken.expiresAt` reflects it.
4. `initialAccessTokens` query lists newly created tokens; does not list a revoked one. **Also assert the plaintext is not re-exposed:** querying `initialAccessTokens { id createdAt expiresAt type }` succeeds, and a query that additionally selects `token` on `InitialAccessToken` is rejected by the schema as an unknown field. Repeat via `node(id: <IAT global id>)` — same result. This is the regression test for "returned once only."
5. `revokeInitialAccessToken` then `initialAccessTokens` query no longer includes it.
6. `revokeInitialAccessToken` with an ID from a different node type (e.g. a Resource's global ID) → GraphQL error (invalid ID).

## 9. Fixed Behavioral Decisions

- IAT plaintext is never persisted, never logged, and returned exactly once in `createInitialAccessToken`'s response — matches spec exactly. This is structurally enforced, not merely convention, at three layers:
  1. **Storage** — only the SHA-256 hash is written (`token_hash`); the plaintext exists solely as a local variable in `Commands.CreateInitialAccessToken` and is unrecoverable from the row.
  2. **Model** — `model.OAuthInitialAccessToken` (§4) has no token or hash field at all. `InitialAccessToken.TokenHash` lives only on the internal `pkg/lib/dcr` struct and is dropped by `ToModel()`. So no Admin API surface can accidentally expose it: there is nothing to expose.
  3. **Schema** — the `InitialAccessToken` GraphQL object (§7.1) declares exactly `id`, `createdAt`, `expiresAt`, `type`. The once-only plaintext is returned as `token` on `CreateInitialAccessTokenPayload` only (§7.3), which is reachable only from the create mutation. The `initialAccessTokens` query and the `node(id:)` lookup both resolve to `InitialAccessToken`, which has no path to a token value.

  Consequence to preserve in review: never add a `token` (or `tokenHash`) field to `InitialAccessToken`, and never widen `model.OAuthInitialAccessToken`. A re-query of an IAT — by list or by node id — returns metadata only.
- Revocation is a hard delete, not a soft-delete/status flag.
- `initialAccessTokens` is a plain list, not a paginated Connection (per the literal GraphQL SDL in the spec).
- No event is dispatched for IAT create/revoke.
- `oauth.dynamic_client_registration.maximum_clients` is explicitly excluded from this and all four DCR plan parts; it will be scoped separately after Part 4.

## 10. Atomic Commit Plan

1. **`doc: Fix _auth_oauth_initial_access_token schema and IAT examples in dcr.md`** — two fixes in docs/specs/dcr.md:
   - the SQL block gains `token_type text NOT NULL` (see [§2](#2-spec-deviation--token_type-column));
   - UC1's `createInitialAccessToken` example selects `expiresAt` as a field of `CreateInitialAccessTokenPayload`, but per the spec's own SDL the payload has only `token` and `initialAccessToken`, and `expiresAt` lives on `InitialAccessToken`. `token` (the once-only plaintext) is correctly on the payload; only the `expiresAt` selection is misplaced. Correct example:

     ```graphql
     mutation {
       createInitialAccessToken(input: { type: FIRST_PARTY, expiresIn: 3600 }) {
         token      # iat_fp_Xf2kLmNpQrStUvWx
         initialAccessToken {
           expiresAt
         }
       }
     }
     ```
2. **`[DCR] Add initial access token table and migration`** — the migration file only.
3. **`[DCR] Add pkg/lib/dcr package for Initial Access Token storage`** — `pkg/api/model/oauth_initial_access_token.go`, all of `pkg/lib/dcr/*.go` from §5, plus `pkg/lib/dcr/token_test.go` / `queries_test.go`.
4. **`[DCR] Add Admin API for Initial Access Token management`** — all of §7 (GraphQL types, query, mutations, facade, loader, context wiring) + regenerated `wire_gen.go` + regenerated GraphQL schema/gentype artifacts (`make generate`, `make export-schemas`) in the same commit.
5. **`[DCR] Add e2e tests for Initial Access Token Admin API`** — `e2e/tests/dcr_iat.yaml`.
