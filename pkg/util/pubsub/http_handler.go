package pubsub

import (
	"context"
	"net/http"
	"os"
	"time"

	"github.com/go-redis/redis/v8"
	"nhooyr.io/websocket"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type RedisPool interface {
	Get() *redis.Client
}

type WebsocketOutgoingMessage struct {
	MessageType websocket.MessageType
	Data        []byte
}

type Delegate interface {
	Accept(r *http.Request) (channelName string, err error)
}

const (
	WebsocketReadTimeout  = 30 * time.Second
	WebsocketWriteTimeout = 10 * time.Second
	RedisReadTimeout      = 1 * time.Second
)

// HTTPHandler receives incoming websocket messages and delegates them to the delegate.
// For each websocket connection, a redis pubsub connection is established.
// Every message from the redis pubsub connection is forwarded to the websocket connection verbatim.
type HTTPHandler struct {
	RedisPool     RedisPool
	Delegate      Delegate
	LoggerFactory *log.Factory
}

func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.LoggerFactory.New("pubsub-http-handler")

	rootCtx, cancel := context.WithCancel(context.Background())
	doneChan := make(chan struct{}, 2)
	errChan := make(chan error, 2)

	defer func() {
		doneChan <- struct{}{}
		doneChan <- struct{}{}
		cancel()
		logger.Info("canceled root context")
	}()

	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{})
	if err != nil {
		logger.WithError(err).Info("failed to accept websocket connection")
		return
	}

	channelName, err := h.Delegate.Accept(r)
	if err != nil {
		logger.WithError(err).Info("reject websocket connection")
		wsConn.Close(websocket.StatusNormalClosure, "connection rejected")
		return
	}

	defer wsConn.Close(websocket.StatusInternalError, "connection closed")

	logger = logger.WithField("channel", channelName)

	go func() {
		redisClient := h.RedisPool.Get()
		psc := redisClient.Subscribe(rootCtx, channelName)

		defer func() {
			psc.Close()
			logger.Info("redis goroutine is tearing down")
		}()

		for {
			select {
			case <-doneChan:
				return
			default:
				// Even though ReceiveTimeout takes context,
				// cancelling the context does not immediately make ReceiveTimeout return error.
				// Therefore, we have to use a small enough timeout to ensure pubsub is closed within a reasonble delay.
				iface, err := psc.ReceiveTimeout(rootCtx, RedisReadTimeout)
				if os.IsTimeout(err) {
					continue
				}
				if err != nil {
					errChan <- err
					return
				}

				switch n := iface.(type) {
				case *redis.Message:
					writeCtx, cancel := context.WithTimeout(rootCtx, WebsocketWriteTimeout)
					defer cancel()
					err := wsConn.Write(writeCtx, websocket.MessageText, []byte(n.Payload))
					if err != nil {
						errChan <- err
						return
					}
				case *redis.Subscription:
					// Nothing special to do.
					break
				case *redis.Pong:
					// Nothing special to do.
					break
				}
			}
		}
	}()

	go func() {
		defer func() {
			logger.Info("websocket goroutine is tearing down")
		}()

		for {
			select {
			case <-doneChan:
				return
			default:
				readCtx, cancel := context.WithTimeout(rootCtx, WebsocketReadTimeout)
				defer cancel()

				// Read everything from the connection and discard them.
				_, _, err := wsConn.Read(readCtx)
				if err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	err = <-errChan
	logger.WithError(err).Info("closing websocket connection due to error")
}
