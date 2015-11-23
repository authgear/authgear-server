package pubsub

import (
	. "github.com/smartystreets/goconvey/convey"
	"sync"
	"testing"
	"time"
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

		Convey("No receive message after unsubscribe", func(c C) {
			hub := NewHub()
			go hub.run()
			conn := connection{
				Send: make(chan Parcel),
			}
			hub.Subscribe <- Parcel{
				Channel:    "correct",
				Connection: &conn,
			}
			hub.Unsubscribe <- Parcel{
				Channel:    "correct",
				Connection: &conn,
			}
			hub.Broadcast <- Parcel{
				Channel: "correct",
				Data:    []byte("Hello"),
			}

			select {
			case <-conn.Send:
				t.Fatal("received pubsub message after unsubscribe.")
			case <-time.After(50 * time.Millisecond):
				// do nothing
			}
		})

		Convey("Subscribe to multiple channel broadcast message", func(c C) {
			hub := NewHub()
			go hub.run()
			conn := connection{
				Send: make(chan Parcel),
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				hub.Subscribe <- Parcel{
					Channel:    "first",
					Connection: &conn,
				}
				hub.Subscribe <- Parcel{
					Channel:    "second",
					Connection: &conn,
				}
				wg.Done()
			}()
			wg.Wait()
			hub.Broadcast <- Parcel{
				Channel: "first",
				Data:    []byte("1"),
			}
			recv := <-conn.Send
			c.So(recv.Channel, ShouldEqual, "first")
			c.So(recv.Data, ShouldResemble, []byte("1"))
			hub.Broadcast <- Parcel{
				Channel: "second",
				Data:    []byte("2"),
			}
			recv2 := <-conn.Send
			c.So(recv2.Channel, ShouldEqual, "second")
			c.So(recv2.Data, ShouldResemble, []byte("2"))
			hub.stop <- 1
		})

		Convey("Received broadcast message time out", func(c C) {
			hub := NewHub()
			hub.timeout = 0
			go hub.run()
			conn := connection{
				Send: make(chan Parcel),
			}
			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				hub.Subscribe <- Parcel{
					Channel:    "first",
					Connection: &conn,
				}
				wg.Done()
			}()
			wg.Wait()
			hub.Broadcast <- Parcel{
				Channel: "first",
				Data:    []byte("1"),
			}
			time.Sleep(1 * time.Second)
			select {
			case <-conn.Send:
				t.Fatalf("Message should be time out!")
			default:
			}
			hub.stop <- 1
		})
	})
}
