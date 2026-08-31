package oauthclient

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/infra/db"
	databaseutil "github.com/authgear/authgear-server/pkg/util/databaseutil"
	"github.com/authgear/authgear-server/pkg/util/graphqlutil"
)

// normalizeClientName normalizes an explicit empty string to nil so "no
// client_name given" always reads the same way (nil), regardless of whether
// the field was omitted or sent empty. The "Client <clientID>" fallback is
// never stored -- it's computed on read by Client.DisplayName(), since it's
// a pure function of ClientID, which never changes.
func normalizeClientName(clientName *string) *string {
	if clientName != nil && *clientName == "" {
		return nil
	}
	return clientName
}

func (s *Store) NewClient(options *NewClientOptions) *Client {
	now := s.Clock.NowUTC()

	clientName := normalizeClientName(options.ClientName)

	return &Client{
		ID:              uuid.NewString(),
		ClientID:        options.ClientID,
		Source:          options.Source,
		CreatedAt:       now,
		UpdatedAt:       now,
		Kind:            options.Kind,
		ApplicationType: options.ApplicationType,
		ClientName:      clientName,
		ClientURI:       options.ClientURI,
		LogoURI:         options.LogoURI,
		TOSURI:          options.TOSURI,
		PolicyURI:       options.PolicyURI,
		RedirectURIs:    options.RedirectURIs,
		GrantTypes:      options.GrantTypes,
		ResponseTypes:   options.ResponseTypes,
	}
}

func (s *Store) CreateClient(ctx context.Context, c *Client) error {
	redirectURIs, err := json.Marshal(c.RedirectURIs)
	if err != nil {
		return err
	}
	grantTypes, err := json.Marshal(c.GrantTypes)
	if err != nil {
		return err
	}
	responseTypes, err := json.Marshal(c.ResponseTypes)
	if err != nil {
		return err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_oauth_client")).
		Columns(
			"id",
			"client_id",
			"source",
			"created_at",
			"updated_at",
			"last_fetched_at",
			"kind",
			"application_type",
			"client_name",
			"client_uri",
			"logo_uri",
			"tos_uri",
			"policy_uri",
			"redirect_uris",
			"grant_types",
			"response_types",
		).
		Values(
			c.ID,
			c.ClientID,
			string(c.Source),
			c.CreatedAt,
			c.UpdatedAt,
			c.LastFetchedAt,
			string(c.Kind),
			c.ApplicationType,
			c.ClientName,
			c.ClientURI,
			c.LogoURI,
			c.TOSURI,
			c.PolicyURI,
			redirectURIs,
			grantTypes,
			responseTypes,
		)

	_, err = s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		if databaseutil.IsDuplicateKeyError(err) {
			return ErrDynamicClientDuplicateClientID
		}
		return err
	}

	return nil
}

// UpsertCIMDClientOptions carries the CIMD-specific write. Source and Kind
// are not parameters: spec § Mapping to the Unified Client Model fixes
// source: CIMD and kind: THIRD_PARTY for every CIMD client, and making them
// arguments would invite a caller to write a first-party or DCR row through
// this path.
type UpsertCIMDClientOptions struct {
	ClientID        string
	ApplicationType string
	ClientName      *string
	ClientURI       *string
	LogoURI         *string
	TOSURI          *string
	PolicyURI       *string
	RedirectURIs    []string
	GrantTypes      []string
	ResponseTypes   []string
}

