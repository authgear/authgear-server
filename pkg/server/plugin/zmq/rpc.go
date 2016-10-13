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
	"encoding/json"
	"fmt"
	"runtime/debug"

	"github.com/Sirupsen/logrus"

	skyplugin "github.com/skygeario/skygear-server/pkg/server/plugin"
	"github.com/skygeario/skygear-server/pkg/server/plugin/common"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/zeromq/goczmq"
	"golang.org/x/net/context"
)

const initRequestTimeout = 2000

type zmqTransport struct {
	state       skyplugin.TransportState
	name        string
	iaddr       string // the internal addr used by goroutines to make request to plugin
	broker      *Broker
	initHandler skyplugin.TransportInitHandler
	logger      *logrus.Entry
	config      skyconfig.Configuration
}

type request struct {
	Context context.Context
	Kind    string
	Name    string
	Param   interface{}
	Timeout int // timeout in millisecond
}

type hookRequest struct {
	Record   interface{} `json:"record"`
	Original interface{} `json:"original"`
}

// type-safe constructors for request.Param assignment

func newLambdaRequest(ctx context.Context, name string, args json.RawMessage) *request {
	return &request{Kind: "op", Name: name, Param: args, Context: ctx}
}

func newHandlerRequest(ctx context.Context, name string, input json.RawMessage) *request {
	return &request{Kind: "handler", Name: name, Param: input, Context: ctx}
}

func newHookRequest(hookName string, record *skydb.Record, originalRecord *skydb.Record, ctx context.Context) *request {
	param := hookRequest{
		Record:   (*skyconv.JSONRecord)(record),
		Original: (*skyconv.JSONRecord)(originalRecord),
	}
	return &request{Kind: "hook", Name: hookName, Param: param, Context: ctx}
}

func newAuthRequest(authReq *skyplugin.AuthRequest) *request {
	return &request{
		Kind: "provider",
		Name: authReq.ProviderName,
		Param: struct {
			Action   string                 `json:"action"`
			AuthData map[string]interface{} `json:"auth_data"`
		}{authReq.Action, authReq.AuthData},
	}
}

// TODO(limouren): reduce copying of this method
func (req *request) MarshalJSON() ([]byte, error) {
	pluginCtx := skyplugin.ContextMap(req.Context)
	if rawParam, ok := req.Param.(json.RawMessage); ok {
		rawParamReq := struct {
			Kind    string                 `json:"kind"`
			Name    string                 `json:"name,omitempty"`
			Param   json.RawMessage        `json:"param,omitempty"`
			Context map[string]interface{} `json:"context,omitempty"`
		}{req.Kind, req.Name, rawParam, pluginCtx}
		return json.Marshal(&rawParamReq)
	}

	paramReq := struct {
		Kind    string                 `json:"kind"`
		Name    string                 `json:"name,omitempty"`
		Param   interface{}            `json:"param,omitempty"`
		Context map[string]interface{} `json:"context,omitempty"`
	}{req.Kind, req.Name, req.Param, pluginCtx}

	return json.Marshal(&paramReq)
}

func (p *zmqTransport) State() skyplugin.TransportState {
	return p.state
}

func (p *zmqTransport) SetInitHandler(f skyplugin.TransportInitHandler) {
	p.initHandler = f
}

func (p *zmqTransport) setState(state skyplugin.TransportState) {
	if state != p.state {
		oldState := p.state
		p.state = state
		p.logger.Infof("Transport state changes from %v to %v.", oldState, p.state)
	}
}

func (p *zmqTransport) RequestInit() {
	for {
		address := <-p.broker.freshWorkers

		if p.state != skyplugin.TransportStateUninitialized {
			// Although the plugin is only initialized once, we need
			// to clear the channel buffer so that broker doesn't get stuck
			continue
		}

		p.logger.Debugf("zmq transport got fresh worker %s", string(address))

		// TODO: Only send init to the new address. For now, we let
		// the broker decide.
		out, err := p.RunInit()
		if p.initHandler != nil {
			handlerError := p.initHandler(out, err)
			if err != nil || handlerError != nil {
				p.setState(skyplugin.TransportStateError)
			}
		}
		p.setState(skyplugin.TransportStateReady)
	}
}

func (p *zmqTransport) RunInit() (out []byte, err error) {
	param := struct {
		Config skyconfig.Configuration `json:"config"`
	}{p.config}
	req := request{Kind: "init", Param: param, Timeout: initRequestTimeout}
	for {
		out, err = p.ipc(&req)
		if err == nil {
			break
		}
		p.logger.WithField("err", err).Warnf(`zmq/rpc: Unable to send init request to plugin "%s". Retrying...`, p.name)
	}
	return
}

