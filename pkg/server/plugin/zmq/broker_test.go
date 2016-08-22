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

	"github.com/zeromq/goczmq"
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

func TestBrokerEndToEnd(t *testing.T) {
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
