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
	"math/rand"
	"strconv"
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
	Request = "REQ"
	Response = "RES"
)

// parcel is used to multiplex the chan with zmq worker
type parcel struct {
	worker   string
	bounceCount int
	requestID   string
	respChan chan []byte
	frame    []byte
	retry    int
}

func newParcel(frame []byte) *parcel {
	return &parcel{
		worker:   "",
		respChan: make(chan []byte),
		frame:    frame,
		retry:    0,
	}
}

var requestIDPrefix = ""

func generateRequestID() string {
	rand.Seed(time.Now().UnixNano())
	if requestIDPrefix == "" {
		requestIDPrefix = fmt.Sprintf("%X", rand.Intn(0x10000))
	}
	return fmt.Sprintf("%s-%X", requestIDPrefix, rand.Intn(0x10000))
}

// This is used for addressChan, the "key to channel" map,
// the return key is used for adding/finding callback channels
func requestToChannelKey(requestID string, bounceCount int) string {
	return fmt.Sprintf("%s-%d", requestID, bounceCount)
}

// Broker implements the Paranoid Pirate queue described as follows:
// Related RFC: https://rfc.zeromq.org/spec:6/PPP
// refs: http://zguide.zeromq.org/py:all#Robust-Reliable-Queuing-Paranoid-Pirate-Pattern
// with the addition of:
//
// 1. Shutdown signal, which signifies a normal termination of worker to provide
//    a fast path of worker removal
// TODO: channeler can be separated into a separate struct, communiate with
// broker using frontend chan, workers and pull/push sock address.
type Broker struct {
	// name is assume to be unique and used to construct the zmq address
	name string
	// backendAddr is the address to communicate with plugin
	backendAddr string
	// frontend chan is receive zmq messgae and handle at Channeler
	frontend chan [][]byte
	// recvChan receive RPC request and dispatch to zmq Run Loop using
	// push/pull zsock
	recvChan chan *parcel
	// addressChan is use zmq worker addess as key to route the message to
	// correct go chan
	addressChan map[string]chan []byte
	// ReqChan is a channel for sending incoming request to handler
	ReqChan chan *parcel
	// respChan is a channel for sending response from handler to zmq
	respChan chan [][]byte
	// for RPC timeout, used by Channeler
	timeout chan string
	// determine how long will timeout happen, relative to the
	// HeartBeatInterval
	timeoutInterval time.Duration
	workers         workerQueue
	logger          *logrus.Entry
	// for stopping the channeler
	stop chan int
	// internal state for stopping the zmq Run when true.
	// Should use the stop chan to stop. The stop chan will set this variable.
	stopping bool
}

// NewBroker returns a new *Broker.
func NewBroker(name, backendAddr string, timeoutInterval int) (*Broker, error) {
	namedLogger := log.WithFields(logrus.Fields{
		"plugin": name,
		"eaddr":  backendAddr,
	})

	broker := &Broker{
		name:            name,
		backendAddr:     backendAddr,
		frontend:        make(chan [][]byte, 10),
		recvChan:        make(chan *parcel, 10),
		addressChan:     map[string]chan []byte{},
		ReqChan:         make(chan *parcel, 10),
		respChan:        make(chan [][]byte, 10),
		timeout:         make(chan string),
		timeoutInterval: time.Duration(timeoutInterval),
		workers:         newWorkerQueue(),
		logger:          namedLogger,
		stop:            make(chan int),
		stopping:        false,
	}

	go broker.Run()
	go broker.Channeler()
	return broker, nil
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
		if lb.stopping {
			return
		}
	}
}

