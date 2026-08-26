package model

import "time"

type OAuthClientSource string

const (
	OAuthClientSourceStatic OAuthClientSource = "STATIC"
	OAuthClientSourceDCR    OAuthClientSource = "DCR"
	OAuthClientSourceCIMD   OAuthClientSource = "CIMD" // reserved; unused until cimd.md ships
)

type OAuthClientKind string

const (
	OAuthClientKindFirstParty OAuthClientKind = "FIRST_PARTY"
	OAuthClientKindThirdParty OAuthClientKind = "THIRD_PARTY"
)

// OAuthClient is the Admin API-facing representation of a dynamic (DCR or
// CIMD) client. Static (authgear.yaml) clients are not represented by this
// type at all — dcr.md's dynamicClients query explicitly excludes them.
//
// model.Meta carries the internal row uuid (NOT the OAuth client_id) and
// the creation timestamp. It is required for two independent reasons:
//
//  1. pkg/admin/graphql's entityIDField / entityCreatedAtField do an
//     unchecked obj.(EntityRef) assertion, where EntityRef is
//     `GetMeta() model.Meta` — a model without Meta panics at resolve time,
//     not compile time.
//  2. The relay Node global id and the DataLoader key are both the row
//     uuid, since client_id is an externally-chosen-looking string that
//     does not belong in a global id.
//
// Meta.UpdatedAt is set equal to Meta.CreatedAt for DCR clients (a DCR
// client is immutable until RFC 7592 lands) and is never exposed — the
// GraphQL OAuthClient type declares no "updatedAt" field and does not
// implement the Entity interface.
type OAuthClient struct {
	Meta

	ClientID                               string
	Source                                 OAuthClientSource
	Kind                                   OAuthClientKind
	IsConfidential                         bool
	IsServiceClient                        bool
	ApplicationType                        *string
	Name                                   string
	ClientName                             *string
	ClientURI                              *string
	LogoURI                                *string
	TOSURI                                 *string
	PolicyURI                              *string
	RedirectURIs                           []string
	PostLogoutRedirectURIs                 []string
	GrantTypes                             []string
	ResponseTypes                          []string
	AccessTokenLifetimeSeconds             int
	RefreshTokenLifetimeSeconds            int
	RefreshTokenIdleTimeoutEnabled         bool
	RefreshTokenIdleTimeoutSeconds         int
	RefreshTokenRotationEnabled            bool
	IssueJWTAccessToken                    bool
	MaxConcurrentSession                   int
	CustomUIURI                            *string
	App2appEnabled                         bool
	App2appInsecureDeviceKeyBindingEnabled bool
	DPoPDisabled                           bool
	PreAuthenticatedURLEnabled             bool
	PreAuthenticatedURLAllowedOrigins      []string
	ReplaceProjectLogoWithLogoURI          bool
	RegisteredAt                           *time.Time
	LastFetchedAt                          *time.Time
}
