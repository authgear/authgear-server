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
	"errors"
	"fmt"
	"sync"
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
	retry    int
}

func newParcel(frame []byte) *parcel {
	return &parcel{
		respChan: make(chan []byte),
		frame:    frame,
		retry:    0,
	}
}

// Broker implements the Paranoid Pirate queue described as follow:
// Related RFC: https://rfc.zeromq.org/spec:6/PPP
// refs: http://zguide.zeromq.org/py:all#Robust-Reliable-Queuing-Paranoid-Pirate-Pattern
// with the addition of:
//
// 1. Shutdown signal, which signifies a normal termination of worker to provide
//    a fast path of worker removal
type Broker struct {
	name        string
	backendAddr string
	frontend    chan [][]byte
	recvChan    chan *parcel
	addressChan map[string]chan []byte
	timeout     chan string
	workers     workerQueue
	logger      *logrus.Entry
	stop        chan int
}

// NewBroker returns a new *Broker.
func NewBroker(name, backendAddr string) (*Broker, error) {
	namedLogger := log.WithFields(logrus.Fields{
		"plugin": name,
		"eaddr":  backendAddr,
	})

	return &Broker{
		name:        name,
		backendAddr: backendAddr,
		frontend:    make(chan [][]byte, 10),
		recvChan:    make(chan *parcel, 10),
		addressChan: map[string]chan []byte{},
		timeout:     make(chan string),
		workers:     newWorkerQueue(),
		logger:      namedLogger,
		stop:        make(chan int),
	}, nil
}

// Run the Broker and listens for zmq requests.
func (lb *Broker) Run() {
	backend, err := goczmq.NewRouter(lb.backendAddr)
	if err != nil {
		panic(err)
	}
	defer backend.Destroy()

	pull, err := goczmq.NewPull(fmt.Sprintf("inproc://chanpipeline%d", lb.name))
	if err != nil {
		panic(err)
	}
	defer pull.Destroy()

	backendPoller, err := goczmq.NewPoller(backend, pull)
	if err != nil {
		panic(err)
	}

	heartbeatAt := time.Now().Add(HeartbeatInterval)
	for {
		sock := backendPoller.Wait(heartbeatIntervalMS)
		lb.workers.Lock()

		switch sock {
		case backend:
			frames, err := backend.RecvMessage()
			if err != nil {
				panic(err)
			}

			address := string(frames[0])
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
				lb.handleWorkerStatus(address, status)
			} else {
				lb.frontend <- msg
				lb.logger.Debugf("zmq/broker: plugin => server: %#x, %s\n", msg[0], msg)
			}
		case pull:
			frames, err := pull.RecvMessage()
			if err != nil {
				panic(err)
			}
			backend.SendMessage(frames)
		case nil:
			// do nothing
		default:
			panic("zmq/broker: received unknown socket")
		}

		if heartbeatAt.Before(time.Now()) {
			for _, worker := range lb.workers.pworkers {
				msg := [][]byte{
					[]byte(worker.address),
					[]byte(Heartbeat),
				}
				backend.SendMessage(msg)
			}
			heartbeatAt = time.Now().Add(HeartbeatInterval)
		}

		lb.workers.Purge()
		lb.workers.Unlock()
	}
}

