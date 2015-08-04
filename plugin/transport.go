package plugin

import "github.com/oursky/ourd/oddb"

// AuthRequest is sent by Ourd to plugin which contains data for authentication
type AuthRequest struct {
	ProviderName string
	Action       string
	AuthData     map[string]interface{}
}

// AuthResponse is sent by plugin to Ourd which contains authenticated data
type AuthResponse struct {
	PrincipalID string                 `json:"principal_id"`
	AuthData    map[string]interface{} `json:"auth_data"`
}

// A Transport represents the interface of data transfer between ourd
// and remote process.
type Transport interface {
	RunInit() ([]byte, error)
	RunLambda(name string, in []byte) ([]byte, error)
	RunHandler(name string, in []byte) ([]byte, error)

	// RunHook runs the hook specified by recordType and trigger, passing in
	// record as a parameter. Transport may not modify the record passed in.
	//
	// A oddb.Record is returned as a result of invocation. Such record must be
	// a newly allocated instance, and may not share any reference type values
	// in any of its memebers with the record being passed in.
	RunHook(recordType string, trigger string, record *oddb.Record) (*oddb.Record, error)
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
}

func (t nullTransport) RunInit() (out []byte, err error) {
	out = []byte{}
	return
}
func (t nullTransport) RunLambda(name string, in []byte) (out []byte, err error) {
	out = in
	return
}
func (t nullTransport) RunHandler(name string, in []byte) (out []byte, err error) {
	out = in
	return
}
func (t nullTransport) RunHook(recordType string, trigger string, reocrd *oddb.Record) (record *oddb.Record, err error) {
	return
}
func (t nullTransport) RunTimer(name string, in []byte) (out []byte, err error) {
	out = in
	return
}

func (t nullTransport) RunProvider(request *AuthRequest) (response *AuthResponse, err error) {
	response = &AuthResponse{
		AuthData: request.AuthData,
	}
	return
}

type nullFactory struct {
}

func (f nullFactory) Open(path string, args []string) Transport {
	return nullTransport{}
}
