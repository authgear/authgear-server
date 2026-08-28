package idpsession

import (
	"time"

	"github.com/authgear/authgear-server/pkg/api/model"
	"github.com/authgear/authgear-server/pkg/lib/authn/authenticationinfo"
	"github.com/authgear/authgear-server/pkg/lib/session"
	"github.com/authgear/authgear-server/pkg/lib/session/access"
	"github.com/authgear/authgear-server/pkg/util/geoip"
	"github.com/authgear/authgear-server/pkg/util/setutil"
)

type IDPSession struct {
	ID    string `json:"id"`
	AppID string `json:"app_id"`

	// CreatedAt is the timestamp that the user was initially authenticated at.
	CreatedAt time.Time `json:"created_at"`
	// Authenticated is the timestamp that the user was authenticated at.
	// It is equal to CreatedAt if the user has not reauthenticated at all.
	AuthenticatedAt time.Time     `json:"authenticated_at"`
	Attrs           session.Attrs `json:"attrs"`

	AccessInfo access.Info `json:"access_info"`

	TokenHash string `json:"token_hash"`

	ParticipatedSAMLServiceProviderIDs []string `json:"participated_saml_service_provider_ids,omitempty"`

	// ExpireAtForResolvedSession is a transient field that tells when the session will exire at, computed now.
	// Note that ExpireAtForResolvedSession will keep changing if idle timeout is enabled.
	// This is NOT supposed to be stored, hence it is json-ignored.
	ExpireAtForResolvedSession time.Time `json:"-"`
}

var _ session.ResolvedSession = &IDPSession{}
var _ session.ListableSession = &IDPSession{}

func (s *IDPSession) Session()         {}
func (s *IDPSession) ListableSession() {}

func (s *IDPSession) SessionID() string         { return s.ID }
func (s *IDPSession) SessionType() session.Type { return session.TypeIdentityProvider }

// GetTokenType is always TokenTypeCookies: an IDPSession is resolved as a
// ResolvedSession either via the IDP session cookie, or as the session
// backing an access grant with no offline grant/refresh token
// (GrantSessionKindSession in oauth.Resolver, selected via an
// id_token_hint whose sid decodes to session.TypeIdentityProvider). The
// latter never carries a third-party client: such an id_token can only be
// minted by TokenHandler.handleIDToken (handler_token.go), which requires
// client.HasFullAccessScope() -- true only for first-party SPA/Native/
// TraditionalWeb clients (OAuthClientApplicationType.HasFullAccessScope,
// pkg/lib/config/oauth.go), never for a third-party (static or dynamic) or
// M2M client. Every other id_token issued via authorization_code/
// refresh_token encodes an offline-grant SID instead (see the other
// EncodeSID call sites in handler_token.go), so it can never produce this
// grant kind at all.
func (s *IDPSession) GetTokenType() session.TokenType { return session.TokenTypeCookies }

func (s *IDPSession) GetCreatedAt() time.Time               { return s.CreatedAt }
func (s *IDPSession) GetExpireAt() time.Time                { return s.ExpireAtForResolvedSession }
func (s *IDPSession) GetAuthenticatedAt() time.Time         { return s.AuthenticatedAt }
func (s *IDPSession) GetClientID() string                   { return "" }
func (s *IDPSession) GetAccessInfo() *access.Info           { return &s.AccessInfo }
func (s *IDPSession) GetDeviceInfo() (map[string]any, bool) { return nil, false }
func (s *IDPSession) GetUserID() string                     { return s.Attrs.UserID }
func (s *IDPSession) GetOIDCAMR() ([]string, bool)          { return s.Attrs.GetAMR() }

func (s *IDPSession) ToAPIModel() *model.Session {
	ua := model.ParseUserAgent(s.AccessInfo.LastAccess.UserAgent)
	amr, _ := s.Attrs.GetAMR()
	apiModel := &model.Session{
		Meta: model.Meta{
			ID:        s.ID,
			CreatedAt: s.CreatedAt,
			// TODO(session): Session Updated At should be the time user actively updates it.
			UpdatedAt: s.AccessInfo.LastAccess.Timestamp,
		},
		Type: model.SessionTypeIDP,

		AMR: amr,

		LastAccessedAt:   s.AccessInfo.LastAccess.Timestamp,
		CreatedByIP:      s.AccessInfo.InitialAccess.RemoteIP,
		LastAccessedByIP: s.AccessInfo.LastAccess.RemoteIP,

		DisplayName: ua.Format(),
		UserAgent:   ua.Raw,
	}

	ipInfo, ok := geoip.IPString(s.AccessInfo.LastAccess.RemoteIP)
	if ok {
		apiModel.LastAccessedByIPCountryCode = ipInfo.CountryCode
		apiModel.LastAccessedByIPEnglishCountryName = ipInfo.EnglishCountryName
	}

	return apiModel
}

func (s *IDPSession) GetAuthenticationInfo() authenticationinfo.T {
	amr, _ := s.GetOIDCAMR()
	return authenticationinfo.T{
		UserID:          s.GetUserID(),
		AuthenticatedAt: s.GetAuthenticatedAt(),
		AMR:             amr,
	}
}

func (s *IDPSession) CreateNewAuthenticationInfoByThisSession() authenticationinfo.T {
	amr, _ := s.GetOIDCAMR()
	return authenticationinfo.T{
		UserID:                     s.GetUserID(),
		AuthenticatedAt:            s.GetAuthenticatedAt(),
		AMR:                        amr,
		AuthenticatedBySessionType: string(s.SessionType()),
		AuthenticatedBySessionID:   s.SessionID(),
	}
}

func (s *IDPSession) SSOGroupIDPSessionID() string {
	return s.SessionID()
}

// IsSameSSOGroup returns true when the session argument
// - is the same idp session
// - is sso enabled offline grant that in the same sso group
func (s *IDPSession) IsSameSSOGroup(ss session.SessionBase) bool {
	if s.EqualSession(ss) {
		return true
	}
	if s.SSOGroupIDPSessionID() == "" {
		return false
	}
	return s.SSOGroupIDPSessionID() == ss.SSOGroupIDPSessionID()
}

func (s *IDPSession) EqualSession(ss session.SessionBase) bool {
	return s.SessionID() == ss.SessionID() && s.SessionType() == ss.SessionType()
}

func (s *IDPSession) GetParticipatedSAMLServiceProviderIDsSet() setutil.Set[string] {
	return setutil.NewSetFromSlice(s.ParticipatedSAMLServiceProviderIDs, setutil.Identity)
}
