package zmq

import (
	"bytes"
	"time"

	log "github.com/Sirupsen/logrus"
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
// with the addition of:
//
// 1. Shutdown signal, which signifies a normal termination of worker to provide
//    a fast path of worker removal
//
// NOTE(limouren): it might make a good interface
type Broker struct {
	// NOTE: goroutines are caller of plugin, so frontend is Go side,
	// backend is plugin side
	frontendAddr, backendAddr string
}

// NewBroker returns a new *Broker.
func NewBroker(frontendAddr, backendAddr string) (*Broker, error) {
	return &Broker{frontendAddr, backendAddr}, nil
}

// Run kicks start the Broker and listens for requests. It blocks function
// execution.
func (lb *Broker) Run() {
	frontend, backend := mustInitEndpoints(lb.frontendAddr, lb.backendAddr)
	backendPoller, bothPoller := mustInitPollers(frontend, backend)

	workers := workerQueue{}
	heartbeatAt := time.Now().Add(HeartbeatInterval)
	for {
		var sock *goczmq.Sock
		if workers.Len() == 0 {
			sock = backendPoller.Wait(heartbeatIntervalMS)
		} else {
			sock = bothPoller.Wait(heartbeatIntervalMS)
		}

		switch sock {
		case backend:
			frames, err := backend.RecvMessage()
			if err != nil {
				panic(err)
			}

			address := frames[0]
			workers.Ready(newWorker(address))

			msg := frames[1:]
			if len(msg) == 1 {
				status := string(msg[0])
				handleWorkerStatus(&workers, address, status)
			} else {
				frontend.SendMessage(msg)
				log.Debugf("zmq/broker: backend => frontend: %#x, %s\n", msg[0], msg)
			}
		case frontend:
			frames, err := frontend.RecvMessage()
			if err != nil {
				panic(err)
			}

			frames = append([][]byte{workers.Next()}, frames...)
			backend.SendMessage(frames)
			log.Debugf("zmq/broker: frontend => backend: %#x, %s\n", frames[0], frames)
		case nil:
			// do nothing
		default:
			panic("zmq/broker: received unknown socket")
		}

		if heartbeatAt.Before(time.Now()) {
			for _, worker := range workers {
				msg := [][]byte{worker.address, []byte(Heartbeat)}
				backend.SendMessage(msg)
			}
			heartbeatAt = time.Now().Add(HeartbeatInterval)
		}

		workers.Purge()
	}
}

func mustInitEndpoints(frontendAddr, backendAddr string) (*goczmq.Sock, *goczmq.Sock) {
	frontend, err := goczmq.NewRouter(frontendAddr)
	if err != nil {
		panic(err)
	}

	backend, err := goczmq.NewRouter(backendAddr)
	if err != nil {
		panic(err)
	}

	return frontend, backend
}

func mustInitPollers(frontend, backend *goczmq.Sock) (*goczmq.Poller, *goczmq.Poller) {
	backendPoller, err := goczmq.NewPoller(backend)
	if err != nil {
		panic(err)
	}

	bothPoller, err := goczmq.NewPoller(frontend, backend)
	if err != nil {
		panic(err)
	}

	return backendPoller, bothPoller
}

func handleWorkerStatus(workers *workerQueue, address []byte, status string) {
	switch status {
	case Ready:
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

var heartbeatIntervalMS = int(HeartbeatInterval.Seconds() * 1000)

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
	workers := *q

	var (
		i int
		w pworker
	)
	for i, w = range workers {
		if bytes.Equal(w.address, worker.address) {
			*q = append(append(workers[:i], workers[i+1:]...), worker)
			return
		}
	}
	*q = append(workers, worker)
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
