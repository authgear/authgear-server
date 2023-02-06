package webapp

import (
	"net/http"
	"strconv"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/auth/webapp"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/pubsub"
)

func ConfigureWebsocketRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/_internals/ws")
}

type WebsocketHandler struct {
	AppID         config.AppID
	LoggerFactory *log.Factory
	RedisHandle   *appredis.Handle
	Publisher     *webapp.Publisher
}

func (h *WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	handler := &pubsub.HTTPHandler{
		RedisHub:      h.RedisHandle,
		Delegate:      h,
		LoggerFactory: h.LoggerFactory,
	}

	handler.ServeHTTP(w, r)
}

func (h *WebsocketHandler) Accept(r *http.Request) (channelName string, err error) {
	s := webapp.GetSession(r.Context())
	if s == nil {
		err = webapp.ErrSessionNotFound
		return
	}

	channelName = webapp.WebsocketChannelName(string(h.AppID), s.ID)
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
		msg := &webapp.WebsocketMessage{
			Kind: webapp.WebsocketMessageKindRefresh,
		}

		err := h.Publisher.Publish(s, msg)
		if err != nil {
			return err
		}
	}

	return nil
}
