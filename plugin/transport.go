package plugin

import "github.com/oursky/ourd/oddb"

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

type nullFactory struct {
}

func (f nullFactory) Open(path string, args []string) Transport {
	return nullTransport{}
}
