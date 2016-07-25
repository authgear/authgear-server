// Copyright 2015-present Oursky Ltd.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pubsub

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/websocket"
)

type connection struct {
	ws       *websocket.Conn
	channels []string
	Send     chan Parcel
	done     chan bool
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
		CheckOrigin: func(r *http.Request) bool {
			// allow all connections
			return true
		},
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
		done: make(chan bool),
	}
	go w.writer(c)
	go w.reader(c)
}

func (w *WsPubSub) writer(c *connection) {
writer:
	for {
		select {
		case parcel := <-c.Send:
			log.Debugf("Writing ws %p, %s, %s", c.ws, parcel.Channel, parcel.Data)
			d := json.RawMessage(parcel.Data)
			message, _ := json.Marshal(wsPayload{
				Channel: parcel.Channel,
				Data:    &d,
			})
			c.ws.WriteMessage(websocket.TextMessage, message)
		case <-c.done:
			break writer
		}
	}
	log.Debugf("Close ws writer goroutine %p", c.ws)
}

func (w *WsPubSub) reader(c *connection) {
	defer func() {
		log.Debugf("Close ws reader connection %p", c.ws)
		c.ws.Close()
		for _, channel := range c.channels {
			w.hub.Unsubscribe <- Parcel{
				Channel:    channel,
				Connection: c,
			}
		}
		c.done <- true
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
		case "unsub":
			w.hub.Unsubscribe <- Parcel{
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
