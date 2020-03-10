package session

import (
	"testing"
	"time"
	gotime "time"

	"github.com/skygeario/skygear-server/pkg/core/auth"
	"github.com/skygeario/skygear-server/pkg/core/config"
	. "github.com/smartystreets/goconvey/convey"
)

func TestComputeSessionExpiry(t *testing.T) {
	Convey("computeSessionExpiry", t, func() {
		session := &auth.Session{
			ID:                   "session-id",
			ClientID:             "web-app",
			UserID:               "user-id",
			PrincipalID:          "principal-id",
			CreatedAt:            time.Date(2006, 1, 1, 0, 0, 0, 0, gotime.UTC),
			AccessedAt:           time.Date(2006, 1, 1, 0, 25, 0, 0, gotime.UTC),
			AccessTokenHash:      "access-token-hash",
			RefreshTokenHash:     "refresh-token-hash",
			AccessTokenCreatedAt: time.Date(2006, 1, 1, 0, 20, 0, 0, gotime.UTC),
		}
		config := config.APIClientConfiguration{
			ClientName:           "Web App",
			ClientID:             "client_id",
			SessionTransport:     config.SessionTransportTypeHeader,
			AccessTokenLifetime:  1800,
			SessionIdleTimeout:   300,
			RefreshTokenLifetime: 86400,
		}

		Convey("should be correct when refresh token is disabled", func() {
			config.RefreshTokenDisabled = true

			Convey("and idle timeout is disabled", func() {
				config.SessionIdleTimeoutEnabled = false
				expiry := computeSessionStorageExpiry(session, config)
				So(expiry, ShouldResemble, time.Date(2006, 1, 1, 0, 50, 0, 0, gotime.UTC))
			})

			Convey("and idle timeout is enabled", func() {
				config.SessionIdleTimeoutEnabled = true
				expiry := computeSessionStorageExpiry(session, config)
				So(expiry, ShouldResemble, time.Date(2006, 1, 1, 0, 30, 0, 0, gotime.UTC))
			})
		})

		Convey("should be correct when refresh token is enabled", func() {
			config.RefreshTokenDisabled = false

			Convey("and idle timeout is disabled", func() {
				config.SessionIdleTimeoutEnabled = false
				expiry := computeSessionStorageExpiry(session, config)
				So(expiry, ShouldResemble, time.Date(2006, 1, 2, 0, 0, 0, 0, gotime.UTC))
			})

			Convey("and idle timeout is enabled", func() {
				config.SessionIdleTimeoutEnabled = true
				expiry := computeSessionStorageExpiry(session, config)
				So(expiry, ShouldResemble, time.Date(2006, 1, 1, 0, 25, 0, 0, gotime.UTC))
			})
		})
	})
}

func TestCheckSessionExpired(t *testing.T) {
	Convey("checkSessionExpired", t, func() {
		session := &auth.Session{
			ID:                   "session-id",
			ClientID:             "web-app",
			UserID:               "user-id",
			PrincipalID:          "principal-id",
			CreatedAt:            time.Date(2006, 1, 1, 0, 0, 0, 0, gotime.UTC),
			AccessedAt:           time.Date(2006, 1, 1, 0, 25, 0, 0, gotime.UTC),
			AccessTokenHash:      "access-token-hash",
			RefreshTokenHash:     "refresh-token-hash",
			AccessTokenCreatedAt: time.Date(2006, 1, 1, 0, 20, 0, 0, gotime.UTC),
		}
		config := config.APIClientConfiguration{
			ClientName:           "Web App",
			ClientID:             "client_id",
			SessionTransport:     config.SessionTransportTypeHeader,
			AccessTokenLifetime:  1800,
			SessionIdleTimeout:   600,
			RefreshTokenLifetime: 86400,
		}

		doCheckSessionExpired := func(mins int, kind auth.SessionTokenKind) bool {
			return checkSessionExpired(session, time.Date(2006, 1, 1, 0, mins, 0, 0, gotime.UTC), config, kind)
		}

		Convey("should treat refresh tokens as expired if disabled", func() {
			So(doCheckSessionExpired(0, auth.SessionTokenKindRefreshToken), ShouldBeFalse)
			config.RefreshTokenDisabled = true
			So(doCheckSessionExpired(0, auth.SessionTokenKindRefreshToken), ShouldBeTrue)
		})

		Convey("should check refresh token lifetime expiry", func() {
			So(doCheckSessionExpired(1440, auth.SessionTokenKindRefreshToken), ShouldBeFalse)
			So(doCheckSessionExpired(1441, auth.SessionTokenKindRefreshToken), ShouldBeTrue)
		})

		Convey("should check refresh token idle expiry", func() {
			config.SessionIdleTimeoutEnabled = true
			So(doCheckSessionExpired(30, auth.SessionTokenKindRefreshToken), ShouldBeFalse)
			So(doCheckSessionExpired(31, auth.SessionTokenKindRefreshToken), ShouldBeTrue)
		})

		Convey("should check access token expiry", func() {
			So(doCheckSessionExpired(50, auth.SessionTokenKindAccessToken), ShouldBeFalse)
			So(doCheckSessionExpired(51, auth.SessionTokenKindAccessToken), ShouldBeTrue)
		})

		Convey("should check access token idle expiry", func() {
			config.SessionIdleTimeoutEnabled = true
			So(doCheckSessionExpired(30, auth.SessionTokenKindAccessToken), ShouldBeFalse)
			So(doCheckSessionExpired(31, auth.SessionTokenKindAccessToken), ShouldBeTrue)

			config.RefreshTokenDisabled = true
			So(doCheckSessionExpired(35, auth.SessionTokenKindAccessToken), ShouldBeFalse)
			So(doCheckSessionExpired(36, auth.SessionTokenKindAccessToken), ShouldBeTrue)
		})

		Convey("should treat access token as expired if refresh token is expired", func() {
			So(doCheckSessionExpired(25, auth.SessionTokenKindAccessToken), ShouldBeFalse)
			So(doCheckSessionExpired(26, auth.SessionTokenKindAccessToken), ShouldBeFalse)
			So(doCheckSessionExpired(25, auth.SessionTokenKindRefreshToken), ShouldBeFalse)
			So(doCheckSessionExpired(26, auth.SessionTokenKindRefreshToken), ShouldBeFalse)
			config.RefreshTokenLifetime = 25 * 60
			So(doCheckSessionExpired(25, auth.SessionTokenKindAccessToken), ShouldBeFalse)
			So(doCheckSessionExpired(26, auth.SessionTokenKindAccessToken), ShouldBeTrue)
			So(doCheckSessionExpired(25, auth.SessionTokenKindRefreshToken), ShouldBeFalse)
			So(doCheckSessionExpired(26, auth.SessionTokenKindRefreshToken), ShouldBeTrue)
		})
	})
}
