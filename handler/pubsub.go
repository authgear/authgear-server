package handler

import (
	"github.com/oursky/skygear/pubsub"
	"github.com/oursky/skygear/router"
)

func NewPubSubHandler(ws *pubsub.WsPubSub) router.Handler {
	return func(payload *router.Payload, response *router.Response) {
		ws.Handle(response, payload.Req)
	}
}
