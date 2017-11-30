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
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/router"
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

		Convey("Borrow a worker", func() {
			q := newWorkerQueue()
			q.Add(newWorker(address1))
			q.Add(newWorker(address2))
			q.Add(newWorker(address3))
			So(q.Len(), ShouldEqual, 3)

			q.Borrow(address1)
			So(q.Next(), ShouldResemble, address3)
			So(q.Next(), ShouldResemble, address2)
		})
	})
}
func TestBrokerWorker(t *testing.T) {
	if runtime.GOMAXPROCS(0) == 1 {
		t.Skip("skipping zmq test in GOMAXPROCS=1")
	}

	Convey("Test Broker", t, func() {
		const (
			workerAddr = "inproc://plugin.test"
		)
		broker, err := NewBroker("test", workerAddr, 10, 10)
		if err != nil {
			t.Fatalf("Failed to init broker: %v", err)
		}

		Reset(func() {
			broker.stop <- 1
			time.Sleep(HeartbeatInterval * 2)
		})

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

		Convey("receive Heartbeat without Ready will not register the worker", func() {
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

		Convey("receive worker message without Ready will be ignored", func() {
			w := workerSock(t, "unregistered", workerAddr)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage([][]byte{
				[]byte("unregistered"),
				[]byte{0},
				[]byte(Request),
				[]byte("0"),
				[]byte("request-id"),
				[]byte{},
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

		Convey("receive RPC without Ready worker will wait for Heartbeat Liveness time", func() {
			reqChan := make(chan chan []byte)
			timeout := time.Now().Add(HeartbeatInterval * HeartbeatLiveness)
			broker.RPC(reqChan, []byte(("from server")))
			respChan := <-reqChan
			resp := <-respChan
			So(resp, ShouldResemble, []byte{0})
			So(time.Now(), ShouldHappenAfter, timeout)
		})

		Convey("worker after receive RPC and before timeout will got the message", func() {
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
			So(len(msg), ShouldEqual, 7)
			So(msg[6], ShouldResemble, []byte("from server"))
			msg[2] = []byte(Response)
			msg[6] = []byte("from worker")
			w.SendMessage(msg)

			resp := <-respChan
			So(resp, ShouldResemble, []byte("from worker"))
		})

		Convey("broker RPC receive worker reply", func() {
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
			So(len(msg), ShouldEqual, 7)
			So(msg[6], ShouldResemble, []byte("from server"))
			msg[2] = []byte(Response)
			msg[6] = []byte("from worker")
			w.SendMessage(msg)

			resp := <-respChan
			So(resp, ShouldResemble, []byte("from worker"))
		})

		Convey("broker RPC receive worker request", func() {
			w := workerSock(t, "worker", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS * 2)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))

			w.SendMessage([][]byte{
				[]byte("worker"),
				[]byte{},
				[]byte(Request),
				[]byte("0"),
				[]byte("request-id"),
				[]byte{},
				[]byte("request from plugin"),
			})

			parcel := <-broker.ReqChan

			So(parcel.requestID, ShouldResemble, "request-id")
			So(parcel.workers, ShouldResemble, map[string]string{
				"test": "worker",
			})
			So(parcel.bounceCount, ShouldEqual, 0)

			parcel.respChan <- []byte("server response")

			msg := recvNonControlFrame(w)
			So(len(msg), ShouldEqual, 7)
			So(msg[0], ShouldResemble, []byte("worker"))
			So(msg[1], ShouldResemble, []byte{})
			So(msg[2], ShouldResemble, []byte(Response))
			So(msg[3], ShouldResemble, []byte("0"))
			So(msg[4], ShouldResemble, []byte("request-id"))
			So(msg[5], ShouldResemble, []byte{})
			So(msg[6], ShouldResemble, []byte("server response"))
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
				if len(msg) == 7 {
					c.So(msg[6], ShouldResemble, []byte("from server"))
					msg[2] = []byte(Response)
					msg[6] = []byte("from worker")
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
				if len(msg) == 7 {
					c.So(msg[6], ShouldResemble, []byte("from server"))
					msg[2] = []byte(Response)
					msg[6] = []byte("from worker")
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
				c.So(len(msg), ShouldEqual, 7)
				c.So(msg[6], ShouldResemble, []byte("from server"))
				msg[6] = []byte("from worker1")
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
				c.So(len(msg), ShouldEqual, 7)
				c.So(msg[6], ShouldResemble, []byte("from server"))
				msg[6] = []byte("from worker2")
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

		Convey("broker RPC handle nested request", func() {
			w := workerSock(t, "worker", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS * 2)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))

			reqChan := make(chan chan []byte)
			// worker send req to plugin
			broker.RPCWithWorker(
				reqChan,
				[]byte("from server"),
				make(map[string]string),
				"request-id",
				0,
			)

			time.Sleep(HeartbeatInterval)

			// plugin send nested request
			w.SendMessage([][]byte{
				[]byte("worker"),
				[]byte{},
				[]byte(Request),
				[]byte("1"),
				[]byte("request-id"),
				[]byte{},
				[]byte("request from plugin"),
			})

			// handle nested request
			parcel := <-broker.ReqChan

			So(parcel.requestID, ShouldResemble, "request-id")
			So(parcel.workers, ShouldResemble, map[string]string{
				"test": "worker",
			})
			So(parcel.bounceCount, ShouldEqual, 1)
			So(parcel.frame, ShouldResemble, []byte("request from plugin"))
			// server reply nested request
			parcel.respChan <- []byte("from server")

			// plugin get nested reply
			msg := recvNonControlFrame(w)
			So(len(msg), ShouldEqual, 7)
			So(msg[6], ShouldResemble, []byte("from server"))
			// plugin main reply
			w.SendMessage([][]byte{
				[]byte("worker"),
				[]byte{},
				[]byte(Response),
				[]byte("0"),
				[]byte("request-id"), // here update + assert?!
				[]byte{},
				[]byte("response from worker"),
			})

			// server get main response
			respChan := <-reqChan
			resp := <-respChan
			So(resp, ShouldResemble, []byte("response from worker"))
		})

		Convey("broker returns error when bounce count is over maximum", func() {
			w := workerSock(t, "worker", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS * 2)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))

			reqChan := make(chan chan []byte)
			broker.maxBounce = 10
			broker.RPCWithWorker(
				reqChan,
				[]byte("from server"),
				make(map[string]string),
				"request-id",
				11,
			)
			respChan := <-reqChan
			resp := <-respChan
			So(resp, ShouldResemble, []byte{1})
		})

		Convey("Parcel should create payload", func() {
			w := workerSock(t, "worker", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS * 2)
			defer func() {
				w.SendMessage(bytesArray(Shutdown))
				w.Destroy()
			}()
			w.SendMessage(bytesArray(Ready))

			w.SendMessage([][]byte{
				[]byte("worker"),
				[]byte{},
				[]byte(Request),
				[]byte("0"),
				[]byte("request-id"),
				[]byte{},
				[]byte("{\"method\":\"POST\", \"payload\":{\"action\":\"foo:bar\", \"key\":\"value\"}}"),
			})

			parcel := <-broker.ReqChan
			payload, err := parcel.makePayload()

			if err != nil {
				t.Fatalf("Failed to create payload: %v", err)
			}

			So(payload.Meta["method"], ShouldResemble, "POST")
			So(payload.Meta["path"], ShouldResemble, "foo/bar")
			So(payload.Data["key"], ShouldResemble, "value")
			So(payload.AccessKey, ShouldEqual, router.MasterAccessKey)

			ctx := payload.Context
			So(ctx.Value(ZMQRequestIDContextKey), ShouldEqual, "request-id")
			So(ctx.Value(ZMQBounceCountContextKey), ShouldEqual, 0)
			So(ctx.Value(ZMQWorkerIDsContextKey), ShouldResemble, map[string]string{
				"test": "worker",
			})
		})
	})
}
