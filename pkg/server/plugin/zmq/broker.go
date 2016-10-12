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

// Broker implements the Paranoid Pirate queue described in the zguide:
// http://zguide.zeromq.org/py:all#Robust-Reliable-Queuing-Paranoid-Pirate-Pattern
// Related RFC: https://rfc.zeromq.org/spec:6/PPP
// with the addition of:
//
// 1. Shutdown signal, which signifies a normal termination of worker to provide
//    a fast path of worker removal
type Broker struct {
	name string
	// NOTE: goroutines are caller of plugin, so frontend is Go side,
	// backend is plugin side
	frontend      *goczmq.Sock
	backend       *goczmq.Sock
	bothPoller    *goczmq.Poller
	backendPoller *goczmq.Poller
	freshWorkers  chan []byte
	logger        *logrus.Entry
}

// NewBroker returns a new *Broker.
func NewBroker(name, frontendAddr, backendAddr string) (*Broker, error) {
	namedLogger := log.WithFields(logrus.Fields{"plugin": name})
	frontend, err := goczmq.NewRouter(frontendAddr)
	if err != nil {
		panic(err)
	}

	backend, err := goczmq.NewRouter(backendAddr)
	if err != nil {
		panic(err)
	}

	backendPoller, err := goczmq.NewPoller(backend)
	if err != nil {
		panic(err)
	}

	bothPoller, err := goczmq.NewPoller(frontend, backend)
	if err != nil {
		panic(err)
	}

	return &Broker{
		name:          name,
		frontend:      frontend,
		backend:       backend,
		bothPoller:    bothPoller,
		backendPoller: backendPoller,
		freshWorkers:  make(chan []byte, 1),
		logger:        namedLogger,
	}, nil
}

// Run kicks start the Broker and listens for requests. It blocks function
// execution.
func (lb *Broker) Run() {
	workers := workerQueue{}
	heartbeatAt := time.Now().Add(HeartbeatInterval)
	for {
		var sock *goczmq.Sock
		if workers.Len() == 0 {
			sock = lb.backendPoller.Wait(heartbeatIntervalMS)
		} else {
			sock = lb.bothPoller.Wait(heartbeatIntervalMS)
		}

		switch sock {
		case lb.backend:
			frames, err := lb.backend.RecvMessage()
			if err != nil {
				panic(err)
			}

			address := frames[0]

			tErr := workers.Tick(newWorker(address))
			if tErr != nil {
				lb.logger.Warnln(tErr)
			}

			msg := frames[1:]
			if len(msg) == 1 {
				status := string(msg[0])
				lb.handleWorkerStatus(&workers, address, status)
			} else {
				lb.frontend.SendMessage(msg)
				lb.logger.Debugf("zmq/broker: backend => frontend: %#x, %s\n", msg[0], msg)
			}
		case lb.frontend:
			frames, err := lb.frontend.RecvMessage()
			if err != nil {
				panic(err)
			}

			frames = append([][]byte{workers.Next()}, frames...)
			lb.backend.SendMessage(frames)
			lb.logger.Debugf("zmq/broker: frontend => backend: %#x, %s\n", frames[0], frames)
		case nil:
			// do nothing
		default:
			panic("zmq/broker: received unknown socket")
		}

		if heartbeatAt.Before(time.Now()) {
			for _, worker := range workers {
				msg := [][]byte{worker.address, []byte(Heartbeat)}
				lb.backend.SendMessage(msg)
			}
			heartbeatAt = time.Now().Add(HeartbeatInterval)
		}

		workers.Purge()
	}
}

func (lb *Broker) handleWorkerStatus(workers *workerQueue, address []byte, status string) {
	switch status {
	case Ready:
		workers.Ready(newWorker(address))
		lb.freshWorkers <- address
		log.Infof("zmq/broker: ready worker = %s", address)
	case Heartbeat:
		// do nothing
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

type workerQueue []pworker

func (q workerQueue) Len() int {
	return len(q)
}

func (q *workerQueue) Next() []byte {
	workers := *q
	worker := workers[len(workers)-1]
	*q = workers[:len(workers)-1]
	return worker.address
}

func (q *workerQueue) Ready(worker pworker) {
	err := q.Tick(worker)
	if err == nil {
		return
	}
	workers := *q
	*q = append(workers, worker)
}

func (q *workerQueue) Tick(worker pworker) error {
	workers := *q

	var (
		i int
		w pworker
	)
	for i, w = range workers {
		if bytes.Equal(w.address, worker.address) {
			*q = append(append(workers[:i], workers[i+1:]...), worker)
			return nil
		}
	}

	return errors.New(fmt.Sprintf("zmq/broker: Ticking non-existing worker = %s", worker.address))
}

func (q *workerQueue) Purge() {
	workers := *q

	now := time.Now()
	for i, w := range workers {
		if w.expiry.After(now) {
			break
		}
		*q = workers[i+1:]
		log.Infof("zmq/broker: disconnected worker = %s", w.address)
	}
}

func (q *workerQueue) Remove(address []byte) {
	workers := *q

	for i, w := range workers {
		if bytes.Equal(w.address, address) {
			*q = append(workers[:i], workers[i+1:]...)
			break
		}
	}
}