// UpsertCIMDClient inserts or updates the single row for options.ClientID
// and reports whether a new row was created, so the caller can apply the
// client limit only to creations (spec § Client Limit: "A client_id that
// already has a persisted record is unaffected regardless of the limit").
//
// The WHERE clause on the DO UPDATE branch is a guard, not an optimisation:
// it makes it structurally impossible for this path to overwrite a DCR row.
// In practice a dcrc_-prefixed client_id can never equal a CIMD URL so the
// collision is unreachable, but the guard means a future change to either
// id shape cannot silently let a fetched document take over a registered
// client's row. When the guard blocks the update, ON CONFLICT ... DO UPDATE
// affects zero rows, RETURNING yields no row, and this function returns
// ErrDynamicClientSourceConflict.
//
// created_at and id are preserved on update -- they are not in the SET
// list, so RETURNING returns the ORIGINAL values on an update. id is the
// GraphQL Node id and the DataLoader key, so churning it on every refetch
// would break relay identity for a client an admin is looking at.
//
// (xmax = 0) AS created is the standard Postgres idiom for distinguishing
// an insert from an update inside ON CONFLICT DO UPDATE. It is used rather
// than a separate SELECT because the alternative has a TOCTOU window that
// would let two concurrent first-resolutions both report created = true
// and both consume a quota slot.
func (s *Store) UpsertCIMDClient(ctx context.Context, options *UpsertCIMDClientOptions) (client *Client, created bool, err error) {
	now := s.Clock.NowUTC()
	clientName := normalizeClientName(options.ClientName)

	redirectURIs, err := json.Marshal(options.RedirectURIs)
	if err != nil {
		return nil, false, err
	}
	grantTypes, err := json.Marshal(options.GrantTypes)
	if err != nil {
		return nil, false, err
	}
	responseTypes, err := json.Marshal(options.ResponseTypes)
	if err != nil {
		return nil, false, err
	}

	q := s.SQLBuilder.
		Insert(s.SQLBuilder.TableName("_auth_oauth_client")).
		Columns("id", "client_id", "source", "created_at", "updated_at",
			"last_fetched_at", "kind", "application_type", "client_name",
			"client_uri", "logo_uri", "tos_uri", "policy_uri",
			"redirect_uris", "grant_types", "response_types").
		Values(uuid.NewString(), options.ClientID, string(model.OAuthClientSourceCIMD),
			now, now, now, string(model.OAuthClientKindThirdParty),
			options.ApplicationType, clientName, options.ClientURI,
			options.LogoURI, options.TOSURI, options.PolicyURI,
			redirectURIs, grantTypes, responseTypes).
		Suffix(`ON CONFLICT (app_id, client_id) DO UPDATE SET
			updated_at = EXCLUDED.updated_at,
			last_fetched_at = EXCLUDED.last_fetched_at,
			application_type = EXCLUDED.application_type,
			client_name = EXCLUDED.client_name,
			client_uri = EXCLUDED.client_uri,
			logo_uri = EXCLUDED.logo_uri,
			tos_uri = EXCLUDED.tos_uri,
			policy_uri = EXCLUDED.policy_uri,
			redirect_uris = EXCLUDED.redirect_uris,
			grant_types = EXCLUDED.grant_types,
			response_types = EXCLUDED.response_types
		WHERE `+s.SQLBuilder.TableName("_auth_oauth_client")+`.source = ?
		RETURNING id, created_at, (xmax = 0) AS created`, string(model.OAuthClientSourceCIMD))

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, false, err
	}

	var id string
	var createdAt time.Time
	if err := row.Scan(&id, &createdAt, &created); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, ErrDynamicClientSourceConflict
		}
		return nil, false, err
	}

	return &Client{
		ID:              id,
		ClientID:        options.ClientID,
		Source:          model.OAuthClientSourceCIMD,
		CreatedAt:       createdAt,
		UpdatedAt:       now,
		LastFetchedAt:   &now,
		Kind:            model.OAuthClientKindThirdParty,
		ApplicationType: options.ApplicationType,
		ClientName:      clientName,
		ClientURI:       options.ClientURI,
		LogoURI:         options.LogoURI,
		TOSURI:          options.TOSURI,
		PolicyURI:       options.PolicyURI,
		RedirectURIs:    options.RedirectURIs,
		GrantTypes:      options.GrantTypes,
		ResponseTypes:   options.ResponseTypes,
	}, created, nil
}

func (s *Store) GetClientByClientID(ctx context.Context, clientID string) (*Client, error) {
	q := s.selectClientQuery().Where("client_id = ?", clientID)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	c, err := s.scanClient(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDynamicClientNotFound
		}
		return nil, err
	}

	return c, nil
}

func (s *Store) GetClientByID(ctx context.Context, id string) (*Client, error) {
	q := s.selectClientQuery().Where("id = ?", id)

	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return nil, err
	}

	c, err := s.scanClient(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrDynamicClientNotFound
		}
		return nil, err
	}

	return c, nil
}

func (s *Store) GetManyClientsByID(ctx context.Context, ids []string) ([]*Client, error) {
	q := s.selectClientQuery().Where("id = ANY (?)", pq.Array(ids))
	return s.queryClients(ctx, q)
}

func (s *Store) DeleteClientByClientID(ctx context.Context, clientID string) error {
	q := s.SQLBuilder.Delete(s.SQLBuilder.TableName("_auth_oauth_client")).
		Where("client_id = ?", clientID)

	result, err := s.SQLExecutor.ExecWith(ctx, q)
	if err != nil {
		return err
	}

	count, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if count != 1 {
		return ErrDynamicClientNotFound
	}

	return nil
}

type storeListClientResult struct {
	Items      []*Client
	Offset     uint64
	TotalCount uint64
}

