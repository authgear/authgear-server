package passkey

import (
	"crypto/tls"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

// A fingerprint and origin pair observed on a device: the Android debug
// certificate and the origin the Android WebView reported for it.
const (
	testAndroidFingerprint = "E5:31:72:44:EF:62:B2:30:13:04:EB:89:9A:02:E7:39:21:86:99:E3:BF:28:F6:D8:48:9F:4C:8B:80:F0:E1:84"
	testAndroidOrigin      = "android:apk-key-hash:5TFyRO9isjATBOuJmgLnOSGGmeO_KPbYSJ9Mi4Dw4YQ"
)

func TestAndroidOrigin(t *testing.T) {
	Convey("AndroidOrigin", t, func() {
		Convey("converts a fingerprint to the origin an Android app reports", func() {
			origin, err := AndroidOrigin(testAndroidFingerprint)
			So(err, ShouldBeNil)
			So(origin, ShouldEqual, testAndroidOrigin)
		})

		Convey("accepts lower case", func() {
			origin, err := AndroidOrigin(strings.ToLower(testAndroidFingerprint))
			So(err, ShouldBeNil)
			So(origin, ShouldEqual, testAndroidOrigin)
		})

		Convey("rejects a wrong length", func() {
			_, err := AndroidOrigin("E5:31:72")
			So(err, ShouldNotBeNil)
		})

		Convey("rejects non-hex input", func() {
			_, err := AndroidOrigin("ZZ:31:72:44:EF:62:B2:30:13:04:EB:89:9A:02:E7:39:21:86:99:E3:BF:28:F6:D8:48:9F:4C:8B:80:F0:E1:84")
			So(err, ShouldNotBeNil)
		})
	})
}

func TestEffectiveNativeApps(t *testing.T) {
	Convey("EffectiveNativeApps", t, func() {
		ios := &config.OAuthClientNativeAppConfig{Platform: config.OAuthClientNativeAppPlatformIOS, TeamID: "ABCDE12345", BundleID: "com.example.a"}
		android := &config.OAuthClientNativeAppConfig{Platform: config.OAuthClientNativeAppPlatformAndroid, PackageName: "com.example.b", SHA256CertFingerprints: []string{testAndroidFingerprint}}
		oauthConfig := &config.OAuthConfig{
			Clients: []config.OAuthClientConfig{
				{ClientID: "a", NativeApps: []*config.OAuthClientNativeAppConfig{ios}},
				{ClientID: "no-apps"},
				{ClientID: "b", NativeApps: []*config.OAuthClientNativeAppConfig{android}},
			},
		}

		Convey("returns every client's entries in client order", func() {
			So(EffectiveNativeApps(oauthConfig), ShouldResemble, []*config.OAuthClientNativeAppConfig{ios, android})
		})

		Convey("is nil-safe", func() {
			So(EffectiveNativeApps(nil), ShouldBeNil)
		})
	})
}

func TestMakeRPOriginsAndroid(t *testing.T) {
	Convey("makeRPOrigins with x_native_apps", t, func() {
		newConfigService := func(clients ...config.OAuthClientConfig) *ConfigService {
			req := httptest.NewRequest("GET", "https://auth.example.com", nil)
			req.TLS = &tls.ConnectionState{}
			return &ConfigService{
				Request:            req,
				TranslationService: &testTranslationService{},
				OAuthConfig:        &config.OAuthConfig{Clients: clients},
			}
		}
		publicOrigin := url.URL{Scheme: "https", Host: "auth.example.com"}
		androidApp := func(packageName string) *config.OAuthClientNativeAppConfig {
			return &config.OAuthClientNativeAppConfig{
				Platform:               config.OAuthClientNativeAppPlatformAndroid,
				PackageName:            packageName,
				SHA256CertFingerprints: []string{testAndroidFingerprint},
			}
		}

		Convey("accepts the origin of each listed Android certificate and records its package names", func() {
			s := newConfigService(
				config.OAuthClientConfig{ClientID: "a", NativeApps: []*config.OAuthClientNativeAppConfig{androidApp("com.example.a")}},
				config.OAuthClientConfig{ClientID: "b", NativeApps: []*config.OAuthClientNativeAppConfig{androidApp("com.example.b")}},
			)
			origins, packageNames := s.makeRPOrigins(publicOrigin)
			So(origins, ShouldResemble, []string{"https://auth.example.com", testAndroidOrigin})
			So(packageNames[testAndroidOrigin], ShouldResemble, []string{"com.example.a", "com.example.b"})
		})

		Convey("ignores iOS entries", func() {
			s := newConfigService(config.OAuthClientConfig{
				ClientID: "a",
				NativeApps: []*config.OAuthClientNativeAppConfig{
					{Platform: config.OAuthClientNativeAppPlatformIOS, TeamID: "ABCDE12345", BundleID: "com.example.a"},
				},
			})
			origins, _ := s.makeRPOrigins(publicOrigin)
			So(origins, ShouldResemble, []string{"https://auth.example.com"})
		})
	})
}

func TestCheckAndroidPackageName(t *testing.T) {
	Convey("checkAndroidPackageName", t, func() {
		cfg := &Config{
			AndroidPackageNames: map[string][]string{
				testAndroidOrigin: {"com.example.app"},
			},
		}

		Convey("ignores a web origin", func() {
			err := checkAndroidPackageName(cfg, "https://auth.example.com", []byte(`{"androidPackageName":"com.evil.app"}`))
			So(err, ShouldBeNil)
		})

		Convey("accepts a listed package name", func() {
			err := checkAndroidPackageName(cfg, testAndroidOrigin, []byte(`{"origin":"x","androidPackageName":"com.example.app"}`))
			So(err, ShouldBeNil)
		})

		Convey("rejects a package name not listed with the certificate", func() {
			err := checkAndroidPackageName(cfg, testAndroidOrigin, []byte(`{"origin":"x","androidPackageName":"com.evil.app"}`))
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, "androidPackageName is not listed")
		})

		Convey("accepts a ceremony that does not report a package name", func() {
			err := checkAndroidPackageName(cfg, testAndroidOrigin, []byte(`{"origin":"x"}`))
			So(err, ShouldBeNil)
		})

		Convey("rejects invalid clientDataJSON", func() {
			err := checkAndroidPackageName(cfg, testAndroidOrigin, []byte(`not json`))
			So(err, ShouldNotBeNil)
		})
	})
}
