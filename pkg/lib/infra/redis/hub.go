package redis

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

type SubscriberID struct {
	Counter int64
}

func (i *SubscriberID) Next() int64 {
	for {
		curr := atomic.LoadInt64(&i.Counter)
		if atomic.CompareAndSwapInt64(&i.Counter, curr, curr+1) {
			return curr
		}
	}
}

type Subscriber struct {
	ChannelName  string
	SubscriberID int64
	// The drop pattern
	// See https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html
	MessageChannel chan *redis.Message
}

// ControlMessageJoin is essentially Subscriber.
type ControlMessageJoin = Subscriber

// ControlMessageLeave is basically Subscriber without MessageChannel.
type ControlMessageLeave struct {
	ChannelName  string
	SubscriberID int64
}

// PubSub is an actor wrapping redis.PubSub
type PubSub struct {
	PubSub     *redis.PubSub
	Subscriber map[string][]Subscriber
	// Mailbox handles 3 messages.
	// 1. When the mailbox is closed, that means the actor should die.
	// 2. ControlMessageJoin
	// 3. ControlMessageLeave
	Mailbox chan interface{}
}

// Hub aims to multiplex multiple subscription to Redis PubSub channel over a single connection.
type Hub struct {
	Pool         *Pool
	SubscriberID *SubscriberID
	Logger       *log.Logger

	mutex  sync.RWMutex
	pubsub map[string]*PubSub
}

func NewHub(pool *Pool, lf *log.Factory) *Hub {
	h := &Hub{
		Pool:         pool,
		SubscriberID: &SubscriberID{},
		Logger:       lf.New("redis-hub"),
		pubsub:       make(map[string]*PubSub),
	}
	return h
}

func (h *Hub) Subscribe(
	cfg *config.RedisConfig,
	credentials *config.RedisCredentials,
	channelName string,
) (chan *redis.Message, func()) {
	pubsub := h.getPubSub(cfg, credentials)

	id := h.SubscriberID.Next()
	// The drop pattern
	// See https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html
	c := make(chan *redis.Message, 10)

	pubsub.Mailbox <- ControlMessageJoin{
		ChannelName:    channelName,
		SubscriberID:   id,
		MessageChannel: c,
	}

	return c, func() {
		pubsub.Mailbox <- ControlMessageLeave{
			SubscriberID: id,
			ChannelName:  channelName,
		}
	}
}

func (h *Hub) getPubSub(cfg *config.RedisConfig, credentials *config.RedisCredentials) *PubSub {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	connKey := credentials.ConnKey()
	pubsub, ok := h.pubsub[connKey]
	if ok {
		return pubsub
	}

	ctx := context.Background()
	client := h.Pool.Client(cfg, credentials)
	redisPubSub := client.Subscribe(ctx)
	mailbox := make(chan interface{})
	pubsub = &PubSub{
		PubSub:     redisPubSub,
		Subscriber: make(map[string][]Subscriber),
		Mailbox:    mailbox,
	}
	h.pubsub[connKey] = pubsub

	// This goroutine is the actor code of PubSub.
	go func() {
		c := pubsub.PubSub.Channel()
		for {
			select {
			case n, ok := <-pubsub.Mailbox:
				// The mailbox is closed. This actor is dying.
				if !ok {
					h.mutex.Lock()
					defer h.mutex.Unlock()

					// Close all subscriber channels
					for channelName, subscribers := range pubsub.Subscriber {
						for _, sub := range subscribers {
							h.Logger.Infof("closing channel ID: %v", sub.SubscriberID)
							close(sub.MessageChannel)
						}
						delete(pubsub.Subscriber, channelName)
					}

					delete(h.pubsub, connKey)

					h.Logger.Infof("closing pubsub: %v", connKey)
					err := pubsub.PubSub.Close()
					if err != nil {
						h.Logger.WithError(err).Errorf("failed to clean up all")
					}

					// Exit this actor.
					return
				}

				switch n := n.(type) {
				case ControlMessageJoin:
					pubsub.Subscriber[n.ChannelName] = append(
						pubsub.Subscriber[n.ChannelName],
						Subscriber{
							ChannelName:    n.ChannelName,
							SubscriberID:   n.SubscriberID,
							MessageChannel: n.MessageChannel,
						},
					)

					// Subscribe if it is the first subscriber.
					if len(pubsub.Subscriber[n.ChannelName]) == 1 {
						h.Logger.Infof("subscribe because the first subscriber is joining")
						ctx := context.Background()
						err := pubsub.PubSub.Subscribe(ctx, n.ChannelName)
						if err != nil {
							h.Logger.WithError(err).Errorf("failed to subscribe: %v", n.ChannelName)
							close(pubsub.Mailbox)
							break
						}
					}
				case ControlMessageLeave:
					numRemaining := -1
					// Close the specific subscriber channel
					subscribers := pubsub.Subscriber[n.ChannelName]
					idx := -1
					for i, sub := range subscribers {
						if sub.SubscriberID == n.SubscriberID {
							idx = i
							break
						}
					}
					if idx != -1 {
						sub := subscribers[idx]
						h.Logger.Infof("closing channel ID: %v", sub.SubscriberID)
						close(sub.MessageChannel)
					}

					subscribers = append(subscribers[:idx], subscribers[idx+1:]...)
					pubsub.Subscriber[n.ChannelName] = subscribers
					numRemaining = len(subscribers)

					// Unsubscribe if subscriber is the last one subscribing the channelName.
					if numRemaining == 0 {
						h.Logger.Infof("unsubscribe because the last subscriber is leaving")
						ctx := context.Background()
						err := pubsub.PubSub.Unsubscribe(ctx, n.ChannelName)
						if err != nil {
							h.Logger.WithError(err).Errorf("failed to unsubscribe: %v", n.ChannelName)
						}
					}
				}
			case n, ok := <-c:
				if !ok {
					close(pubsub.Mailbox)
					break
				}

				for _, s := range pubsub.Subscriber[n.Channel] {
					// The drop pattern
					// See https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html
					select {
					case s.MessageChannel <- n:
						break
					default:
						h.Logger.Debugf("dropped message to subscriber ID: %v", s.SubscriberID)
					}
				}
			}
		}
	}()

	return pubsub
}
