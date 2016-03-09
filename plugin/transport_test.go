package plugin

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	"github.com/oursky/skygear/skyconfig"
	"github.com/oursky/skygear/skydb"
)

type nullTransport struct {
	initHandler TransportInitHandler
	lastContext context.Context
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
func (t *nullTransport) RunLambda(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out = in
	t.lastContext = ctx
	return
}
func (t *nullTransport) RunHandler(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out = in
	return
}
func (t *nullTransport) RunHook(ctx context.Context, hookName string, reocrd *skydb.Record, oldRecord *skydb.Record) (record *skydb.Record, err error) {
	t.lastContext = ctx
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

func (f nullFactory) Open(path string, args []string, config skyconfig.Configuration) Transport {
	return &nullTransport{}
}

type fakeTransport struct {
	nullTransport
	outBytes []byte
	outErr   error
}

func (t *fakeTransport) RunLambda(ctx context.Context, name string, in []byte) (out []byte, err error) {
	t.lastContext = ctx
	if t.outErr == nil {
		out = t.outBytes
	} else {
		err = t.outErr
	}
	return
}

func (t *fakeTransport) RunHandler(ctx context.Context, name string, in []byte) (out []byte, err error) {
	t.lastContext = ctx
	if t.outErr == nil {
		out = t.outBytes
	} else {
		err = t.outErr
	}
	return
}

func TestContextMap(t *testing.T) {
	Convey("blank", t, func() {
		ctx := context.Background()
		So(ContextMap(ctx), ShouldResemble, map[string]interface{}{})
	})

	Convey("nil", t, func() {
		So(ContextMap(nil), ShouldResemble, map[string]interface{}{})
	})

	Convey("UserID", t, func() {
		ctx := context.Background()
		ctx = context.WithValue(ctx, "UserID", "42")
		So(ContextMap(ctx), ShouldResemble, map[string]interface{}{
			"user_id": "42",
		})
	})
}
