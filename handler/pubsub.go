package handler

import (
	"github.com/oursky/skygear/pubsub"
	"github.com/oursky/skygear/router"
)

type PubSubHandler struct {
	WebSocket     *pubsub.WsPubSub
	AccessKey     router.Processor `preprocessor:"accesskey"`
	preprocessors []router.Processor
}

func (h *PubSubHandler) Setup() {
	h.preprocessors = []router.Processor{
		h.AccessKey,
	}
}

func (h *PubSubHandler) GetPreprocessors() []router.Processor {
	return h.preprocessors
}

func (h *PubSubHandler) Handle(payload *router.Payload, response *router.Response) {
	h.WebSocket.Handle(response, payload.Req)
}
