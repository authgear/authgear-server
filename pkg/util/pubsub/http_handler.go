package pubsub

import (
	"context"
	"net/http"
	"time"

	"github.com/coder/websocket"
	"github.com/redis/go-redis/v9"

	"github.com/authgear/authgear-server/pkg/util/log"
)

type RedisHub interface {
	Subscribe(channelName string) (chan *redis.Message, func())
}

type WebsocketOutgoingMessage struct {
	MessageType websocket.MessageType
	Data        []byte
}

type Delegate interface {
	Accept(r *http.Request) (channelName string, err error)
	OnRedisSubscribe(r *http.Request) error
}

type WebsocketOriginMatcher interface {
	MatchOrigin(origin string) bool
}

const (
	WebsocketReadTimeout  = 30 * time.Second
	WebsocketWriteTimeout = 10 * time.Second
)

// HTTPHandler receives incoming websocket messages and delegates them to the delegate.
// For each websocket connection, a redis pubsub connection is established.
// Every message from the redis pubsub connection is forwarded to the websocket connection verbatim.
type HTTPHandler struct {
	RedisHub      RedisHub
	Delegate      Delegate
	LoggerFactory *log.Factory
	OriginMatcher WebsocketOriginMatcher
}

//nolint:gocognit
func (h *HTTPHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	logger := h.LoggerFactory.New("pubsub-http-handler")

	rootCtx, cancel := context.WithCancel(r.Context())
	doneChan := make(chan struct{}, 2)
	errChan := make(chan error, 2)

	defer func() {
		doneChan <- struct{}{}
		doneChan <- struct{}{}
		cancel()
		logger.Debug("canceled root context")
	}()

	insecureSkipVerify := false
	// If OriginMatcher is non-nil, check CORS here.
	// Otherwise we let the library to do CORS checking for us.
	// By default, the library does not allow CORS.
	if h.OriginMatcher != nil {

		// Non-browser user agent does not send origin header.
		if origin := r.Header.Get("Origin"); origin != "" {
			matched := h.OriginMatcher.MatchOrigin(origin)
			if !matched {
				http.Error(w, "origin is not allowed", http.StatusForbidden)
				return
			}
			// origin is matched, we can tell the library to skip verify.
			insecureSkipVerify = true
		}
	}

	wsConn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: insecureSkipVerify,
	})
	if err != nil {
		logger.WithError(err).Debug("failed to accept websocket connection")
		return
	}

	channelName, err := h.Delegate.Accept(r)
	if err != nil {
		logger.WithError(err).Debug("reject websocket connection")
		wsConn.Close(websocket.StatusNormalClosure, "connection rejected")
		return
	}

	defer wsConn.Close(websocket.StatusInternalError, "connection closed")

	logger = logger.WithField("channel", channelName)

	go func() {
		msgChan, cancel := h.RedisHub.Subscribe(channelName)

		err := h.Delegate.OnRedisSubscribe(r)
		if err != nil {
			logger.WithError(err).Debug("failed to call on redis subscribe")
			return
		}

		defer func() {
			cancel()
			logger.Debug("redis goroutine is tearing down")
		}()

		for {
			select {
			case <-doneChan:
				return
			case n, ok := <-msgChan:
				if !ok {
					return
				}

				writeCtx, cancel := context.WithTimeout(rootCtx, WebsocketWriteTimeout)
				defer cancel()
				err := wsConn.Write(writeCtx, websocket.MessageText, []byte(n.Payload))
				if err != nil {
					errChan <- err
					return
				}
			}
		}
	}()

	go func() {
		defer func() {
			logger.Debug("websocket goroutine is tearing down")
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
	logger.WithError(err).Debug("closing websocket connection due to error")
}
