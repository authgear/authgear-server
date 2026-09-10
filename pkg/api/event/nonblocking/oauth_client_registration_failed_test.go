package nonblocking_test

import (
	"encoding/json"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
)

func TestOAuthClientRegistrationFailedEventPayload(t *testing.T) {
	Convey("OAuthClientRegistrationFailedEventPayload", t, func() {
		payload := &nonblocking.OAuthClientRegistrationFailedEventPayload{
			Reason:  nonblocking.OAuthClientRegistrationReasonInvalidInitialAccessToken,
			Message: "unknown",
		}

		So(payload.NonBlockingEventType(), ShouldEqual, nonblocking.OAuthClientRegistrationFailed)
		So(payload.UserID(), ShouldEqual, "")
		So(payload.ForAudit(), ShouldBeTrue)
		So(payload.ForHook(), ShouldBeFalse)
		So(payload.GetTriggeredBy(), ShouldEqual, event.TriggeredByTypeUser)

		Convey("FillContext sets nothing -- no client was created, so none exists to name", func() {
			ctx := &event.Context{}
			payload.FillContext(ctx)
			So(ctx.ClientID, ShouldEqual, "")
		})

		// event.md documents an absent key as "no body was parsed", which
		// only holds if an empty request still serializes to a key.
		Convey("an absent request is distinguishable from an empty one", func() {
			b, err := json.Marshal(payload)
			So(err, ShouldBeNil)
			So(string(b), ShouldNotContainSubstring, `"request"`)

			payload.Request = &nonblocking.OAuthClientRegistrationFailedEventPayloadRequest{}
			b, err = json.Marshal(payload)
			So(err, ShouldBeNil)
			So(string(b), ShouldContainSubstring, `"request":{}`)
		})

		Convey("the request is recorded as sent, including the ignored auth method", func() {
			payload.Request = &nonblocking.OAuthClientRegistrationFailedEventPayloadRequest{
				GrantTypes:              []string{"client_credentials"},
				TokenEndpointAuthMethod: "client_secret_post",
			}
			b, err := json.Marshal(payload)
			So(err, ShouldBeNil)
			So(string(b), ShouldContainSubstring, `"grant_types":["client_credentials"]`)
			So(string(b), ShouldContainSubstring, `"token_endpoint_auth_method":"client_secret_post"`)
			// Unset fields stay out rather than appear as empty values.
			So(string(b), ShouldNotContainSubstring, `"redirect_uris"`)
			So(string(b), ShouldNotContainSubstring, `"client_name"`)
		})
	})
}
