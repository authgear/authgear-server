package pubsub

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNormalSubscription(t *testing.T) {
	Convey("Hub Single connection", t, func(c C) {
		Convey("Received broadcast message", func(c C) {
			hub := NewHub()
			go hub.run()
			conn := connection{
				Send: make(chan Parcel),
			}
			go func() {
				recv := <-conn.Send
				c.So(recv.Channel, ShouldEqual, "correct")
				c.So(recv.Data, ShouldResemble, []byte("Hello"))
				hub.stop <- 1
			}()
			hub.Subscribe <- Parcel{
				Channel:    "correct",
				Connection: &conn,
			}
			hub.Broadcast <- Parcel{
				Channel: "correct",
				Data:    []byte("Hello"),
			}
		})

		Convey("Subscribe to multiple channel broadcast message", func(c C) {
			hub := NewHub()
			go hub.run()
			conn := connection{
				Send: make(chan Parcel),
			}
			go func() {
				recv := <-conn.Send
				c.So(recv.Channel, ShouldEqual, "first")
				c.So(recv.Data, ShouldResemble, []byte("1"))
				recv2 := <-conn.Send
				c.So(recv2.Channel, ShouldEqual, "second")
				c.So(recv2.Data, ShouldEqual, []byte("2"))
				hub.stop <- 1
			}()
			hub.Subscribe <- Parcel{
				Channel:    "first",
				Connection: &conn,
			}
			hub.Subscribe <- Parcel{
				Channel:    "second",
				Connection: &conn,
			}
			hub.Broadcast <- Parcel{
				Channel: "first",
				Data:    []byte("1"),
			}
			hub.Broadcast <- Parcel{
				Channel: "second",
				Data:    []byte("2"),
			}
		})
	})
}
