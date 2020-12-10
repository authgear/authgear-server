package pubsub

import (
	"context"
	"net/http"
	"time"

	"github.com/gomodule/redigo/redis"
	"nhooyr.io/websocket"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type RedisPool interface {
	Get() redis.Conn
}

type WebsocketOutgoingMessage struct {
	MessageType websocket.MessageType
	Data        []byte
}

type Delegate interface {
	Accept(r *http.Request) (channelName string, err error)
	OnWebsocketMessage(messageType websocket.MessageType, data []byte) (*WebsocketOutgoingMessage, error)
}

const (
	WebsocketReadTimeout  = 30 * time.Second
	WebsocketWriteTimeout = 10 * time.Second
	RedisReadTimeout      = 30 * time.Second
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
		redisConn := h.RedisPool.Get()
		psc := redis.PubSubConn{Conn: redisConn}
		defer func() {
			psc.Close()
			logger.Info("redis goroutine is tearing down")
		}()

		err := psc.Subscribe(channelName)
		if err != nil {
			errChan <- err
			return
		}

		for {
			select {
			case <-doneChan:
				return
			default:
				// Once ReceiveWithTimeout results in error once.
				// The next call to ReceiveWithTimeout immediately fails with closed connection :(
				// Therefore we have to make RedisReadTimeout the same as WebsocketReadTimeout.
				// The redis connection will have a delay of RedisReadTimeout before closing.
				switch n := psc.ReceiveWithTimeout(RedisReadTimeout).(type) {
				case error:
					errChan <- n
					return
				case redis.Message:
					writeCtx, cancel := context.WithTimeout(rootCtx, WebsocketWriteTimeout)
					defer cancel()
					err := wsConn.Write(writeCtx, websocket.MessageText, n.Data)
					if err != nil {
						errChan <- err
						return
					}
				case redis.Subscription:
					// Nothing special to do.
					break
				case redis.Pong:
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

				messageType, data, err := wsConn.Read(readCtx)
				if err != nil {
					errChan <- err
					return
				}

				outgoing, err := h.Delegate.OnWebsocketMessage(messageType, data)
				if err != nil {
					errChan <- err
					return
				}

				if outgoing != nil {
					writeCtx, cancel := context.WithTimeout(rootCtx, WebsocketWriteTimeout)
					defer cancel()

					err = wsConn.Write(writeCtx, outgoing.MessageType, outgoing.Data)
					if err != nil {
						errChan <- err
						return
					}
				}
			}
		}
	}()

	err = <-errChan
	logger.WithError(err).Info("closing websocket connection due to error")
}
