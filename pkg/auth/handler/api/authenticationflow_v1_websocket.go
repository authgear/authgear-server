package api

import (
	"net/http"

	"github.com/iawaknahc/originmatcher"

	authflow "github.com/authgear/authgear-server/pkg/lib/authenticationflow"
	"github.com/authgear/authgear-server/pkg/lib/infra/redis/appredis"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/pubsub"
)

func ConfigureAuthenticationFlowV1WebsocketRoute(route httproute.Route) httproute.Route {
	return route.WithMethods("OPTIONS", "GET").WithPathPattern(authflow.WebsocketEndpointV1)
}

type AuthenticationFlowV1WebsocketEventStore interface {
	ChannelName(websocketChannelName string) string
}

type AuthenticationFlowV1WebsocketOriginMatcher interface {
	PrepareOriginMatcher(r *http.Request) (*originmatcher.T, error)
}

type AuthenticationFlowV1WebsocketHandler struct {
	RedisHandle   *appredis.Handle
	OriginMatcher AuthenticationFlowV1WebsocketOriginMatcher
	Events        AuthenticationFlowV1WebsocketEventStore
}

func (h *AuthenticationFlowV1WebsocketHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	matcher, err := h.OriginMatcher.PrepareOriginMatcher(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	handler := &pubsub.HTTPHandler{
		RedisHub:      h.RedisHandle,
		Delegate:      h,
		OriginMatcher: matcher,
	}

	handler.ServeHTTP(w, r)
}

func (h *AuthenticationFlowV1WebsocketHandler) Accept(r *http.Request) (string, error) {
	websocketChannelName := authflow.WebsocketChannelName(r)
	return h.Events.ChannelName(websocketChannelName), nil
}

func (h *AuthenticationFlowV1WebsocketHandler) OnRedisSubscribe(r *http.Request) error {
	return nil
}
