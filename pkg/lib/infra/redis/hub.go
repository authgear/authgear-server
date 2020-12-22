package redis

import (
	"context"
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

type Subscription struct {
	MessageChannel chan *redis.Message
	Cancel         func()
}

// PubSubMessageJoin is essentially Subscriber.
type PubSubMessageJoin Subscriber

// PubSubMessageLeave is basically Subscriber without MessageChannel.
type PubSubMessageLeave struct {
	ChannelName  string
	SubscriberID int64
}

// PubSub is an actor wrapping redis.PubSub
type PubSub struct {
	ConnKey           string
	SupervisorMailbox chan interface{}

	Logger     *log.Logger
	PubSub     *redis.PubSub
	Subscriber map[string][]Subscriber
	// Mailbox handles 3 messages.
	// 1. When the mailbox is closed, that means the actor should die.
	// 2. PubSubMessageJoin
	// 3. PubSubMessageLeave
	Mailbox chan interface{}
}

// NewPubSub creates a running PubSub actor.
func NewPubSub(logger *log.Logger, client *redis.Client, connKey string, supervisorMailbox chan interface{}) *PubSub {
	ctx := context.Background()
	redisPubSub := client.Subscribe(ctx)
	mailbox := make(chan interface{})
	pubsub := &PubSub{
		ConnKey:           connKey,
		SupervisorMailbox: supervisorMailbox,
		Logger:            logger,
		PubSub:            redisPubSub,
		Subscriber:        make(map[string][]Subscriber),
		Mailbox:           mailbox,
	}

	go func() {
		c := pubsub.PubSub.Channel()
		for {
			select {
			case n, ok := <-pubsub.Mailbox:
				// The mailbox is closed. This actor is dying.
				if !ok {
					// Close all subscriber channels
					for channelName, subscribers := range pubsub.Subscriber {
						for _, sub := range subscribers {
							pubsub.Logger.Infof("closing channel ID: %v", sub.SubscriberID)
							close(sub.MessageChannel)
						}
						delete(pubsub.Subscriber, channelName)
					}
					err := pubsub.PubSub.Close()
					if err != nil {
						pubsub.Logger.WithError(err).Errorf("failed to clean up all")
					}
					// Exit this actor.
					return
				}

				switch n := n.(type) {
				case PubSubMessageJoin:
					pubsub.Subscriber[n.ChannelName] = append(
						pubsub.Subscriber[n.ChannelName],
						// nolint: gosimple
						Subscriber{
							ChannelName:    n.ChannelName,
							SubscriberID:   n.SubscriberID,
							MessageChannel: n.MessageChannel,
						},
					)

					// Subscribe if it is the first subscriber.
					if len(pubsub.Subscriber[n.ChannelName]) == 1 {
						pubsub.Logger.Infof("subscribe because the first subscriber is joining")
						ctx := context.Background()
						err := pubsub.PubSub.Subscribe(ctx, n.ChannelName)
						if err != nil {
							pubsub.Logger.WithError(err).Errorf("failed to subscribe: %v", n.ChannelName)
							pubsub.SupervisorMailbox <- HubMessagePubSubDead{
								ConnKey: pubsub.ConnKey,
							}
							break
						}
					}
				case PubSubMessageLeave:
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
						pubsub.Logger.Infof("closing channel ID: %v", sub.SubscriberID)
						close(sub.MessageChannel)
					}

					subscribers = append(subscribers[:idx], subscribers[idx+1:]...)
					pubsub.Subscriber[n.ChannelName] = subscribers
					numRemaining := len(subscribers)

					// Unsubscribe if subscriber is the last one subscribing the channelName.
					if numRemaining == 0 {
						pubsub.Logger.Infof("unsubscribe because the last subscriber is leaving")
						ctx := context.Background()
						err := pubsub.PubSub.Unsubscribe(ctx, n.ChannelName)
						if err != nil {
							pubsub.Logger.WithError(err).Errorf("failed to unsubscribe: %v", n.ChannelName)
						}
					}
				}
			case n, ok := <-c:
				if !ok {
					pubsub.SupervisorMailbox <- HubMessagePubSubDead{
						ConnKey: pubsub.ConnKey,
					}
					break
				}

				for _, s := range pubsub.Subscriber[n.Channel] {
					// The drop pattern
					// See https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html
					select {
					case s.MessageChannel <- n:
						break
					default:
						pubsub.Logger.Debugf("dropped message to subscriber ID: %v", s.SubscriberID)
					}
				}
			}
		}
	}()

	return pubsub
}

type HubMessageSubscribe struct {
	Config      *config.RedisConfig
	Credentials *config.RedisCredentials
	ChannelName string
	Result      chan Subscription
}

type HubMessagePubSubDead struct {
	ConnKey string
}

type HubMessagePubSubCancel struct {
	ConnKey      string
	SubscriberID int64
	ChannelName  string
}

// Hub aims to multiplex multiple subscription to Redis PubSub channel over a single connection.
type Hub struct {
	Pool         *Pool
	Logger       *log.Logger
	SubscriberID *SubscriberID
	PubSub       map[string]*PubSub
	// Mailbox handles 3 messages.
	// This mailbox is never closed.
	Mailbox chan interface{}
	// 1. HubMessageSubscribe
	// 2. HubMessagePubSubCancel
	// 3. HubMessagePubSubDead
}

func NewHub(pool *Pool, lf *log.Factory) *Hub {
	mailbox := make(chan interface{})
	h := &Hub{
		Pool:         pool,
		SubscriberID: &SubscriberID{},
		Logger:       lf.New("redis-hub"),
		PubSub:       make(map[string]*PubSub),
		Mailbox:      mailbox,
	}

	go func() {
		for n := range h.Mailbox {
			switch n := n.(type) {
			case HubMessageSubscribe:
				connKey := n.Credentials.ConnKey()
				pubsub, ok := h.PubSub[connKey]
				if !ok {
					client := h.Pool.Client(n.Config, n.Credentials)
					pubsub = NewPubSub(
						h.Logger,
						client,
						connKey,
						mailbox,
					)
					h.PubSub[connKey] = pubsub
				}

				id := h.SubscriberID.Next()
				// The drop pattern
				// See https://www.ardanlabs.com/blog/2017/10/the-behavior-of-channels.html
				c := make(chan *redis.Message, 10)
				pubsub.Mailbox <- PubSubMessageJoin{
					ChannelName:    n.ChannelName,
					SubscriberID:   id,
					MessageChannel: c,
				}

				n.Result <- Subscription{
					MessageChannel: c,
					Cancel: func() {
						h.Mailbox <- HubMessagePubSubCancel{
							ConnKey:      connKey,
							SubscriberID: id,
							ChannelName:  n.ChannelName,
						}
					},
				}
			case HubMessagePubSubCancel:
				pubsub, ok := h.PubSub[n.ConnKey]
				if !ok {
					break
				}
				pubsub.Mailbox <- PubSubMessageLeave{
					SubscriberID: n.SubscriberID,
					ChannelName:  n.ChannelName,
				}
			case HubMessagePubSubDead:
				pubsub, ok := h.PubSub[n.ConnKey]
				if !ok {
					break
				}
				close(pubsub.Mailbox)
				delete(h.PubSub, n.ConnKey)
			}
		}
	}()

	return h
}

func (h *Hub) Subscribe(
	cfg *config.RedisConfig,
	credentials *config.RedisCredentials,
	channelName string,
) Subscription {
	result := make(chan Subscription)
	h.Mailbox <- HubMessageSubscribe{
		Config:      cfg,
		Credentials: credentials,
		ChannelName: channelName,
		Result:      result,
	}
	return <-result
}
