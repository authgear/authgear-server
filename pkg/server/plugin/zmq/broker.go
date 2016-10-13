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
	"bytes"
	"errors"
	"fmt"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/zeromq/goczmq"
)

const (
	// HeartbeatInterval is the interval that broker and worker send
	// heartbeats to each other.
	HeartbeatInterval = time.Second
	// HeartbeatLiveness defines the liveness of each heartbeat. Generally
	// it should be >= 3, otherwise workers will keep being discarded and
	// reconnecting.
	HeartbeatLiveness = 3
)

var heartbeatIntervalMS = int(HeartbeatInterval.Seconds() * 1000)

const (
	// Ready is sent by worker to signal broker that it is ready to receive
	// jobs.
	Ready = "\001"
	// Heartbeat is sent by both broker and worker to signal a heartbeat.
	Heartbeat = "\002"
	// Shutdown is sent by worker while being killed (probably by CTRL C).
	// It is an addition to original PPP to shorten the time needed for
	// broker to detect a normal shutdown of worker.
	Shutdown = "\003"
)

// parcel is used to multiplex the chan with zmq worker
type parcel struct {
	respChan chan []byte
	frame    []byte
}

func newParcel(frame []byte) *parcel {
	return &parcel{
		respChan: make(chan []byte),
		frame:    frame,
	}
}

// Broker implements the Paranoid Pirate queue described in the zguide:
// http://zguide.zeromq.org/py:all#Robust-Reliable-Queuing-Paranoid-Pirate-Pattern
// Related RFC: https://rfc.zeromq.org/spec:6/PPP
// with the addition of:
//
// 1. Shutdown signal, which signifies a normal termination of worker to provide
//    a fast path of worker removal
type Broker struct {
	name          string
	backend       *goczmq.Sock
	bothPoller    *goczmq.Poller
	backendPoller *goczmq.Poller
	frontend      chan [][]byte
	recvChan      chan *parcel
	addressChan   map[string]chan []byte
	timeout       chan string
	workers       workerQueue
	freshWorkers  chan []byte
	logger        *logrus.Entry
	stop          chan int
}

// NewBroker returns a new *Broker.
func NewBroker(name, backendAddr string) (*Broker, error) {
	namedLogger := log.WithFields(logrus.Fields{
		"plugin": name,
		"eaddr":  backendAddr,
	})

	backend, err := goczmq.NewRouter(backendAddr)
	if err != nil {
		panic(err)
	}

	backendPoller, err := goczmq.NewPoller(backend)
	if err != nil {
		panic(err)
	}

	return &Broker{
		name:          name,
		backend:       backend,
		backendPoller: backendPoller,
		frontend:      make(chan [][]byte),
		recvChan:      make(chan *parcel),
		addressChan:   map[string]chan []byte{},
		timeout:       make(chan string),
		workers:       newWorkerQueue(),
		freshWorkers:  make(chan []byte, 1),
		logger:        namedLogger,
		stop:          make(chan int),
	}, nil
}

// Run kicks start the Broker and listens for requests. It blocks function
// execution.
func (lb *Broker) Run() {
	lb.logger.Infof("Running zmq broker")
	heartbeatAt := time.Now().Add(HeartbeatInterval)
	for {
		sock := lb.backendPoller.Wait(heartbeatIntervalMS)

		switch sock {
		case lb.backend:
			frames, err := lb.backend.RecvMessage()
			if err != nil {
				panic(err)
			}

			address := frames[0]
			msg := frames[1:]
			tErr := lb.workers.Tick(newWorker(address))
			if tErr != nil {
				status := string(msg[0])
				if status != Ready {
					lb.logger.Warnln(tErr)
				}
			}
			if len(msg) == 1 {
				status := string(msg[0])
				lb.handleWorkerStatus(&lb.workers, address, status)
			} else {
				lb.frontend <- msg
				lb.logger.Debugf("zmq/broker: plugin => server: %#x, %s\n", msg[0], msg)
			}
		case nil:
			// do nothing
		default:
			panic("zmq/broker: received unknown socket")
		}

		lb.logger.Debugf("zmq/broker: idle worker count %d\n", lb.workers.Len())
		if heartbeatAt.Before(time.Now()) {
			for _, worker := range lb.workers.pworkers {
				msg := [][]byte{worker.address, []byte(Heartbeat)}
				lb.logger.Debugf("zmq/broker: server => plugin Heartbeat: %s\n", worker.address)
				lb.backend.SendMessage(msg)
			}
			heartbeatAt = time.Now().Add(HeartbeatInterval)
		}

		lb.workers.Purge()
	}
}

