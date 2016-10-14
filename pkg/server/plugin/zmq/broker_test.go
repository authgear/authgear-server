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
	"reflect"
	"strings"
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

func bytesArray(ss ...string) (bs [][]byte) {
	for _, s := range ss {
		bs = append(bs, []byte(s))
	}
	return
}

func TestBrokerSock(t *testing.T) {
	const (
		clientAddr = "inproc://client.test"
		workerAddr = "inproc://worker.test"
	)
	broker, err := NewBroker("", clientAddr, workerAddr)
	if err != nil {
		t.Fatalf("Failed to init broker: %v", err)
	}
	go broker.Run()

	w0 := workerSock(t, "w0", workerAddr)
	defer w0.Destroy()
	w0.SendMessage(bytesArray(Ready))
	<-broker.freshWorkers

	c0 := clientSock(t, "c0", clientAddr)
	defer c0.Destroy()
	c0.SendMessage(bytesArray("simple job"))

	msg, _ := w0.RecvMessage()
	expectedMsg := bytesArray(c0.Identity(), "simple job")
	if !reflect.DeepEqual(msg, expectedMsg) {
		t.Fatalf(`want %v, got %v`, expectedMsg, msg)
	}

	w0.SendMessage(expectedMsg)
	msg, _ = c0.RecvMessage()
	expectedMsg = bytesArray("simple job")
	if !reflect.DeepEqual(msg, expectedMsg) {
		t.Fatalf("want %v, got %v", expectedMsg, msg)
	}

	// multiple workers, multiple clients
	w1 := workerSock(t, "w1", workerAddr)
	defer w1.Destroy()
	w1.SendMessage(bytesArray(Ready))
	<-broker.freshWorkers

	c1 := clientSock(t, "c1", clientAddr)
	defer c1.Destroy()
	c2 := clientSock(t, "c2", clientAddr)
	defer c2.Destroy()

	c0.SendMessage(bytesArray("job 0"))
	c1.SendMessage(bytesArray("job 1"))
	c2.SendMessage(bytesArray("job 2"))

	work0, _ := w0.RecvMessage()
	if !strings.HasPrefix(string(work0[1]), "job") {
		t.Fatalf(`want job *, got %s`, work0[1])
	}

	work1, _ := w1.RecvMessage()
	if !strings.HasPrefix(string(work1[1]), "job") {
		t.Fatalf(`want job, got %v`, work1[1])
	}

	// let w0 complete its work and receive new job
	w0.SendMessage(work0)
	work2, _ := w0.RecvMessage()
	if !strings.HasPrefix(string(work2[1]), "job") {
		t.Fatalf(`want job, got %v`, work2[1])
	}

	// complete all jobs
	w0.SendMessage(work2)
	w1.SendMessage(work1)

	// clients receive all jobs
	resp0, err := c0.RecvMessage()
	resp1, err := c1.RecvMessage()
	resp2, err := c2.RecvMessage()

	resps := []string{
		string(resp0[0]),
		string(resp1[0]),
		string(resp2[0]),
	}
	if !reflect.DeepEqual(resps, []string{
		"job 0",
		"job 1",
		"job 2",
	}) {
		t.Fatalf(`want ["job 0", "job 1", "job 2"], got %v`, resps)
	}
}

func TestWorker(t *testing.T) {
	Convey("Test workerQueue", t, func() {
		address1 := []byte("address1")
		address2 := []byte("address2")
		address3 := []byte("address3")

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
		clientAddr = "inproc://server.test"
		workerAddr = "inproc://plugin.test"
	)
	broker, err := NewBroker("", clientAddr, workerAddr)
	if err != nil {
		t.Fatalf("Failed to init broker: %v", err)
	}
	go broker.Run()

	Convey("Test Broker with worker", t, func() {
		Convey("receive Ready signal will register the worker", func() {
			w := workerSock(t, "ready", workerAddr)
			defer w.Destroy()
			w.SendMessage(bytesArray(Ready))
			<-broker.freshWorkers

			So(broker.workers.Len(), ShouldEqual, 1)
			w.SendMessage(bytesArray(Shutdown))
		})

		Convey("receive multiple Ready signal will register all workers", func() {
			w1 := workerSock(t, "ready1", workerAddr)
			defer w1.Destroy()
			w1.SendMessage(bytesArray(Ready))
			<-broker.freshWorkers
			w2 := workerSock(t, "ready2", workerAddr)
			defer w2.Destroy()
			w2.SendMessage(bytesArray(Ready))
			<-broker.freshWorkers

			So(broker.workers.Len(), ShouldEqual, 2)
			w1.SendMessage(bytesArray(Shutdown))
			w2.SendMessage(bytesArray(Shutdown))
		})

		Convey("reveice Heartbeat without Reay will not register the worker", func() {
			w := workerSock(t, "heartbeat", workerAddr)
			defer w.Destroy()
			w.SendMessage(bytesArray(Heartbeat))
			// Wait the poller to get the message
			time.Sleep(HeartbeatInterval)
			So(broker.workers.Len(), ShouldEqual, 0)
		})

		Convey("send message from server to plugin", func() {
			s := clientSock(t, "server", clientAddr)
			defer s.Destroy()
			w := workerSock(t, "worker", workerAddr)
			w.SetRcvtimeo(heartbeatIntervalMS)
			defer w.Destroy()
			w.SendMessage(bytesArray(Ready))
			<-broker.freshWorkers
			w2 := workerSock(t, "worker2", workerAddr)
			w2.SetRcvtimeo(heartbeatIntervalMS)
			defer w2.Destroy()
			w2.SendMessage(bytesArray(Ready))
			<-broker.freshWorkers

			So(broker.workers.Len(), ShouldEqual, 2)

			s.SendMessage(bytesArray("from server"))
			// Wait the poller to get the message
			time.Sleep(HeartbeatInterval)
			So(broker.workers.Len(), ShouldEqual, 1)
			msg, _ := w2.RecvMessage()
			So(msg[1], ShouldResemble, []byte("from server"))
		})
	})
}
