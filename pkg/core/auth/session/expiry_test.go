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
			AccessToken:          "access-token",
			RefreshToken:         "refresh-token",
			AccessTokenCreatedAt: time.Date(2006, 1, 1, 0, 20, 0, 0, gotime.UTC),
		}
		config := config.APIClientConfiguration{
			Name:                 "Web App",
			Disabled:             false,
			APIKey:               "api_key",
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
