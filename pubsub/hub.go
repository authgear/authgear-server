package pubsub

import (
	log "github.com/Sirupsen/logrus"
)

// Parcel is the protocol that Hub talk with
type Parcel struct {
	Channel    string
	Data       []byte
	Connection *connection
}

// Hub is the struct that hold the subscription and do the broadcast logic
type Hub struct {
	Subscribe    chan Parcel
	Unsubscribe  chan Parcel
	Broadcast    chan Parcel
	stop         chan int
	subscription map[string][]*connection
	channels     map[string]chan []byte
}

// NewHub is factory for Hub
func NewHub() *Hub {
	return &Hub{
		Subscribe:    make(chan Parcel),
		Unsubscribe:  make(chan Parcel),
		Broadcast:    make(chan Parcel),
		subscription: map[string][]*connection{},
		channels:     map[string]chan []byte{},
	}
}

func (h *Hub) run() {
	log.Debugf("Hub runing %p", h)
	for {
		select {
		case p := <-h.Subscribe:
			h.subscribe(p.Channel, p.Connection)
		case p := <-h.Unsubscribe:
			h.unsubscribe(p.Channel, p.Connection)
		case p := <-h.Broadcast:
			h.publish(p.Channel, p.Data)
		case <-h.stop:
			break
		}
	}
}

func (h *Hub) subscribe(channel string, c *connection) {
	log.Debugf("subscribe %v, %p", channel, c)
	h.subscription[channel] = append(h.subscription[channel], c)
}

func (h *Hub) unsubscribe(channel string, c *connection) {
	log.Debugf("unsubscribe %v, %p", channel, c)
	newSubscription := []*connection{}
	for _, conn := range h.subscription[channel] {
		if conn != c {
			newSubscription = append(newSubscription, conn)
		}
	}
	h.subscription[channel] = newSubscription
}

func (h *Hub) publish(channel string, data []byte) {
	log.Debugf("publish %v, %s", channel, data)
	for _, c := range h.subscription[channel] {
		c.Send <- Parcel{
			Channel: channel,
			Data:    data,
		}
	}
}
