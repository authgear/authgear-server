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
	"github.com/skygeario/skygear-server/pkg/server/router"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"golang.org/x/net/context"
)

// AuthRequest is sent by Skygear Server to plugin which contains data for authentication
type AuthRequest struct {
	ProviderName string
	Action       string
	AuthData     map[string]interface{}
}

// AuthResponse is sent by plugin to Skygear Server which contains authenticated data
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

	// TransportStateInitialized is the state when the transport has been
	// initialized. During this state, only requests from plugins with master key
	// will be accepted.
	TransportStateInitialized

	// TransportStateReady is the state when the transport is ready for
	// the requests from client.
	TransportStateReady

	// TransportStateWorkerUnavailable is the state when all workers
	// for the transport is not available
	TransportStateWorkerUnavailable

	// TransportStateError is the state when an error has occurred
	// in the transport and it is not able to serve requests
	TransportStateError
)

// TransportInitHandler models the handler for transport init
type TransportInitHandler func([]byte, error) error

// A Transport represents the interface of data transfer between skygear
// and remote process.
type Transport interface {
	State() TransportState
	SetState(TransportState)

	SendEvent(name string, in []byte) ([]byte, error)

	RunLambda(ctx context.Context, name string, in []byte) ([]byte, error)
	RunHandler(ctx context.Context, name string, in []byte) ([]byte, error)

	// RunHook runs the hook with a name recognized by plugin, passing in
	// record as a parameter. Transport may not modify the record passed in.
	//
	// A skydb.Record is returned as a result of invocation. Such record must be
	// a newly allocated instance, and may not share any reference type values
	// in any of its memebers with the record being passed in.
	RunHook(ctx context.Context, hookName string, record *skydb.Record, oldRecord *skydb.Record) (*skydb.Record, error)

	RunTimer(name string, in []byte) ([]byte, error)

	// RunProvider runs the auth provider with the specified AuthRequest.
	RunProvider(request *AuthRequest) (*AuthResponse, error)
}

// A TransportFactory is a generic interface to instantiates different
// kinds of Plugin Transport.
type TransportFactory interface {
	Open(path string, args []string, config skyconfig.Configuration) Transport
}

// ContextMap returns a map of the user request context.
func ContextMap(ctx context.Context) map[string]interface{} {
	if ctx == nil {
		return map[string]interface{}{}
	}
	pluginCtx := map[string]interface{}{}
	if userID, ok := ctx.Value(router.UserIDContextKey).(string); ok {
		pluginCtx["user_id"] = userID
	}
	if accessKeyType, ok := ctx.Value(router.AccessKeyTypeContextKey).(router.AccessKeyType); ok {
		switch accessKeyType {
		case router.ClientAccessKey:
			pluginCtx["access_key_type"] = "client"
		case router.MasterAccessKey:
			pluginCtx["access_key_type"] = "master"
		}
	}
	return pluginCtx
}
