package passkey

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// androidOriginPrefix is the origin form an Android app reports for a passkey
// ceremony, followed by the unpadded base64url SHA-256 digest of its signing
// certificate.
const androidOriginPrefix = "android:apk-key-hash:"

// EffectiveNativeApps returns every client's x_native_apps in config order
// (client order, then entry order). It does not consult
// authentication.identities, because a custom authentication flow can offer
// passkeys without it. See docs/specs/passkey-webview.md.
func EffectiveNativeApps(oauth *config.OAuthConfig) []*config.OAuthClientNativeAppConfig {
	if oauth == nil {
		return nil
	}

	var apps []*config.OAuthClientNativeAppConfig
	for _, client := range oauth.Clients {
		apps = append(apps, client.NativeApps...)
	}
	return apps
}

// AndroidOrigin converts a colon-separated SHA-256 certificate fingerprint to
// the origin an Android app reports for it.
func AndroidOrigin(sha256CertFingerprint string) (string, error) {
	digest, err := hex.DecodeString(strings.ReplaceAll(sha256CertFingerprint, ":", ""))
	if err != nil {
		return "", fmt.Errorf("passkey: invalid SHA-256 certificate fingerprint: %w", err)
	}
	if len(digest) != 32 {
		return "", fmt.Errorf("passkey: SHA-256 certificate fingerprint must be 32 bytes, got %d", len(digest))
	}
	return androidOriginPrefix + base64.RawURLEncoding.EncodeToString(digest), nil
}
