package oauthclient

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
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
