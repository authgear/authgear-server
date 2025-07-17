package webapp

import (
	"fmt"
	"net/http"

	"strconv"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/pubsub"
)

func ConfigureWebsocketRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/_internals/ws")
}

type WebsocketHandler struct {
	AppID       config.AppID
	RedisHandle *appredis.Handle
	Publisher   *Publisher
}

func (h *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := &pubsub.HTTPHandler{
		RedisHub: h.RedisHandle,
		Delegate: h,
		// Let the library to do CORS checking, which by default does not allow CORS.
		// This websocket endpoint is intended for same-origin use only.
		OriginMatcher: nil,
	}

	handler.ServeHTTP(w, r)
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

func (h *WebsocketHandler) OnRedisSubscribe(r *http.Request) error {
	s := webapp.GetSession(r.Context())
	// if session not found, skip sending session update event
	// e.g. native app
	if s == nil {
		return nil
	}

	sessionUpdatedAfter := r.URL.Query().Get("session_updated_after")
	// if session_updated_after is not provided, skip sending session update event
	if sessionUpdatedAfter == "" {
		return nil
	}

	ts, err := strconv.ParseInt(sessionUpdatedAfter, 10, 64)
	if err != nil {
		// Invalid session updated after, skip sending session update event
		return apierrors.NewInvalid("invalid session_updated_after")
	}

	if ts < s.UpdatedAt.Unix() {
		msg := &WebsocketMessage{
			Kind: WebsocketMessageKindRefresh,
		}

		err := h.Publisher.Publish(r.Context(), s, msg)
		if err != nil {
			return err
		}
	}

	return nil
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
