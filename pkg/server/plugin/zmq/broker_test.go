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

// +build zmq

package zmq

import (
	"sync"
	"testing"
	"time"

	"github.com/zeromq/goczmq"

	. "github.com/smartystreets/goconvey/convey"
)

func workerSock(t *testing.T, id string, addr string) *goczmq.Sock {
	sock := goczmq.NewSock(goczmq.Dealer)
	sock.SetIdentity(id)
	if err := sock.Connect(addr); err != nil {
		t.Fatalf("Failed to create worker to addr = %s", addr)
	}
	return sock
}

func clientSock(t *testing.T, id string, addr string) *goczmq.Sock {
	sock := goczmq.NewSock(goczmq.Dealer)
	sock.SetIdentity(id)
	if err := sock.Connect(addr); err != nil {
		t.Fatalf("Failed to create client to addr = %s", addr)
	}
	return sock
}

func recvNonControlFrame(w *goczmq.Sock) [][]byte {
	var msg [][]byte
	for {
		msg, _ = w.RecvMessage()
		if len(msg) != 1 {
			break
		}
	}
	return msg
}

func bytesArray(ss ...string) (bs [][]byte) {
	for _, s := range ss {
		bs = append(bs, []byte(s))
	}
	return
}

func TestWorker(t *testing.T) {
	Convey("Test workerQueue", t, func() {
		address1 := "address1"
		address2 := "address2"
		address3 := "address3"

		Convey("Add and pick a worker", func() {
			q := newWorkerQueue()
			q.Add(newWorker(address1))
			So(q.Len(), ShouldEqual, 1)
			addr := q.Next()
			So(addr, ShouldResemble, address1)
			So(q.Len(), ShouldEqual, 0)
		})

		Convey("Add duplicated worker", func() {
			q := newWorkerQueue()
			q.Add(newWorker(address1))
			So(q.Len(), ShouldEqual, 1)
			q.Add(newWorker(address1))
			So(q.Len(), ShouldEqual, 1)

			q.Add(newWorker(address2))
			q.Add(newWorker(address3))
			So(q.Len(), ShouldEqual, 3)
		})

		Convey("Return the last Add worker first", func() {
			q := newWorkerQueue()
			q.Add(newWorker(address1))
			q.Add(newWorker(address2))
			q.Add(newWorker(address3))
			So(q.Len(), ShouldEqual, 3)

			So(q.Next(), ShouldResemble, address3)
			So(q.Next(), ShouldResemble, address2)
			So(q.Next(), ShouldResemble, address1)
		})

		Convey("Tick a non exist worker", func() {
			q := newWorkerQueue()
			q.Tick(newWorker(address1))
			So(q.Len(), ShouldEqual, 0)
			q.Add(newWorker(address2))
			So(q.Len(), ShouldEqual, 1)
			q.Tick(newWorker(address1))
			So(q.Len(), ShouldEqual, 1)
		})

		Convey("Pruge multiple expired workers", func() {
			q := newWorkerQueue()
			q.Add(newWorker(address1))
			q.Add(newWorker(address2))
			So(q.Len(), ShouldEqual, 2)

			// Wait the worker to time out
			time.Sleep((HeartbeatLiveness + 1) * HeartbeatInterval)
			q.Add(newWorker(address3))
			So(q.Len(), ShouldEqual, 3)
			q.Purge()
			So(q.Len(), ShouldEqual, 1)
		})
	})
}
func TestBrokerWorker(t *testing.T) {
	const (
		workerAddr = "inproc://plugin.test"
	)
	broker, err := NewBroker("test", workerAddr)
	if err != nil {
		t.Fatalf("Failed to init broker: %v", err)
	}

	Convey("Test Broker", t, func() {
		Convey("receive Ready signal will register the worker", func() {
			w := workerSock(t, "ready", workerAddr)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))
			time.Sleep(HeartbeatInterval)

			So(broker.workers.Len(), ShouldEqual, 1)
			w.SendMessage(bytesArray(Shutdown))
		})

		Convey("receive multiple Ready signal will register all workers", func() {
			w1 := workerSock(t, "ready1", workerAddr)
			defer func() {
				w1.SendMessage(bytesArray(Shutdown))
				w1.Destroy()
			}()
			w1.SendMessage(bytesArray(Ready))
			w2 := workerSock(t, "ready2", workerAddr)
			defer func() {
				w2.SendMessage(bytesArray(Shutdown))
				w2.Destroy()
			}()
			w2.SendMessage(bytesArray(Ready))
			time.Sleep(HeartbeatInterval)

			So(broker.workers.Len(), ShouldEqual, 2)
			w1.SendMessage(bytesArray(Shutdown))
			w2.SendMessage(bytesArray(Shutdown))
		})

		Convey("reveice Heartbeat without Ready will not register the worker", func() {
			w := workerSock(t, "heartbeat", workerAddr)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Heartbeat))
			// Wait the poller to get the message
			time.Sleep(HeartbeatInterval)
			So(broker.workers.Len(), ShouldEqual, 0)
		})

		Convey("reveice worker message without Reay will be ignored", func() {
			w := workerSock(t, "unregistered", workerAddr)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage([][]byte{
				[]byte("unregistered"),
				[]byte{0},
				[]byte("Message to be ignored"),
			})
			// Wait the poller to get the message
			time.Sleep(HeartbeatInterval)
			So(broker.workers.Len(), ShouldEqual, 0)
		})

		Convey("receive RPC will timeout", func() {
			w := workerSock(t, "timeout", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))
			time.Sleep(HeartbeatInterval)

			So(broker.workers.Len(), ShouldEqual, 1)

			reqChan := make(chan chan []byte)
			broker.RPC(reqChan, []byte(("from server")))
			respChan := <-reqChan
			msg := <-respChan
			So(msg, ShouldResemble, []byte{0})
		})

		Convey("recive RPC without Ready worker will wait for Heartbeat Liveness time", func() {
			reqChan := make(chan chan []byte)
			timeout := time.Now().Add(HeartbeatInterval * HeartbeatLiveness)
			broker.RPC(reqChan, []byte(("from server")))
			respChan := <-reqChan
			resp := <-respChan
			So(resp, ShouldResemble, []byte{0})
			So(time.Now(), ShouldHappenAfter, timeout)
		})

		Convey("worker after recive RPC and before timeout will got the message", func() {
			reqChan := make(chan chan []byte)
			broker.RPC(reqChan, []byte(("from server")))
			respChan := <-reqChan
			time.Sleep(HeartbeatInterval)

			w := workerSock(t, "lateworker", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS * 2)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))

			msg := recvNonControlFrame(w)
			So(len(msg), ShouldEqual, 3)
			So(msg[2], ShouldResemble, []byte("from server"))
			msg[2] = []byte("from worker")
			w.SendMessage(msg)

			resp := <-respChan
			So(resp, ShouldResemble, []byte("from worker"))
		})

		Convey("broker RPC recive worker reply", func() {
			w := workerSock(t, "worker", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS * 2)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))

			reqChan := make(chan chan []byte)
			broker.RPC(reqChan, []byte(("from server")))
			respChan := <-reqChan

			msg := recvNonControlFrame(w)
			So(len(msg), ShouldEqual, 3)
			So(msg[2], ShouldResemble, []byte("from server"))
			msg[2] = []byte("from worker")
			w.SendMessage(msg)

			resp := <-respChan
			So(resp, ShouldResemble, []byte("from worker"))
		})

		Convey("send message from server to multiple plugin", func(c C) {
			go func() {
				w := workerSock(t, "worker1", workerAddr)
				w.SetRcvtimeo(heartbeatIntervalMS * 2)
				defer func() {
					w.SendMessage(bytesArray(Shutdown))
					time.Sleep((HeartbeatLiveness + 1) * HeartbeatInterval)
					w.Destroy()
				}()
				w.SendMessage(bytesArray(Ready))
				msg := recvNonControlFrame(w)
				if len(msg) == 3 {
					c.So(msg[2], ShouldResemble, []byte("from server"))
					msg[2] = []byte("from worker")
					w.SendMessage(msg)
				}
			}()

			go func() {
				w2 := workerSock(t, "worker2", workerAddr)
				w2.SetRcvtimeo(heartbeatIntervalMS * 2)
				defer func() {
					w2.SendMessage(bytesArray(Shutdown))
					time.Sleep((HeartbeatLiveness + 1) * HeartbeatInterval)
					w2.Destroy()
				}()
				w2.SendMessage(bytesArray(Ready))
				msg := recvNonControlFrame(w2)
				if len(msg) == 3 {
					c.So(msg[2], ShouldResemble, []byte("from server"))
					msg[2] = []byte("from worker")
					w2.SendMessage(msg)
				}
			}()

			reqChan := make(chan chan []byte)
			broker.RPC(reqChan, []byte(("from server")))
			respChan := <-reqChan
			resp := <-respChan
			So(resp, ShouldResemble, []byte("from worker"))
		})

		Convey("send multiple message from server to multple plugin", func(c C) {
			go func() {
				w := workerSock(t, "mworker1", workerAddr)
				w.SetRcvtimeo(heartbeatIntervalMS * 2)
				defer func() {
					w.SendMessage(bytesArray(Shutdown))
					w.Destroy()
				}()
				w.SendMessage(bytesArray(Ready))

				msg := recvNonControlFrame(w)
				c.So(len(msg), ShouldEqual, 3)
				c.So(msg[2], ShouldResemble, []byte("from server"))
				msg[2] = []byte("from worker1")
				w.SendMessage(msg)
			}()

			go func() {
				w2 := workerSock(t, "mworker2", workerAddr)
				w2.SetRcvtimeo(heartbeatIntervalMS * 2)
				defer func() {
					w2.SendMessage(bytesArray(Shutdown))
					w2.Destroy()
				}()
				w2.SendMessage(bytesArray(Ready))

				msg := recvNonControlFrame(w2)
				c.So(len(msg), ShouldEqual, 3)
				c.So(msg[2], ShouldResemble, []byte("from server"))
				msg[2] = []byte("from worker2")
				w2.SendMessage(msg)
			}()

			var wg sync.WaitGroup
			wg.Add(2)
			go func() {
				defer wg.Done()
				reqChan := make(chan chan []byte)
				broker.RPC(reqChan, []byte(("from server")))
				respChan := <-reqChan
				resp := <-respChan
				c.So(resp, ShouldNotBeEmpty)
			}()
			go func() {
				defer wg.Done()
				req2Chan := make(chan chan []byte)
				broker.RPC(req2Chan, []byte(("from server")))
				resp2Chan := <-req2Chan
				resp := <-resp2Chan
				c.So(resp, ShouldNotBeEmpty)
			}()
			wg.Wait()

		})
	})
	broker.stop <- 1

}