// Channeler accept message from RPC and dispatch to zmq if avalible.
// It retry and timeout the request if the zmq worker is not yet avalible.
func (lb *Broker) Channeler() {
	lb.logger.Infof("zmq channler running %p\n", lb)
	push, err := goczmq.NewPush(fmt.Sprintf("inproc://chanpipeline%d", lb.name))
	if err != nil {
		panic(err)
	}
	defer push.Destroy()
ChannelerLoop:
	for {
		select {
		case frames := <-lb.frontend:
			lb.logger.Debugf("zmq/broker: zmq => channel %#x, %s\n", frames[0], frames)
			// Dispacth back to the channel based on the zmq first frame
			address := string(frames[0])
			respChan, ok := lb.addressChan[address]
			if !ok {
				lb.logger.Infof("zmq/broker: chan not found for worker %s\n", address)
				break
			}
			delete(lb.addressChan, address)
			respChan <- frames[2]
		case p := <-lb.recvChan:
			// Save the chan and dispatch the message to zmq
			// If current no worker ready, will retry after HeartbeatInterval.
			// Retry for HeartbeatLiveness tine
			address := lb.workers.Next()
			if address == "" {
				if p.retry < HeartbeatLiveness {
					p.retry += 1
					lb.logger.Infof("zmq/broker: no worker avaliable, retry %d...\n", p.retry)
					go func(p2 *parcel) {
						time.Sleep(HeartbeatInterval)
						lb.recvChan <- p2
					}(p)
					break
				}
				lb.logger.Infof("zmq/broker: no worker avaliable, timeout.\n")
				p.respChan <- []byte{0}
				break
			}
			addr := []byte(address)
			frames := [][]byte{
				addr,
				addr,
				[]byte{},
				p.frame,
			}
			lb.addressChan[address] = p.respChan
			push.SendMessage(frames)
			lb.logger.Debugf("zmq/broker: channel => zmq: %#x, %s\n", addr, frames)
			go lb.setTimeout(address, HeartbeatInterval*HeartbeatLiveness)
		case address := <-lb.timeout:
			respChan, ok := lb.addressChan[address]
			if !ok {
				break
			}
			lb.logger.Infof("zmq/broker: chan time out for  worker %s\n", address)
			delete(lb.addressChan, address)
			respChan <- []byte{0}
		case <-lb.stop:
			break ChannelerLoop
		}
	}
	lb.logger.Infof("zmq channler stopped %p!\n", lb)
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

func (lb *Broker) handleWorkerStatus(address string, status string) {
	switch status {
	case Ready:
		log.Infof("zmq/broker: ready worker = %s", address)
		lb.workers.Add(newWorker(address))
	case Heartbeat:
		// no-op
	case Shutdown:
		lb.workers.Remove(address)
		log.Infof("zmq/broker: shutdown of worker = %s", address)
	default:
		log.Errorf("zmq/broker: invalid status from worker = %s: %s", address, status)
	}
}

type pworker struct {
	address string
	expiry  time.Time
}

func newWorker(address string) pworker {
	return pworker{
		address,
		time.Now().Add(HeartbeatLiveness * HeartbeatInterval),
	}
}

// workerQueue is a last tick fist out queue.
//
// Worker is expect to register itself on ready. Tick itself when it is
// avalible. The most recently Tick worker will got the job.
// A worker do not Tick itself within the expiry will regards as disconnected
// and requires to Add itself again to become avaliable.
//
// workerQueue is not goroutine safe. To use it safely across goroutine.
// Please use the Lock/Unlock interace before manupliate the queue item via
// methods like Add/Tick/Purge.
// Comsuming method Next is the only method will acquire the mutex lock
// by itself.
type workerQueue struct {
	pworkers  []pworker
	addresses map[string]bool
	mu        *sync.Mutex
}

func newWorkerQueue() workerQueue {
	mu := &sync.Mutex{}
	return workerQueue{
		[]pworker{},
		map[string]bool{},
		mu,
	}
}
func (q *workerQueue) Lock() {
	q.mu.Lock()
}

func (q *workerQueue) Unlock() {
	q.mu.Unlock()
}

func (q *workerQueue) Len() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pworkers)
}

// Next will pop the next avaliable worker, and the worker will not avalible
// until it Tick back to the workerQueue again.
// This worker comsuing method will acquire mutex lock.
func (q *workerQueue) Next() string {
	q.mu.Lock()
	defer q.mu.Unlock()
	cnt := len(q.pworkers)
	if cnt == 0 {
		return ""
	}
	workers := q.pworkers
	worker := workers[len(workers)-1]
	q.pworkers = workers[:len(workers)-1]
	return worker.address
}

// Add will register the worker as live worker and call Tick to make itself to
// the next avaliable worker.
func (q *workerQueue) Add(worker pworker) {
	q.addresses[worker.address] = true
	err := q.Tick(worker)
	if err == nil {
		return
	}
}

// Tick will make the worker to be the next avalible worker. Ticking an un
// registered worker will be no-op.
func (q *workerQueue) Tick(worker pworker) error {
	if _, ok := q.addresses[worker.address]; !ok {
		return errors.New(fmt.Sprintf("zmq/broker: Ticking non-registered worker = %s", worker.address))
	}
	workers := q.pworkers

	for i, w := range workers {
		if w.address == worker.address {
			q.pworkers = append(append(workers[:i], workers[i+1:]...), worker)
			return nil
		}
	}
	q.pworkers = append(q.pworkers, worker)
	log.Debugf("zmq/broker: worker return to poll = %s", worker.address)
	return nil
}

// Purge will unregister the worker that is not heathly. i.e. haven't Tick for
// a while.
func (q *workerQueue) Purge() {
	workers := q.pworkers

	now := time.Now()
	for i, w := range workers {
		if w.expiry.After(now) {
			break
		}
		q.pworkers = workers[i+1:]
		delete(q.addresses, w.address)
		log.Infof("zmq/broker: disconnected worker = %s", w.address)
	}
}

// Remove will unregister the worker with specified address regardless of its
// expiry. Intented for clean shutdown and fast removal of worker.
func (q *workerQueue) Remove(address string) {
	delete(q.addresses, address)
	workers := q.pworkers

	for i, w := range workers {
		if w.address == address {
			q.pworkers = append(workers[:i], workers[i+1:]...)
			break
		}
	}
}
