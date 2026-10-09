package passkey

import (
	"context"
	"net/http"
	"net/url"

	"github.com/go-webauthn/webauthn/protocol"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/urlutil"
)

type TranslationService interface {
	RenderText(ctx context.Context, key string, args any) (string, error)
}

type ConfigService struct {
	Request            *http.Request
	TrustProxy         config.TrustProxy
	TranslationService TranslationService
	OAuthConfig        *config.OAuthConfig
}

// makeRPOrigins returns every origin from which a ceremony for this project may
// legitimately be performed.
//
// The public origin is always included: a request can only reach a handler on
// that host, because PublicOriginMiddleware redirects anything else.
//
// A Custom UI is the exception. It is served from the customer's own origin and
// drives the Authentication Flow API cross-origin, so the ceremony happens there
// and the browser reports that origin rather than ours. Its origin is taken from
// x_custom_ui_uri, the same field the CORS allowlist is derived from, so one
// declaration covers both.
//
// Deliberately not included: http.allowed_origins, redirect_uris and
// x_pre_authenticated_url_allowed_origins. Those say an origin may call the API
// or receive a callback; neither means it may present credentials for a user.
//
// An Android app listed in x_native_apps reports android:apk-key-hash:<hash>
// instead of a web origin, so each listed certificate adds one. The returned
// map records which package names were listed with each such origin.
func (s *ConfigService) makeRPOrigins(publicOrigin url.URL) ([]string, map[string][]string) {
	publicOriginString := publicOrigin.String()
	origins := []string{publicOriginString}

	if s.OAuthConfig == nil {
		return origins, nil
	}

	seen := map[string]struct{}{publicOriginString: {}}
	for _, client := range s.OAuthConfig.Clients {
		if client.CustomUIURI == "" {
			continue
		}
		u, err := url.Parse(client.CustomUIURI)
		if err != nil {
			// Unparseable values are rejected by the config schema's "uri"
			// format; skip rather than fail the ceremony.
			continue
		}
		// x_custom_ui_uri is a full URI. Reduce it to scheme and host, keeping
		// any port, since a port is part of an origin.
		origin := urlutil.ExtractOrigin(u).String()
		if _, ok := seen[origin]; ok {
			continue
		}
		seen[origin] = struct{}{}
		origins = append(origins, origin)
	}

	androidPackageNames := map[string][]string{}
	origins = s.appendAndroidOrigins(origins, seen, androidPackageNames)

	return origins, androidPackageNames
}

// appendAndroidOrigins appends the android:apk-key-hash origin of every listed
// Android certificate, and records its package names in androidPackageNames.
func (s *ConfigService) appendAndroidOrigins(origins []string, seen map[string]struct{}, androidPackageNames map[string][]string) []string {
	for _, app := range EffectiveNativeApps(s.OAuthConfig) {
		if app.Platform != config.OAuthClientNativeAppPlatformAndroid {
			continue
		}
		for _, fingerprint := range app.SHA256CertFingerprints {
			origin, err := AndroidOrigin(fingerprint)
			if err != nil {
				// Malformed fingerprints are rejected by the config schema;
				// skip rather than fail the ceremony.
				continue
			}
			androidPackageNames[origin] = append(androidPackageNames[origin], app.PackageName)
			if _, ok := seen[origin]; ok {
				continue
			}
			seen[origin] = struct{}{}
			origins = append(origins, origin)
		}
	}
	return origins
}

func (s *ConfigService) MakeConfig(ctx context.Context) (*Config, error) {
	origin := url.URL{
		Scheme: httputil.GetProto(s.Request, bool(s.TrustProxy)),
		Host:   httputil.GetHost(s.Request, bool(s.TrustProxy)),
	}

	appName, err := s.TranslationService.RenderText(ctx, "app.name", nil)
	if err != nil {
		return nil, err
	}

	rpOrigins, androidPackageNames := s.makeRPOrigins(origin)

	return &Config{
		RPDisplayName: appName,

		// The RPID must be a domain only.
		RPID: origin.Hostname(),
		// Origins must be the actual origins as observed by the browser.
		RPOrigins:           rpOrigins,
		AndroidPackageNames: androidPackageNames,

		AttestationPreference: protocol.PreferDirectAttestation,
		AuthenticatorSelection: protocol.AuthenticatorSelection{
			// AuthenticatorAttachment is intentionally left blank so that the user
			// can choose "platform" or "cross-platform" attachment.
			// This means the authenticator can either be on-device or off-device.
			// AuthenticatorAttachment:,

			// ResidentKey is "preferred" to maximize compatibility across platforms.
			// On iOS 16, client-side discoverable credential is created if the value is set to "preferred" or "required".
			// So the created credential can be later on used with Autofill.
			// On Android, client-side discoverable credential is NOT supported.
			// Therefore, specifying "required" will cause the ceremony to fail.
			ResidentKey: protocol.ResidentKeyRequirementPreferred,
			// RequireResidentKey is a deprecated field.
			// https://www.w3.org/TR/webauthn-2/#dom-authenticatorselectioncriteria-requireresidentkey
			// It MUST BE true if ResidentKey is "required".
			// Since we set ResidentKey to "preferred", it MUST BE left blank.
			// ResidentKey:,

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
