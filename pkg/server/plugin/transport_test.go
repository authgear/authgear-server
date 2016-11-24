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

package plugin

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"golang.org/x/net/context"

	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
)

type nullTransport struct {
	state       TransportState
	initHandler TransportInitHandler
	lastContext context.Context
}

func (t *nullTransport) State() TransportState {
	return t.state
}
func (t *nullTransport) SetState(newState TransportState) {
	t.state = newState
}
func (t nullTransport) SendEvent(name string, in []byte) (out []byte, err error) {
	out = in
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
		ctx = context.WithValue(ctx, router.UserIDContextKey, "42")
		ctx = context.WithValue(ctx, router.AccessKeyTypeContextKey, router.MasterAccessKey)
		So(ContextMap(ctx), ShouldResemble, map[string]interface{}{
			"user_id":         "42",
			"access_key_type": "master",
		})
	})
}
