package authenticationflow

import (
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWebsocket(t *testing.T) {
	Convey("Websocket", t, func() {
		origin := "http://localhost"
		channel := "mychannel"

		websocketURL, err := WebsocketURL(origin, channel)
		So(err, ShouldBeNil)
		So(websocketURL, ShouldEqual, "ws://localhost/api/v1/authentication_flows/ws?channel=mychannel")

		r, _ := http.NewRequest("GET", websocketURL, nil)
		channelName := WebsocketChannelName(r)
		So(channelName, ShouldEqual, channel)
	})
}
