package oauthclient

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

// Source and Kind reuse the API-facing enums rather than redeclaring them:
// the stored strings are the GraphQL enum values verbatim.
type Source = model.OAuthClientSource // "DCR" | "CIMD"
type Kind = model.OAuthClientKind     // "FIRST_PARTY" | "THIRD_PARTY"

type Client struct {
	ID              string // row uuid, GraphQL Node id
	ClientID        string // dcrc_... (DCR) or an https:// URL (CIMD)
	Source          Source
	CreatedAt       time.Time
	UpdatedAt       time.Time
	LastFetchedAt   *time.Time // CIMD only; always nil for Source == DCR
	Kind            Kind
	ApplicationType string // "web" | "native"
	ClientName      *string
	ClientURI       *string
	LogoURI         *string
	TOSURI          *string
	PolicyURI       *string
	RedirectURIs    []string
	GrantTypes      []string
	ResponseTypes   []string
}

type NewClientOptions struct {
	ClientID        string
	Source          Source
	Kind            Kind
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

// RegisteredAt is derived, not stored: client.md defines registeredAt as the
// DCR registration timestamp and null for CIMD ("there is no registration
// event, only a fetch").
func (c *Client) RegisteredAt() *time.Time {
	if c.Source != model.OAuthClientSourceDCR {
		return nil
	}
	t := c.CreatedAt
	return &t
}

// DisplayName mirrors client.md's "name" field rule: client_name, or a
// generated "Client <clientID>" fallback. ClientName is never backfilled
// with this fallback in storage (Store.NewClient stores nil when omitted)
// -- the fallback is a pure function of ClientID, so computing it here on
// every read is equivalent to storing it, without a stale copy to keep in
// sync if the generation rule ever changes.
func (c *Client) DisplayName() string {
	if c.ClientName != nil && *c.ClientName != "" {
		return *c.ClientName
	}
	return "Client " + c.ClientID
}

func derefOr(s *string, fallback string) string {
	if s == nil {
		return fallback
	}
	return *s
}

// MetadataChangedFrom reports whether applying options via UpsertCIMDClient
// would change any field a fetched document controls -- ClientName,
// ClientURI, LogoURI, TOSURI, PolicyURI, ApplicationType, RedirectURIs,
// GrantTypes, ResponseTypes. Never ID, ClientID, Source, Kind, or any
// timestamp: those are not document-derived, and CreatedAt/UpdatedAt/
// LastFetchedAt change on every refetch by construction, which would make
// this report "changed" on every call.
//
// A nil ClientName/ClientURI/LogoURI/TOSURI/PolicyURI is treated as equal
// to a pointer to "": Store.NewClient stores an explicit "" as nil, so a
// document that switches between omitting a field and sending "" must not
// register as a change. RedirectURIs/GrantTypes/ResponseTypes compare as
// sets: order carries no meaning in them, so a mere reorder is not a
// change either.
func (c *Client) MetadataChangedFrom(options *UpsertCIMDClientOptions) bool {
	sameString := func(a, b *string) bool {
		if a != nil && *a == "" {
			a = nil
		}
		if b != nil && *b == "" {
			b = nil
		}
		return derefOr(a, "") == derefOr(b, "")
	}
	return !sameString(c.ClientName, options.ClientName) ||
		!sameString(c.ClientURI, options.ClientURI) ||
		!sameString(c.LogoURI, options.LogoURI) ||
		!sameString(c.TOSURI, options.TOSURI) ||
		!sameString(c.PolicyURI, options.PolicyURI) ||
		c.ApplicationType != options.ApplicationType ||
		!setutil.SetsEqual(c.RedirectURIs, options.RedirectURIs) ||
		!setutil.SetsEqual(c.GrantTypes, options.GrantTypes) ||
		!setutil.SetsEqual(c.ResponseTypes, options.ResponseTypes)
}
