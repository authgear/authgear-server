package webapp

import (
	"fmt"
	"net/http"

	redigo "github.com/gomodule/redigo/redis"
	"nhooyr.io/websocket"

	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/pubsub"
)

func ConfigureWebsocketRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/ws")
}

type WebsocketHandler struct {
	AppID         config.AppID
	LoggerFactory *log.Factory
	RedisHandle   *redis.Handle
}

func (h *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := &pubsub.HTTPHandler{
		RedisPool:     h,
		Delegate:      h,
		LoggerFactory: h.LoggerFactory,
	}

	handler.ServeHTTP(w, r)
}

func (h *WebsocketHandler) Get() redigo.Conn {
	return h.RedisHandle.Pool().Get()
}

func (h *WebsocketHandler) Accept(r *http.Request) (channelName string, err error) {
	s := webapp.GetSession(r.Context())
	if s == nil {
		err = webapp.ErrSessionNotFound
		return
	}

	channelName = WebsocketChannelName(string(h.AppID), s.ID)
	return
}

func (h *WebsocketHandler) OnWebsocketMessage(messageType websocket.MessageType, data []byte) (*pubsub.WebsocketOutgoingMessage, error) {
	// We do not expect the client to send us anything.
	return nil, nil
}

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
