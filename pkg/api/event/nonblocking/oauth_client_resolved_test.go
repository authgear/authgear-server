package nonblocking_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/api/event"
	"github.com/authgear/authgear-server/pkg/api/event/nonblocking"
	"github.com/authgear/authgear-server/pkg/api/model"
)

func TestOAuthClientResolvedEventPayload(t *testing.T) {
	Convey("OAuthClientResolvedEventPayload", t, func() {
		payload := &nonblocking.OAuthClientResolvedEventPayload{
			Client: nonblocking.OAuthClientResolvedEventPayloadClient{
				ClientID: "https://mcp-client.example.com/oauth/client-metadata.json",
				Source:   model.OAuthClientSourceCIMD,
			},
			Created: true,
		}

		So(payload.NonBlockingEventType(), ShouldEqual, nonblocking.OAuthClientResolved)
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
