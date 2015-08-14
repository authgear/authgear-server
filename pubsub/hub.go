package pubsub

import (
	log "github.com/Sirupsen/logrus"
	"time"
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
	timeout      time.Duration
}

// NewHub is factory for Hub
func NewHub() *Hub {
	return &Hub{
		Subscribe:    make(chan Parcel),
		Unsubscribe:  make(chan Parcel),
		Broadcast:    make(chan Parcel),
		stop:         make(chan int),
		subscription: map[string][]*connection{},
		channels:     map[string]chan []byte{},
		timeout:      1,
	}
}

func (h *Hub) run() {
	log.Debugf("Hub running %p", h)
	for {
		select {
		case p := <-h.Subscribe:
			h.subscribe(p.Channel, p.Connection)
		case p := <-h.Unsubscribe:
			h.unsubscribe(p.Channel, p.Connection)
		case p := <-h.Broadcast:
			log.Warnf("Broadcast %v:%s", p.Channel, p.Data)
			h.publish(p.Channel, p.Data)
		case <-h.stop:
			break
		}
	}
	log.Info("Hub stopped %p!", h)
}

func (h *Hub) timeOut() <-chan time.Time {
	return time.After(h.timeout * time.Second)
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
	parcel := Parcel{
		Channel: channel,
		Data:    data,
	}
	for _, c := range h.subscription[channel] {
		go func() {
			select {
			case c.Send <- parcel:
				log.Debugf("Published to %p", c)
			case <-h.timeOut():
				log.Warnf("Can't publish, %p, %v:%s", c, channel, data)
			}
		}()
	}
}