func (lb *Broker) Channler() {
	lb.logger.Infof("zmq channler running %p", lb)
	for {
		select {
		case frames := <-lb.frontend:
			lb.logger.Debugf("zmq/broker: zmq => channel %#x, %s\n", frames[0], frames)
			// Dispacth back to the channel based on the zmq first frame
			address := string(frames[0])
			respChan, ok := lb.addressChan[address]
			if !ok {
				lb.logger.Infof("zmq/broker: chan not found for worker %#x\n", address)
				return
			}
			delete(lb.addressChan, address)
			respChan <- frames[2]
		case p := <-lb.recvChan:
			// Save the chan and dispatch the message to zmq
			addr := lb.workers.Next()
			frames := append([][]byte{addr}, addr, []byte{}, p.frame)
			address := string(addr)
			lb.addressChan[address] = p.respChan
			lb.backend.SendMessage(frames)
			lb.logger.Debugf("zmq/broker: channel => zmq: %#x, %s\n", addr, frames)
			go lb.setTimeout(address, HeartbeatInterval*HeartbeatLiveness)
		case address := <-lb.timeout:
			respChan, ok := lb.addressChan[address]
			if !ok {
				return
			}
			lb.logger.Infof("zmq/broker: chan time out for  worker %#x\n", address)
			delete(lb.addressChan, address)
			respChan <- []byte{0}
		case <-lb.stop:
			break
		}
	}
	lb.logger.Infof("zmq channler stopped %p!", lb)
}

func (lb *Broker) RPC(requestChan chan chan []byte, in []byte) {
	p := newParcel(in)
	lb.recvChan <- p
	go func() {
		requestChan <- p.respChan
	}()
}

func (lb *Broker) setTimeout(address string, wait time.Duration) {
	time.Sleep(wait)
	lb.timeout <- address
}

func (lb *Broker) handleWorkerStatus(workers *workerQueue, address []byte, status string) {
	switch status {
	case Ready:
		log.Infof("zmq/broker: ready worker = %s", address)
		workers.Add(newWorker(address))
		lb.freshWorkers <- address
	case Heartbeat:
		// no-op
	case Shutdown:
		workers.Remove(address)
		log.Infof("zmq/broker: shutdown of worker = %s", address)
	default:
		log.Errorf("zmq/broker: invalid status from worker = %s: %s", address, status)
	}
}

type pworker struct {
	address []byte
	expiry  time.Time
}

func newWorker(address []byte) pworker {
	return pworker{
		address,
		time.Now().Add(HeartbeatLiveness * HeartbeatInterval),
	}
}

// workerQueue is a last tick fist out queue.
// A worker need to register itself using Add before it can tick.
// Ticking of an non-registered worker will be no-ops.
type workerQueue struct {
	pworkers  []pworker
	addresses map[string]bool
}

func newWorkerQueue() workerQueue {
	return workerQueue{
		[]pworker{},
		map[string]bool{},
	}
}

func (q workerQueue) Len() int {
	return len(q.pworkers)
}

func (q *workerQueue) Next() []byte {
	workers := q.pworkers
	worker := workers[len(workers)-1]
	q.pworkers = workers[:len(workers)-1]
	return worker.address
}

func (q *workerQueue) Add(worker pworker) {
	q.addresses[string(worker.address)] = true
	err := q.Tick(worker)
	if err == nil {
		return
	}
}

func (q *workerQueue) Tick(worker pworker) error {
	if _, ok := q.addresses[string(worker.address)]; !ok {
		return errors.New(fmt.Sprintf("zmq/broker: Ticking non-registered worker = %s", worker.address))
	}
	workers := q.pworkers

	for i, w := range workers {
		if bytes.Equal(w.address, worker.address) {
			q.pworkers = append(append(workers[:i], workers[i+1:]...), worker)
			return nil
		}
	}
	q.pworkers = append(q.pworkers, worker)
	log.Debugf("zmq/broker: worker return to poll = %s", worker.address)
	return nil
}

func (q *workerQueue) Purge() {
	workers := q.pworkers

	now := time.Now()
	for i, w := range workers {
		if w.expiry.After(now) {
			break
		}
		q.pworkers = workers[i+1:]
		delete(q.addresses, string(w.address))
		log.Infof("zmq/broker: disconnected worker = %s", w.address)
	}
}

func (q *workerQueue) Remove(address []byte) {
	delete(q.addresses, string(address))
	workers := q.pworkers

	for i, w := range workers {
		if bytes.Equal(w.address, address) {
			q.pworkers = append(workers[:i], workers[i+1:]...)
			break
		}
	}
}
