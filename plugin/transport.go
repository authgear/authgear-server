package plugin

import "github.com/oursky/skygear/skydb"

// AuthRequest is sent by Skygear to plugin which contains data for authentication
type AuthRequest struct {
	ProviderName string
	Action       string
	AuthData     map[string]interface{}
}

// AuthResponse is sent by plugin to Skygear which contains authenticated data
type AuthResponse struct {
	PrincipalID string                 `json:"principal_id"`
	AuthData    map[string]interface{} `json:"auth_data"`
}

// TransportState refers to the operation state of the transport
//go:generate stringer -type=TransportState
type TransportState int

const (
	// TransportStateUninitialized is the state when the transport has not
	// been initialized
	TransportStateUninitialized TransportState = iota

	// TransportStateReady is the state when the transport is ready for
	// requests
	TransportStateReady

	// TransportStateWorkerUnavailable is the state when all workers
	// for the transport is not available
	TransportStateWorkerUnavailable

	// TransportStateError is the state when an error has occurred
	// in the transport and it is not able to serve requests
	TransportStateError
)

type TransportInitHandler func([]byte, error) error

// A Transport represents the interface of data transfer between skygear
// and remote process.
type Transport interface {
	State() TransportState
	SetInitHandler(TransportInitHandler)
	RequestInit()
	RunInit() ([]byte, error)
	RunLambda(name string, in []byte) ([]byte, error)
	RunHandler(name string, in []byte) ([]byte, error)

	// RunHook runs the hook specified by recordType and trigger, passing in
	// record as a parameter. Transport may not modify the record passed in.
	//
	// A skydb.Record is returned as a result of invocation. Such record must be
	// a newly allocated instance, and may not share any reference type values
	// in any of its memebers with the record being passed in.
	RunHook(recordType string, trigger string, record *skydb.Record, oldRecord *skydb.Record) (*skydb.Record, error)
	RunTimer(name string, n []byte) ([]byte, error)

	// RunProvider runs the auth provider with the specified AuthRequest.
	RunProvider(request *AuthRequest) (*AuthResponse, error)
}

// A TransportFactory is a generic interface to instantiates different
// kinds of Plugin Transport.
type TransportFactory interface {
	Open(path string, args []string) Transport
}

type nullTransport struct {
	initHandler TransportInitHandler
}

func (t *nullTransport) State() TransportState {
	return TransportStateReady
}

func (t *nullTransport) SetInitHandler(f TransportInitHandler) {
	t.initHandler = f
}

func (t *nullTransport) RequestInit() {
	if t.initHandler != nil {
		t.initHandler([]byte{}, nil)
	}
	return
}
func (t nullTransport) RunInit() (out []byte, err error) {
	out = []byte{}
	return
}
func (t *nullTransport) RunLambda(name string, in []byte) (out []byte, err error) {
	out = in
	return
}
func (t *nullTransport) RunHandler(name string, in []byte) (out []byte, err error) {
	out = in
	return
}
func (t *nullTransport) RunHook(recordType string, trigger string, reocrd *skydb.Record, oldRecord *skydb.Record) (record *skydb.Record, err error) {
	return
}
func (t *nullTransport) RunTimer(name string, in []byte) (out []byte, err error) {
	out = in
	return
}

func (t *nullTransport) RunProvider(request *AuthRequest) (response *AuthResponse, err error) {
	if request.AuthData == nil {
		request.AuthData = map[string]interface{}{}
	}
	response = &AuthResponse{
		AuthData: request.AuthData,
	}
	return
}

type nullFactory struct {
}

func (f nullFactory) Open(path string, args []string) Transport {
	return &nullTransport{}
}
