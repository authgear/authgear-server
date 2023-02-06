package webapp

import (
	"fmt"
)

func WebsocketChannelName(appID string, id string) string {
	return fmt.Sprintf("app:%s:webapp-session-ws:%s", appID, id)
}

type WebsocketMessageKind string

const (
	// WebsocketMessageKindRefresh means when the client receives this message, they should refresh the page.
	WebsocketMessageKindRefresh = "refresh"
)

type WebsocketMessage struct {
	Kind WebsocketMessageKind `json:"kind"`
}