func (p *zmqTransport) RunLambda(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out, err = p.rpc(newLambdaRequest(ctx, name, in))
	return
}

func (p *zmqTransport) RunHandler(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out, err = p.rpc(newHandlerRequest(ctx, name, in))
	return
}

func (p *zmqTransport) RunHook(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
	out, err := p.rpc(newHookRequest(hookName, record, originalRecord, ctx))
	if err != nil {
		return nil, err
	}

	var recordout skydb.Record
	if err := json.Unmarshal(out, (*skyconv.JSONRecord)(&recordout)); err != nil {
		p.logger.WithField("data", string(out)).Error("failed to unmarshal record")
		return nil, fmt.Errorf("failed to unmarshal record: %v", err)
	}
	recordout.OwnerID = record.OwnerID
	recordout.CreatedAt = record.CreatedAt
	recordout.CreatorID = record.CreatorID
	recordout.UpdatedAt = record.UpdatedAt
	recordout.UpdaterID = record.UpdaterID

	return &recordout, nil
}

func (p *zmqTransport) RunTimer(name string, in []byte) (out []byte, err error) {
	req := request{Kind: "timer", Name: name}
	out, err = p.rpc(&req)
	return
}

func (p *zmqTransport) RunProvider(request *skyplugin.AuthRequest) (resp *skyplugin.AuthResponse, err error) {
	req := newAuthRequest(request)
	out, err := p.rpc(req)
	if err != nil {
		return
	}

	err = json.Unmarshal(out, &resp)
	return
}

func (p *zmqTransport) rpc(req *request) (out []byte, err error) {
	var rawResp []byte

	rawResp, err = p.ipc(req)
	if err != nil {
		return
	}

	var resp struct {
		Result json.RawMessage   `json:"result"`
		Err    *common.ExecError `json:"error"`
	}

	if err = json.Unmarshal(rawResp, &resp); err != nil {
		return
	}
	if resp.Err != nil {
		err = resp.Err
		return
	}

	out = resp.Result
	return
}

func (p *zmqTransport) ipc(req *request) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.WithField("recovered", r).
				Errorln("panic occurred while calling plugin ipc")
			log.Errorf("%s", debug.Stack())
			return
		}
	}()
	var (
		in      []byte
		reqSock *goczmq.Sock
	)

	in, err = json.Marshal(req)
	if err != nil {
		return
	}

	reqSock, err = goczmq.NewReq(p.iaddr)
	if err != nil {
		return
	}
	defer func() {
		reqSock.Destroy()
	}()
	if req.Timeout > 0 {
		reqSock.SetRcvtimeo(req.Timeout)
	}
	err = reqSock.SendMessage([][]byte{in})
	if err != nil {
		return
	}

	msg, err := reqSock.RecvMessage()
	if err != nil {
		return
	}

	if len(msg) != 1 {
		err = fmt.Errorf("malformed resp msg = %s", msg)
	} else {
		out = msg[0]
	}

	return
}

type zmqTransportFactory struct {
}

func (f zmqTransportFactory) Open(name string, args []string, config skyconfig.Configuration) (transport skyplugin.Transport) {
	const internalAddrFmt = `inproc://%s`

	internalAddr := fmt.Sprintf(internalAddrFmt, name)
	externalAddr := args[0]

	broker, err := NewBroker(name, internalAddr, externalAddr)
	logger := log.WithFields(logrus.Fields{"plugin": name})
	if err != nil {
		logger.Panicf("Failed to init broker for zmq transport: %v", err)
	}

	p := zmqTransport{
		state:  skyplugin.TransportStateUninitialized,
		name:   name,
		iaddr:  internalAddr,
		broker: broker,
		logger: logger,
		config: config,
	}

	go func() {
		logger.Infof("Running zmq broker:\niaddr = %s\neaddr = %s", internalAddr, externalAddr)
		broker.Run()
	}()

	return &p
}

func init() {
	// A new zmq socket is created here so that a new zmq context is created
	// for the process.
	// If the zmq context is not created first, it will be created at the
	// time when zmq sockets are created, which might cause zsys_init to
	// fail since it is not thread-safe.
	// goczmq does not provide function to init context, so we create
	// a new socket here and throw it away.
	router, err := goczmq.NewRouter(`inproc://init`)
	if err != nil {
		panic("unable to initialize zmq")
	}
	defer router.Destroy()

	// Register the zmq transport factory
	skyplugin.RegisterTransport("zmq", zmqTransportFactory{})
}
