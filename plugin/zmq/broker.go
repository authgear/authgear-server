package zmq

import (
	log "github.com/Sirupsen/logrus"
	"github.com/zeromq/goczmq"
)

// Broker handles the exchange of messages between frontend and backend socket.
//
// NOTE(limouren): it might make a good interface
type Broker struct {
	// NOTE: goroutines are caller of plugin, so frontend is Go side,
	// backend is plugin side
	frontend, backend *goczmq.Sock
	poller            *goczmq.Poller
}

// NewBroker returns a new *Broker.
func NewBroker(frontendAddr, backendAddr string) (*Broker, error) {
	frontend, err := goczmq.NewRouter(frontendAddr)
	if err != nil {
		return nil, err
	}

	backend := goczmq.NewSock(goczmq.Dealer)
	_, err = backend.Bind(backendAddr)
	if err != nil {
		return nil, err
	}

	poller, err := goczmq.NewPoller(frontend, backend)
	if err != nil {
		return nil, err
	}

	return &Broker{frontend, backend, poller}, nil
}

// Run kicks start the Broker and listens for requests. It blocks function
// execution.
func (lb *Broker) Run() {
	for {
		sock := lb.poller.Wait(-1)
		switch sock {
		case lb.frontend:
			lb.handleFrontendMsg()
		case lb.backend:
			lb.handleBackendMsg()
		case nil:
			// it probably won't happen as we wait for poller forever
			log.Warnln("zmq/broker: received nil socket, ignoring")
		default:
			panic("zmq/broker: received unknown socket")
		}
	}
}

// Destroy cleans up resources allocated by the Broker.
func (lb *Broker) Destroy() {
	lb.frontend.Destroy()
	lb.backend.Destroy()
}

func (lb *Broker) handleFrontendMsg() {
	msg, err := lb.frontend.RecvMessage()
	if err != nil {
		panic(err)
	}

	lb.backend.SendMessage(msg)
	log.Debugf("zmq/broker: frontend => backend: %#x, %s\n", msg[0], msg)
}

func (lb *Broker) handleBackendMsg() {
	msg, err := lb.backend.RecvMessage()
	if err != nil {
		panic(err)
	}

	lb.frontend.SendMessage(msg)
	log.Debugf("zmq/broker:  backend => frontend: %#x, %s\n", msg[0], msg)
}
