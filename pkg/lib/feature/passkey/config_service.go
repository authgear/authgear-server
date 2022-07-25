package passkey

import (
	"net/http"
	"net/url"

	"github.com/duo-labs/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type TranslationService interface {
	RenderText(key string, args interface{}) (string, error)
}

type ConfigService struct {
	Request            *http.Request
	TrustProxy         config.TrustProxy
	TranslationService TranslationService
}

func (s *ConfigService) MakeConfig() (*Config, error) {
	origin := url.URL{
		Scheme: httputil.GetProto(s.Request, bool(s.TrustProxy)),
		Host:   httputil.GetHost(s.Request, bool(s.TrustProxy)),
	}

	appName, err := s.TranslationService.RenderText("app.name", nil)
	if err != nil {
		return nil, err
	}

	requireResidentKey := true
	return &Config{
		RPDisplayName: appName,

		// The RPID must be a domain only.
		RPID: origin.Hostname(),
		// Origin must be the actual origin as observed by the browser.
		RPOrigin: origin.String(),

		AttestationPreference: protocol.PreferDirectAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			// AuthenticatorAttachment is intentionally left blank so that the user
			// can choose "platform" or "cross-platform" attachment.
			// This means the authenticator can either be on-device or off-device.
			// AuthenticatorAttachment:,

			// ResidentKey is "required" means we want the credential to be client-side discoverable.
			// This implies we do not need to identity the user first and find out allowCredentials.
			// The outcome is that we can present a flow that the user signs in by selecting
			// credentials, without the need of first entering their email address.
			ResidentKey: protocol.ResidentKeyRequirementRequired,
			// RequireResidentKey is a deprecated field.
			// https://www.w3.org/TR/webauthn-2/#dom-authenticatorselectioncriteria-requireresidentkey
			// It MUST BE true if ResidentKey is "required".
			RequireResidentKey: &requireResidentKey,

			// https://www.w3.org/TR/webauthn-2/#user-verification
			// Per the WWDC video https://developer.apple.com/videos/play/wwdc2022/10092/ at 19:12
			// UserVerification MUST be kept as preferred for the best user experience
			// regardless of whether biometric is available.
			UserVerification: protocol.VerificationPreferred,
		},

		// For modal, the timeout is 5 minutes which is relatively short.
		MediationModalTimeout: int(duration.Short.Milliseconds()),

		// For conditional, the timeout is 1 hour which is long.
		MediationConditionalTimeout: int(duration.PerHour.Milliseconds()),
	}, nil
}
