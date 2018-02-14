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
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	zmq "github.com/pebbe/zmq4"
	"github.com/sirupsen/logrus"
	"github.com/skygeario/skygear-server/pkg/server/router"
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
	Request  = "\004"
	Response = "\005"
)

// parcel is used to multiplex the chan with zmq worker
type parcel struct {
	// Worker is a map of "broker name to worker name"
	// Used to support multiple plugins
	// i.e Which worker is used for this request(parcel) for this broker
	workers     map[string]string
	bounceCount int
	requestID   string
	respChan    chan []byte
	frame       []byte
	retry       int
}

func newParcel(frame []byte) *parcel {
	return &parcel{
		workers:  make(map[string]string),
		respChan: make(chan []byte),
		frame:    frame,
		retry:    0,
	}
}

func (p *parcel) makePayload() (*router.Payload, error) {
	buffer := p.frame
	reader := bytes.NewReader(buffer)
	data := map[string]interface{}{}
	if jsonErr := json.NewDecoder(reader).Decode(&data); jsonErr != nil && jsonErr != io.EOF {
		return nil, jsonErr
	}

	payloadData := data["payload"].(map[string]interface{})
	method := data["method"].(string)
	actionString := payloadData["action"].(string)
	path := strings.Replace(actionString, ":", "/", -1)

	ctx := context.Background()
	ctx = context.WithValue(ctx, ZMQWorkerIDsContextKey, p.workers)
	ctx = context.WithValue(ctx, ZMQRequestIDContextKey, p.requestID)
	ctx = context.WithValue(ctx, ZMQBounceCountContextKey, p.bounceCount)

	payload := &router.Payload{
		Context: ctx,
		Meta: map[string]interface{}{
			"method": method,
			"path":   path,
		},
		Data:      payloadData,
		AccessKey: router.MasterAccessKey,
	}

	return payload, nil
}

var requestIDPrefix = ""

func generateRequestID() string {
	rand.Seed(time.Now().UnixNano())
	if requestIDPrefix == "" {
		requestIDPrefix = fmt.Sprintf("%X", rand.Intn(0x10000))
	}
	return fmt.Sprintf("%s-%X", requestIDPrefix, rand.Intn(0x10000))
}

// This is used for parcelChan, the "key to channel" map,
// the return key is used for adding/finding callback channels
func requestToChannelKey(requestID string, bounceCount int) string {
	return fmt.Sprintf("%s-%d", requestID, bounceCount)
}

// Broker implements a protocol based on the Paranoid Pirate queue described as follows:
// Related RFC: https://rfc.zeromq.org/spec:6/PPP
// refs: http://zguide.zeromq.org/py:all#Robust-Reliable-Queuing-Paranoid-Pirate-Pattern
// with the addition of:
//
// 1. Shutdown signal, which signifies a normal termination of worker to provide
//    a fast path of worker removal
// 2. Extra frames for bidirectional and multiplexing, see below
//    Related issue: https://github.com/SkygearIO/skygear-server/issues/295
//
// TODO: channeler can be separated into a separate struct, communiate with
// broker using frontend chan, workers and pull/push sock address.
//
// In PPP, the message has 3 frames, in the extended format, it has 7.
// 1. Worker address
// 2. Empty frame
// 3. Message type
// 4. Bounce Count
// 5. Request ID
// 6. Empty frame
// 7. Message body
//
// Message type is either Request or Response
// Bounce count is an integer that increase when the request is nested, starts from 0
// Request ID is a string that identify the request (including nested requests)
//
// The flow goes like this.
// # Connecting
// 1. A plugin connects to the server via ZMQ
// 2. Plugin sends a Ready message, with a worker ID
// 3. Server remembers that as a ready worker, and starts sending heartbeat.
//
// # Server sends requests to plugin
// 1. Server finds a free worker
// 2. Server sends the request message, e.g.
//    [WORKER-ADDR|0|REQ|0|REQ-ID|0|body....]
// 3. The plugin replies with the following message
//    [WORKER-ADDR|0|RES|0|REQ-ID|0|reply body...]
// Plugin can send requests to server in the exact same format
//
// # Nested requests (send a request when handling a request)
// 1. Server finds a free worker
// 2. Server sends the request message, e.g.
//    [WORKER-ADDR|0|REQ|0|REQ-ID|0|Do you have beer?]
// 3. The plugin sends a NESTED request to the server, bounce_count increases by 1
//    [WORKER-ADDR|0|REQ|1|REQ-ID|0|Do you have money?]
// 4. Server responds:
//    [WORKER-ADDR|0|RES|1|REQ-ID|0|Yes I have money]
// 5. Plugin responds:
//    [WORKER-ADDR|0|RES|0|REQ-ID|0|Yes I have beer]
//
// # Maximum bounce count
// broker has a maximum bounce count (maxBounce),
// when it is exceeded, the request fails immediately, this prevents infinity loop
//
// # About multiple plugins
// When multiple plugins are involved, both request ID and bounce-count is accumulated
// across plugins. For example, this is what happens when Plugin A calls Plugin B
//
// PluginA code:
// ```
// @op('foo:hello')
// def english():
//     send_action('foo:ciao')
//     return {'key':'thanks'}
// ```

