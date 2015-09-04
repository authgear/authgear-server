package handler

import (
	"github.com/oursky/ourd/pubsub"
	"github.com/oursky/ourd/router"
)

func NewPubSubHandler(ws *pubsub.WsPubSub) router.Handler {
	return func(payload *router.Payload, response *router.Response) {
		ws.Handle(response, payload.Req)
	}
}