// ListClients returns every dynamic client (all sources), ordered by
// created_at DESC — dcr.md's dynamicClients query takes only pagination
// args and distinguishes sources via the returned "source" field.
func (s *Store) ListClients(ctx context.Context, pageArgs graphqlutil.PageArgs) (*storeListClientResult, error) {
	q := s.selectClientQuery().OrderBy("created_at DESC")

	q, offset, err := db.ApplyPageArgs(q, pageArgs)
	if err != nil {
		return nil, err
	}

	clients, err := s.queryClients(ctx, q)
	if err != nil {
		return nil, err
	}

	totalCount, err := s.countClients(ctx)
	if err != nil {
		return nil, err
	}

	return &storeListClientResult{
		Items:      clients,
		Offset:     offset,
		TotalCount: totalCount,
	}, nil
}

func (s *Store) countClients(ctx context.Context) (uint64, error) {
	q := s.SQLBuilder.Select("COUNT(*)").From(s.SQLBuilder.TableName("_auth_oauth_client"))

	var count uint64
	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return 0, err
	}
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// CountClientsBySource is filtered by source deliberately: dcr.md and
// cimd.md define separate quotas counted over source: DCR and source: CIMD
// respectively. A source-less count would let one source's quota consume
// the other's.
// LockForClientCount serializes concurrent DCR/CIMD registrations for this
// app so a CountClientsBySource-then-CreateClient sequence is atomic with
// respect to the configured quota -- a plain "SELECT COUNT(*) then INSERT"
// has a TOCTOU window a Postgres COUNT(*) does not close on its own. Must be
// called inside a transaction: pg_advisory_xact_lock releases at
// transaction end, unlike the session-scoped pg_advisory_lock, which would
// leak the lock across the pooled connection's next user.
//
// The lock key is scoped per source (not per table), so a DCR registration
// and a CIMD resolution never serialize against each other.
func (s *Store) LockForClientCount(ctx context.Context, source Source) error {
	key := string(source) + ":" + string(s.AppID)
	// sq.Expr does not go through a SQLBuilder query, so it never gets the
	// "?" -> "$1" placeholder rewrite that SQLBuilder-derived queries get
	// from their PlaceholderFormat(sq.Dollar) setting -- write the dollar
	// placeholder directly.
	_, err := s.SQLExecutor.ExecWith(ctx, sq.Expr("SELECT pg_advisory_xact_lock(hashtext($1))", key))
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) CountClientsBySource(ctx context.Context, source Source) (uint64, error) {
	q := s.SQLBuilder.Select("COUNT(*)").
		From(s.SQLBuilder.TableName("_auth_oauth_client")).
		Where("source = ?", string(source))

	var count uint64
	row, err := s.SQLExecutor.QueryRowWith(ctx, q)
	if err != nil {
		return 0, err
	}
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func (s *Store) queryClients(ctx context.Context, q db.SelectBuilder) ([]*Client, error) {
	rows, err := s.SQLExecutor.QueryWith(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var clients []*Client
	for rows.Next() {
		c, err := s.scanClient(rows)
		if err != nil {
			return nil, err
		}
		clients = append(clients, c)
	}

	return clients, nil
}

func (s *Store) selectClientQuery() db.SelectBuilder {
	return s.SQLBuilder.
		Select(
			"id",
			"client_id",
			"source",
			"created_at",
			"updated_at",
			"last_fetched_at",
			"kind",
			"application_type",
			"client_name",
			"client_uri",
			"logo_uri",
			"tos_uri",
			"policy_uri",
			"redirect_uris",
			"grant_types",
			"response_types",
		).
		From(s.SQLBuilder.TableName("_auth_oauth_client"))
}

func (s *Store) scanClient(scanner db.Scanner) (*Client, error) {
	c := &Client{}

	var source string
	var kind string
	var redirectURIs []byte
	var grantTypes []byte
	var responseTypes []byte

	err := scanner.Scan(
		&c.ID,
		&c.ClientID,
		&source,
		&c.CreatedAt,
		&c.UpdatedAt,
		&c.LastFetchedAt,
		&kind,
		&c.ApplicationType,
		&c.ClientName,
		&c.ClientURI,
		&c.LogoURI,
		&c.TOSURI,
		&c.PolicyURI,
		&redirectURIs,
		&grantTypes,
		&responseTypes,
	)
	if err != nil {
		return nil, err
	}

	c.Source = Source(source)
	c.Kind = Kind(kind)

	if err := json.Unmarshal(redirectURIs, &c.RedirectURIs); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(grantTypes, &c.GrantTypes); err != nil {
		return nil, err
	}
	if err := json.Unmarshal(responseTypes, &c.ResponseTypes); err != nil {
		return nil, err
	}

	return c, nil
}