// PluginB code:
// ```
// @op('foo:ciao')
// def italian():
//     return {'key':'grazie'}
// ```
// Would results in:
// 1. Browser calls /foo/hello
// 2. Server calls pluginA with `REQ|0|request-id`
// 3. PluginA calls server with `REQ|1|request-id`
// 4. Server calls pluginB with `REQ|2|request-id`
// 5. PluginB responses server with `RES|2|request-id`
// 6. Server responses pluginA with `RES|1|request-id`
// 7. PluginA responses server with `RES|0|request-id`

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
	// parcelChan is use zmq worker addess as key to route the message to
	// correct parcel for response chan
	parcelChan map[string]*parcel
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
	// The maximum bounce count, request larger than this number will be aborted
	maxBounce int
}

// NewBroker returns a new *Broker.
func NewBroker(name, backendAddr string, timeoutInterval int, maxBounce int) (*Broker, error) {
	namedLogger := log.WithFields(logrus.Fields{
		"plugin": name,
		"eaddr":  backendAddr,
	})

	broker := &Broker{
		name:            name,
		backendAddr:     backendAddr,
		frontend:        make(chan [][]byte, 10),
		recvChan:        make(chan *parcel, 10),
		parcelChan:      map[string]*parcel{},
		ReqChan:         make(chan *parcel, 10),
		respChan:        make(chan [][]byte, 10),
		timeout:         make(chan string),
		timeoutInterval: time.Duration(timeoutInterval),
		workers:         newWorkerQueue(),
		logger:          namedLogger,
		stop:            make(chan int),
		stopping:        false,
		maxBounce:       maxBounce,
	}

	go broker.Run()
	go broker.Channeler()
	return broker, nil
}

