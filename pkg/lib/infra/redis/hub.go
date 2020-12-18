package redis

import (
	"context"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/log"
)

const (
	RedisReadTimeout = 1 * time.Second
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

type ChannelKey struct {
	ConnKey     string
	ChannelName string
}

type Subscriber struct {
	SubscriberID int64
	Chan         chan *redis.Message
}

// Hub aims to multiplex multiple subscription to Redis PubSub channel over a single connection.
type Hub struct {
	Pool         *Pool
	SubscriberID *SubscriberID
	Logger       *log.Logger

	mutex      sync.RWMutex
	pubsub     map[string]*redis.PubSub
	subscriber map[ChannelKey][]Subscriber
}

func NewHub(pool *Pool, lf *log.Factory) *Hub {
	h := &Hub{
		Pool:         pool,
		SubscriberID: &SubscriberID{},
		Logger:       lf.New("redis-hub"),

		subscriber: make(map[ChannelKey][]Subscriber),
		pubsub:     make(map[string]*redis.PubSub),
	}
	return h
}

func (h *Hub) Subscribe(
	cfg *config.RedisConfig,
	credentials *config.RedisCredentials,
	channelName string,
) (chan *redis.Message, func(), error) {
	connKey := credentials.ConnKey()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	pubsub := h.getPubSub(cfg, credentials)
	id := h.SubscriberID.Next()

	key := ChannelKey{
		ConnKey:     connKey,
		ChannelName: channelName,
	}

	c := make(chan *redis.Message)

	h.subscriber[key] = append(h.subscriber[key], Subscriber{
		SubscriberID: id,
		Chan:         c,
	})

	// Subscribe if it is the first subscriber.
	if len(h.subscriber[key]) == 1 {
		h.Logger.Infof("subscribe because the first subscriber is joining")
		ctx := context.Background()
		err := pubsub.Subscribe(ctx, channelName)
		if err != nil {
			return nil, nil, err
		}
	}

	return c, func() {
		h.cleanupOne(cfg, credentials, channelName, id)
	}, nil
}

func (h *Hub) getPubSub(cfg *config.RedisConfig, credentials *config.RedisCredentials) *redis.PubSub {
	connKey := credentials.ConnKey()

	pubsub, ok := h.pubsub[connKey]
	if ok {
		return pubsub
	}

	ctx := context.Background()
	client := h.Pool.Client(cfg, credentials)
	pubsub = client.Subscribe(ctx)
	h.pubsub[connKey] = pubsub

	go func() {
		for {
			iface, err := pubsub.ReceiveTimeout(ctx, RedisReadTimeout)
			if os.IsTimeout(err) {
				continue
			}
			if err != nil {
				h.cleanupAll(cfg, credentials)
				return
			}

			switch n := iface.(type) {
			case *redis.Subscription:
				// Nothing special to do.
				break
			case *redis.Pong:
				// Nothing special to do.
				break
			case *redis.Message:
				key := ChannelKey{
					ConnKey:     connKey,
					ChannelName: n.Channel,
				}

				h.mutex.RLock()
				subs := make([]Subscriber, len(h.subscriber[key]))
				copy(subs, h.subscriber[key])
				h.mutex.RUnlock()

				for _, s := range subs {
					s.Chan <- n
				}
			}
		}
	}()

	return pubsub
}

func (h *Hub) cleanupAll(
	cfg *config.RedisConfig,
	credentials *config.RedisCredentials,
) {
	connKey := credentials.ConnKey()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Close all subscriber channels
	for key, subscribers := range h.subscriber {
		if key.ConnKey != connKey {
			continue
		}
		for _, sub := range subscribers {
			h.Logger.Infof("closing channel ID: %v", sub.SubscriberID)
			close(sub.Chan)
		}
		delete(h.subscriber, key)
	}

	p, ok := h.pubsub[connKey]
	if !ok {
		return
	}
	defer func() {
		delete(h.pubsub, connKey)
	}()

	h.Logger.Infof("closing pubsub: %v", connKey)
	err := p.Close()
	if err != nil {
		h.Logger.WithError(err).Errorf("failed to clean up all")
	}
}

func (h *Hub) cleanupOne(
	cfg *config.RedisConfig,
	credentials *config.RedisCredentials,
	channelName string,
	id int64,
) {
	connKey := credentials.ConnKey()

	h.mutex.Lock()
	defer h.mutex.Unlock()

	numRemaining := -1
	// Close the specific subscriber channel
	for key, subscribers := range h.subscriber {
		if key.ConnKey != connKey || key.ChannelName != channelName {
			continue
		}

		idx := -1
		for i, sub := range subscribers {
			if sub.SubscriberID == id {
				idx = i
				break
			}
		}
		if idx != -1 {
			sub := subscribers[idx]
			h.Logger.Infof("closing channel ID: %v", sub.SubscriberID)
			close(sub.Chan)
		}

		subscribers = append(subscribers[:idx], subscribers[idx+1:]...)
		h.subscriber[key] = subscribers
		numRemaining = len(subscribers)
	}

	// Unsubscribe if subscriber is the last one subscribing the channelName.
	if numRemaining == 0 {
		h.Logger.Infof("unsubscribe because the last subscriber is leaving")
		p, ok := h.pubsub[connKey]
		if !ok {
			return
		}

		ctx := context.Background()
		err := p.Unsubscribe(ctx, channelName)
		if err != nil {
			h.Logger.WithError(err).Errorf("failed to unsubscribe: %v", channelName)
		}
	}
}
