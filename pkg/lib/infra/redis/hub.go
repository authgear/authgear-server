package redis

import (
	"context"
	"log/slog"
	"sync/atomic"

	"github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/util/slogutil"
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

	PubSub     *redis.PubSub
	Subscriber map[string][]Subscriber
	// Mailbox handles 3 messages.
	// 1. When the mailbox is closed, that means the actor should die.
	// 2. PubSubMessageJoin
	// 3. PubSubMessageLeave
	Mailbox chan interface{}
}

// NewPubSub creates a running PubSub actor.
//
//nolint:gocognit
func NewPubSub(ctx context.Context, client *redis.Client, connKey string, supervisorMailbox chan interface{}) *PubSub {
	redisPubSub := client.Subscribe(ctx)
	mailbox := make(chan interface{})
	pubsub := &PubSub{
		ConnKey:           connKey,
		SupervisorMailbox: supervisorMailbox,
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
					logger := RedisHubLogger.GetLogger(ctx)
					for channelName, subscribers := range pubsub.Subscriber {
						for _, sub := range subscribers {
							logger.Debug(ctx, "closing channel ID", slog.Int64("subscriber_id", sub.SubscriberID))
							close(sub.MessageChannel)
						}
						delete(pubsub.Subscriber, channelName)
					}
					err := pubsub.PubSub.Close()
					if err != nil {
						logger.WithError(err).Error(ctx, "failed to clean up all")
					}
					// Exit this actor.
					return
				}

				switch n := n.(type) {
				case PubSubMessageJoin:
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
						logger := RedisHubLogger.GetLogger(ctx)
						logger.Debug(ctx, "subscribe because the first subscriber is joining")
						err := pubsub.PubSub.Subscribe(ctx, n.ChannelName)
						if err != nil {
							logger.WithError(err).Error(ctx, "failed to subscribe", slog.String("channel_name", n.ChannelName))
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
						logger := RedisHubLogger.GetLogger(ctx)
						logger.Debug(ctx, "closing channel ID", slog.Int64("subscriber_id", sub.SubscriberID))
						close(sub.MessageChannel)
					}

					subscribers = append(subscribers[:idx], subscribers[idx+1:]...)
					pubsub.Subscriber[n.ChannelName] = subscribers
					numRemaining := len(subscribers)

					// Unsubscribe if subscriber is the last one subscribing the channelName.
					if numRemaining == 0 {
						logger := RedisHubLogger.GetLogger(ctx)
						logger.Debug(ctx, "unsubscribe because the last subscriber is leaving")
						err := pubsub.PubSub.Unsubscribe(ctx, n.ChannelName)
						if err != nil {
							logger.WithError(err).Error(ctx, "failed to unsubscribe", slog.String("channel_name", n.ChannelName))
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
						logger := RedisHubLogger.GetLogger(ctx)
						logger.Debug(ctx, "dropped message to subscriber ID", slog.Int64("subscriber_id", s.SubscriberID))
					}
				}
			}
		}
	}()

	return pubsub
}

type HubMessageSubscribe struct {
	ConnectionOptions *ConnectionOptions
	ChannelName       string
	Result            chan Subscription
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
	SubscriberID *SubscriberID
	PubSub       map[string]*PubSub
	// Mailbox handles 3 messages.
	// This mailbox is never closed.
	Mailbox chan interface{}
	// 1. HubMessageSubscribe
	// 2. HubMessagePubSubCancel
	// 3. HubMessagePubSubDead
}

var RedisHubLogger = slogutil.NewLogger("redis-hub")

func NewHub(ctx context.Context, pool *Pool) *Hub {
	mailbox := make(chan interface{})
	h := &Hub{
		Pool:         pool,
		SubscriberID: &SubscriberID{},
		PubSub:       make(map[string]*PubSub),
		Mailbox:      mailbox,
	}

	go func() {
		for n := range h.Mailbox {
			switch n := n.(type) {
			case HubMessageSubscribe:
				connKey := n.ConnectionOptions.ConnKey()
				pubsub, ok := h.PubSub[connKey]
				if !ok {
					client := h.Pool.Client(n.ConnectionOptions)
					pubsub = NewPubSub(
						ctx,
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
	ConnectionOptions *ConnectionOptions,
	channelName string,
) Subscription {
	result := make(chan Subscription)
	h.Mailbox <- HubMessageSubscribe{
		ConnectionOptions: ConnectionOptions,
		ChannelName:       channelName,
		Result:            result,
	}
	return <-result
}
