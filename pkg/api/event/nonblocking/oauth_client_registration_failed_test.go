package nonblocking_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
)

func TestOAuthClientRegistrationFailedEventPayload(t *testing.T) {
	Convey("OAuthClientRegistrationFailedEventPayload", t, func() {
		payload := &nonblocking.OAuthClientRegistrationFailedEventPayload{
			Outcome: nonblocking.OAuthClientRegistrationOutcomeInvalidInitialAccessToken,
			Reason:  "unknown",
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
	})
}
