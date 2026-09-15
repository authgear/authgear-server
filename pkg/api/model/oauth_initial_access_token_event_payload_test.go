package model_test

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestNewEventPayloadInitialAccessToken(t *testing.T) {
	Convey("NewEventPayloadInitialAccessToken", t, func() {
		Convey("nil maps to nil", func() {
			So(model.NewEventPayloadInitialAccessToken(nil), ShouldBeNil)
		})

		Convey("a real token maps to its id, type, created_at and expires_at -- never the token value", func() {
			createdAt := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
			expiresAt := time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)
			iat := &model.OAuthInitialAccessToken{
				Meta: model.Meta{
					ID:        "iat-id-123",
					CreatedAt: createdAt,
				},
				ExpiresAt: expiresAt,
				Type:      model.OAuthInitialAccessTokenTypeThirdParty,
			}
			result := model.NewEventPayloadInitialAccessToken(iat)
			So(result, ShouldNotBeNil)
			So(result.ID, ShouldEqual, "iat-id-123")
			So(result.Type, ShouldEqual, model.OAuthInitialAccessTokenTypeThirdParty)
			So(result.CreatedAt, ShouldEqual, createdAt)
			So(result.ExpiresAt, ShouldEqual, expiresAt)
		})
	})
}
