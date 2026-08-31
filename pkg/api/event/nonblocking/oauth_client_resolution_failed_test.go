package nonblocking_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
)

func TestOAuthClientResolutionFailedEventPayload(t *testing.T) {
	Convey("OAuthClientResolutionFailedEventPayload", t, func() {
		payload := &nonblocking.OAuthClientResolutionFailedEventPayload{
			ClientID: "https://mcp-client.example.com/oauth/client-metadata.json",
			Outcome:  nonblocking.OAuthClientResolutionOutcomeUnavailable,
		}

		So(payload.NonBlockingEventType(), ShouldEqual, nonblocking.OAuthClientResolutionFailed)
		So(payload.UserID(), ShouldEqual, "")
		So(payload.ForAudit(), ShouldBeTrue)
		So(payload.ForHook(), ShouldBeFalse)
		So(payload.GetTriggeredBy(), ShouldEqual, event.TriggeredByTypeUser)

		Convey("FillContext sets ctx.ClientID to the payload's client_id", func() {
			ctx := &event.Context{}
			payload.FillContext(ctx)
			So(ctx.ClientID, ShouldEqual, "https://mcp-client.example.com/oauth/client-metadata.json")
		})
	})
}