// Channeler accept message from RPC and dispatch to zmq if available.
// It retry and timeout the request if the zmq worker is not yet available.
func (lb *Broker) Channeler() {
	lb.logger.Infof("zmq channeler running %p\n", lb)
	defer lb.logger.Infof("zmq channeler stopped %p!\n", lb)
	push, err := goczmq.NewPush(fmt.Sprintf("inproc://chanpipeline%d", lb.name))
	if err != nil {
		panic(err)
	}
	defer push.Destroy()
	for {
		select {
		case frames := <-lb.frontend:
			lb.logger.Debugf("zmq/broker: zmq => channel %#x, %s\n", frames[0], frames)
			// Dispacth back to the channel based on the zmq first frame
			address := string(frames[0])
			messageType := string(frames[2])
			bounceCount, err := strconv.Atoi(string(frames[3]))
			if err != nil {
				lb.logger.Infof("zmq/broker: Cannot parse bounce_count int %v\n", frames)
				break
			}
			requestID := string(frames[4])
			message := frames[6]
			if messageType == Request {
				parcel := newParcel(message)
				parcel.worker = address
				parcel.bounceCount = bounceCount
				parcel.requestID = requestID
				lb.ReqChan <- parcel
				go func (p *parcel) {
					response := <- p.respChan
					frames := [][]byte{
						[]byte(address),
						[]byte(address),
						[]byte{},
						[]byte(Response),
						[]byte(strconv.Itoa(bounceCount)),
						[]byte(requestID),
						[]byte{},
						response,
					}
					lb.logger.Debugf("zmq/broker: zmq => plugin %#x, %s\n", frames[0], frames)
					lb.respChan <- frames
				}(parcel)
			} else if messageType == Response {
				key := requestToChannelKey(requestID, bounceCount)
				respChan, ok := lb.addressChan[key]
				if !ok {
					lb.logger.Infof("zmq/broker: chan not found for worker %s\n", address)
					break
				}
				delete(lb.addressChan, key)
				respChan <- message
			}
		case p := <-lb.recvChan:
			// Save the chan and dispatch the message to zmq
			// If current no worker ready, will retry after HeartbeatInterval.
			// Retry for HeartbeatLiveness times
			var address string
			var bounceCount int
			var requestID string
			if p.worker != "" {
				address = p.worker
				bounceCount = p.bounceCount
				requestID = p.requestID
			} else {
				address = lb.workers.Next()
				bounceCount = 0
				requestID = generateRequestID()
			}
			if address == "" {
				if p.retry < HeartbeatLiveness {
					p.retry += 1
					lb.logger.Infof("zmq/broker: no worker available, retry %d...\n", p.retry)
					go func(p2 *parcel) {
						time.Sleep(HeartbeatInterval)
						lb.recvChan <- p2
					}(p)
					break
				}
				lb.logger.Infof("zmq/broker: no worker available, timeout.\n")
				p.respChan <- []byte{0}
				break
			}
			addr := []byte(address)
			frames := [][]byte{
				addr,
				addr,
				[]byte{},
				[]byte(Request),
				[]byte(strconv.Itoa(bounceCount)),
				[]byte(requestID),
				[]byte{},
				p.frame,
			}
			key := requestToChannelKey(requestID, bounceCount)
			lb.addressChan[key] = p.respChan
			push.SendMessage(frames)
			lb.logger.Debugf("zmq/broker: channel => zmq: %#x, %s\n", addr, frames)
			go lb.setTimeout(key)
		case frames := <-lb.respChan:
			push.SendMessage(frames)
		case key := <-lb.timeout:
			respChan, ok := lb.addressChan[key]
			if !ok {
				break
			}
			lb.logger.Infof("zmq/broker: chan timeout for worker %s\n", key)
			delete(lb.addressChan, key)
			respChan <- []byte{0}
		case <-lb.stop:
			lb.stopping = true
			return
		}
	}
}

func (lb *Broker) RPC(requestChan chan chan []byte, in []byte) {
	lb.RPCWithWorker(requestChan, in, "", "", 0)
}

func (lb *Broker) RPCWithWorker(requestChan chan chan []byte, in []byte, workerID string, requestID string, bounceCount int) {
	p := newParcel(in)
	p.worker = workerID
	p.bounceCount = bounceCount
	p.requestID = requestID
	lb.recvChan <- p
	go func() {
		requestChan <- p.respChan
	}()
}

func (lb *Broker) setTimeout(requestID string) {
	time.Sleep(HeartbeatInterval * lb.timeoutInterval)
	lb.timeout <- requestID
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
// available. The most recently Tick worker will got the job.
// A worker do not Tick itself within the expiry will regard as disconnected
// and requires to Add itself again to become avaliable.
//
// workerQueue is not goroutine safe. To use it safely across goroutine.
// Please use the Lock/Unlock interace before manupliate the queue item via
// methods like Add/Tick/Purge.
// Consuming the queue using Next is the only method will acquire the mutex lock
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
// This method for consuming the queue will acquire mutex lock.
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
// the next available worker.
func (q *workerQueue) Add(worker pworker) {
	q.addresses[worker.address] = true
	err := q.Tick(worker)
	if err == nil {
		return
	}
}

// Tick will make the worker to be the next available worker. Ticking an un-
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
// expiry. Intended for clean shutdown and fast removal of worker.
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
