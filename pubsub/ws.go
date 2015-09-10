package pubsub

import (
	"encoding/json"
	"net/http"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/websocket"
)

type connection struct {
	ws       *websocket.Conn
	channels []string
	Send     chan Parcel
}

type wsPayload struct {
	Action  string           `json:"action,omitempty"`
	Channel string           `json:"channel"`
	Data    *json.RawMessage `json:"data,omitempty"`
}

// WsPubSub is a websocket trsnaport of pubsub
// Protocol: {"action": "sub", "channel": "royuen"}
// {"action": "pub", "channel": "royuen", "data": {"any":"thing"}}
type WsPubSub struct {
	upgrader websocket.Upgrader
	hub      *Hub
}

// NewWsPubsub is factory for WsPubSub
func NewWsPubsub(hub *Hub) *WsPubSub {
	if hub == nil {
		hub = NewHub()
	}
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	ws := WsPubSub{
		upgrader,
		hub,
	}
	go hub.run()
	return &ws
}

// Handle will hijack the http responseWriter and req.
func (w *WsPubSub) Handle(writer http.ResponseWriter, req *http.Request) {
	conn, err := w.upgrader.Upgrade(writer, req, nil)
	if err != nil {
		log.Println(err)
		return
	}
	c := &connection{
		ws:   conn,
		Send: make(chan Parcel),
	}
	go w.writer(c)
	go w.reader(c)
}

func (w *WsPubSub) writer(c *connection) {
	for {
		parcel := <-c.Send
		log.Debugf("Writing ws %p, %s", c.ws, parcel.Data)
		d := json.RawMessage(parcel.Data)
		message, _ := json.Marshal(wsPayload{
			Channel: parcel.Channel,
			Data:    &d,
		})
		c.ws.WriteMessage(websocket.TextMessage, message)
	}
}

func (w *WsPubSub) reader(c *connection) {
	defer func() {
		log.Debugf("Close ws connection %p", c.ws)
		c.ws.Close()
		for _, channel := range c.channels {
			w.hub.Unsubscribe <- Parcel{
				Channel:    channel,
				Connection: c,
			}
		}
	}()
	for {
		log.Debugf("Waiting ws message %p", c.ws)
		messageType, p, err := c.ws.ReadMessage()
		if err != nil {
			log.Debugf("Ws error %v", err)
			return
		}
		log.Debugf("Received %v, %s", messageType, p)
		var payload wsPayload
		err = json.Unmarshal(p, &payload)
		if err != nil {
			log.Debugf("Can't decode Ws message %v", err)
			c.ws.WriteMessage(
				websocket.TextMessage,
				[]byte("Error: "+err.Error()+" \nClosing Connection"))
			c.ws.WriteMessage(websocket.CloseMessage, nil)
			return
		}
		switch payload.Action {
		case "sub":
			w.hub.Subscribe <- Parcel{
				Channel:    payload.Channel,
				Connection: c,
			}
			c.channels = append(c.channels, payload.Channel)
		case "pub":
			w.hub.Broadcast <- Parcel{
				Channel: payload.Channel,
				Data:    []byte(*payload.Data),
			}
		default:
			c.ws.WriteMessage(
				websocket.TextMessage,
				[]byte("Unknow action"))
		}
	}
}