// Run the Broker and listens for zmq requests.
// nolint: gocyclo
func (lb *Broker) Run() {
	backend, err := zmq.NewSocket(zmq.ROUTER)
	if err != nil {
		panic(err)
	}
	defer backend.Close()

	if err := backend.Bind(lb.backendAddr); err != nil {
		panic(err)
	}

	pull, err := zmq.NewSocket(zmq.PULL)
	if err != nil {
		panic(err)
	}
	defer pull.Close()

	if err := pull.Bind(fmt.Sprintf("inproc://chanpipeline%d", lb.name)); err != nil {
		panic(err)
	}

	backendPoller := zmq.NewPoller()
	_ = backendPoller.Add(backend, zmq.POLLIN)
	_ = backendPoller.Add(pull, zmq.POLLIN)

	heartbeatAt := time.Now().Add(HeartbeatInterval)
	for {
		polleds, err := backendPoller.Poll(HeartbeatInterval)
		if err != nil {
			panic(err)
		}

		lb.workers.Lock()
		for _, polled := range polleds {
			switch s := polled.Socket; s {
			case backend:
				frames := mustReceiveMessage(backend)

				address := string(frames[0])
				msg := frames[1:]
				if len(msg) == 1 {
					tErr := lb.workers.Tick(newWorker(address))
					if tErr != nil {
						status := string(msg[0])
						if status != Ready {
							lb.logger.Warnln(tErr)
						}
					}
					status := string(msg[0])
					lb.handleWorkerStatus(address, status)
				} else {
					lb.frontend <- msg
					lb.logger.Debugf("zmq/broker: plugin => server: %q, %s\n", msg[0:6], msg[6])
				}
			case pull:
				frames := mustReceiveMessage(pull)
				mustSendMessage(backend, frames)
			default:
				panic("zmq/broker: received unknown socket")
			}
		}

		if heartbeatAt.Before(time.Now()) {
			for _, worker := range lb.workers.pworkers {
				msg := [][]byte{
					[]byte(worker.address),
					[]byte(Heartbeat),
				}
				mustSendMessage(backend, msg)
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

// Channeler bridges between golang channels and zmq channels,
// it accepts messages from RPC and dispatch to zmq,
// it accepts messages from zmq and dispatch to lb.ReqChan.
// It retries and timeout requests if no zmq worker is available.
func (lb *Broker) Channeler() {
	lb.logger.Infof("zmq channeler running %p\n", lb)
	defer lb.logger.Infof("zmq channeler stopped %p!\n", lb)

	push, err := zmq.NewSocket(zmq.PUSH)
	if err != nil {
		panic(err)
	}
	defer push.Close()

	if err := push.Connect(fmt.Sprintf("inproc://chanpipeline%d", lb.name)); err != nil {
		panic(err)
	}

	for {
		select {
		case frames := <-lb.frontend:
			lb.sendZMQFramesToChannel(frames)
		case p := <-lb.recvChan:
			lb.sendChannelParcelToZMQ(p, push)
		case frames := <-lb.respChan:
			mustSendMessage(push, frames)
		case key := <-lb.timeout:
			parcel, ok := lb.parcelChan[key]
			if !ok {
				break
			}
			lb.logger.Infof("zmq/broker: chan timeout for worker %s\n", key)
			delete(lb.parcelChan, key)
			parcel.respChan <- []byte{0}
		case <-lb.stop:
			lb.stopping = true
			return
		}
	}
}

func (lb *Broker) sendZMQFramesToChannel(frames [][]byte) {
	lb.logger.Debugf("zmq/broker: zmq => channel %q, %s\n", frames[0:6], frames[6])
	// Dispatch back to the channel based on the zmq first frame
	address := string(frames[0])
	messageType := string(frames[2])
	bounceCount, err := strconv.Atoi(string(frames[3]))
	if err != nil {
		lb.logger.Infof("zmq/broker: Cannot parse bounce_count int %v\n", frames)
		return
	}
	requestID := string(frames[4])
	message := frames[6]
	if messageType == Request {
		if bounceCount == 0 {
			// This is a new request from plugin, need to mark the worker as occupied
			lb.workers.Borrow(address)
		}
		go func() {
			parcel := newParcel(message)
			parcel.workers[lb.name] = address
			parcel.bounceCount = bounceCount
			parcel.requestID = requestID
			lb.ReqChan <- parcel
			response := <-parcel.respChan
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
			lb.logger.Debugf("zmq/broker: zmq => plugin %q, %s\n", frames[0:7], frames[7])
			lb.respChan <- frames
			if bounceCount == 0 {
				lb.workers.Lock()
				lb.workers.Tick(newWorker(address))
				lb.workers.Unlock()
			}
		}()
	} else if messageType == Response {
		key := requestToChannelKey(requestID, bounceCount)
		parcel, ok := lb.parcelChan[key]
		if !ok {
			lb.logger.Infof("zmq/broker: chan not found for worker %s\n", address)
			return
		}
		delete(lb.parcelChan, key)
		if bounceCount == 0 {
			lb.workers.Lock()
			lb.workers.Tick(newWorker(address))
			lb.workers.Unlock()
		}
		parcel.respChan <- message
	}
}

func (lb *Broker) sendChannelParcelToZMQ(p *parcel, push *zmq.Socket) {
	// Save the chan and dispatch the message to zmq
	// If current no worker ready, will retry after HeartbeatInterval.
	// Retry for HeartbeatLiveness times
	var address string
	var requestID string
	if _address, ok := p.workers[lb.name]; ok {
		address = _address
	} else {
		address = lb.workers.Next()
	}
	if p.requestID != "" {
		requestID = p.requestID
	} else {
		requestID = generateRequestID()
	}
	bounceCount := p.bounceCount
	if bounceCount > lb.maxBounce {
		lb.logger.Infof("zmq/broker: bounce count of %d exceeded the maximum %d\n", bounceCount, lb.maxBounce)
		p.respChan <- []byte{1}
		return
	}
	if address == "" {
		if p.retry < HeartbeatLiveness {
			p.retry += 1
			lb.logger.Infof("zmq/broker: no worker available, retry %d...\n", p.retry)
			go func(p2 *parcel) {
				time.Sleep(HeartbeatInterval)
				lb.recvChan <- p2
			}(p)
			return
		}
		lb.logger.Infof("zmq/broker: no worker available, timeout.\n")
		p.respChan <- []byte{0}
		return
	}
	p.workers[lb.name] = address
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
	lb.parcelChan[key] = p
	mustSendMessage(push, frames)
	lb.logger.Debugf("zmq/broker: channel => zmq: %q, %s\n", frames[0:7], frames[7])
	go lb.setTimeout(key)
}

func (lb *Broker) RPC(requestChan chan chan []byte, in []byte) {
	lb.RPCWithWorker(requestChan, in, make(map[string]string), "", 0)
}

func (lb *Broker) RPCWithWorker(requestChan chan chan []byte, in []byte, workerIDs map[string]string, requestID string, bounceCount int) {
	p := newParcel(in)
	p.workers = workerIDs
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
// and requires to Add itself again to become available.
//
// workerQueue is not goroutine safe. To use it safely across goroutine.
// Please use the Lock/Unlock interface before manupliate the queue item via
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

// Next will pop the next available worker, and the worker will not available
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

// Borrow will mark the specified worker as being comsumed,
// it is like Next() but you can specific which worker to take.
func (q *workerQueue) Borrow(address string) error {
	q.mu.Lock()
	defer q.mu.Unlock()
	workers := []pworker{}
	for _, w := range q.pworkers {
		if w.address != address {
			workers = append(workers, w)
		}
	}
	if len(workers) == len(q.pworkers) {
		return errors.New(fmt.Sprintf("zmq/broker: Cannot find worker = %s", address))
	}
	q.pworkers = workers
	return nil
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
	log.Debugf("zmq/broker: worker returns to pool = %s", worker.address)
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

func mustReceiveMessage(socket *zmq.Socket) [][]byte {
	frames, err := socket.RecvMessageBytes(0)
	if err != nil {
		panic(err)
	}

	return frames
}

func mustSendMessage(socket *zmq.Socket, message [][]byte) {
	_, err := socket.SendMessage(message)
	if err != nil {
		panic(err)
	}
}
