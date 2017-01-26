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
	"fmt"
	"runtime/debug"

	"github.com/Sirupsen/logrus"

	skyplugin "github.com/skygeario/skygear-server/pkg/server/plugin"
	"github.com/skygeario/skygear-server/pkg/server/plugin/common"
	pluginrequest "github.com/skygeario/skygear-server/pkg/server/plugin/request"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"github.com/zeromq/goczmq"
)

type zmqTransport struct {
	state       skyplugin.TransportState
	name        string
	broker      *Broker
	initHandler skyplugin.TransportInitHandler
	logger      *logrus.Entry
	config      skyconfig.Configuration
}

func (p *zmqTransport) State() skyplugin.TransportState {
	return p.state
}

func (p *zmqTransport) SetState(state skyplugin.TransportState) {
	if state != p.state {
		oldState := p.state
		p.state = state
		p.logger.Infof("Transport state changes from %v to %v.", oldState, p.state)
	}
}

func (p *zmqTransport) SendEvent(name string, in []byte) ([]byte, error) {
	return p.rpc(pluginrequest.NewEventRequest(name, in))
}

func (p *zmqTransport) RunLambda(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out, err = p.rpc(pluginrequest.NewLambdaRequest(ctx, name, in))
	return
}

func (p *zmqTransport) RunHandler(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out, err = p.rpc(pluginrequest.NewHandlerRequest(ctx, name, in))
	return
}

func (p *zmqTransport) RunHook(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
	out, err := p.rpc(pluginrequest.NewHookRequest(ctx, hookName, record, originalRecord))
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
	req := pluginrequest.Request{Kind: "timer", Name: name}
	out, err = p.rpc(&req)
	return
}

func (p *zmqTransport) RunProvider(request *skyplugin.AuthRequest) (resp *skyplugin.AuthResponse, err error) {
	req := pluginrequest.NewAuthRequest(request)
	out, err := p.rpc(req)
	if err != nil {
		return
	}

	err = json.Unmarshal(out, &resp)
	return
}

func (p *zmqTransport) rpc(req *pluginrequest.Request) (out []byte, err error) {
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
		logger := log.WithFields(logrus.Fields{
			"response-content": string(rawResp),
			"err":              err,
		})

		if reqContent, err := json.Marshal(req); err == nil {
			logger = logger.WithFields(logrus.Fields{
				"request-content": string(reqContent),
			})
		}

		logger.Errorln("Fail to unmarshal plugin response")
		err = fmt.Errorf("Failed to parse plugin response: %v", err)
		return
	}
	if resp.Err != nil {
		err = resp.Err
		return
	}

	out = resp.Result
	return
}

func (p *zmqTransport) ipc(req *pluginrequest.Request) (out []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.WithField("recovered", r).
				Errorln("panic occurred while calling plugin ipc")
			log.Errorf("%s", debug.Stack())
			return
		}
	}()

	in, err := json.Marshal(req)
	if err != nil {
		return
	}

	reqChan := make(chan chan []byte)
	p.broker.RPC(reqChan, in)
	respChan := <-reqChan
	// Broker will sent back a null byte if time out
	msg := <-respChan

	if bytes.Equal(msg, []byte{0}) {
		err = fmt.Errorf("Plugin time out")
	} else {
		out = msg
	}

	return
}

type zmqTransportFactory struct {
}

func (f zmqTransportFactory) Open(name string, args []string, config skyconfig.Configuration) (transport skyplugin.Transport) {
	externalAddr := args[0]
	broker, err := NewBroker(name, externalAddr, config.Zmq.Timeout)
	logger := log.WithFields(logrus.Fields{"plugin": name})
	if err != nil {
		logger.Panicf("Failed to init broker for zmq transport: %v", err)
	}

	p := zmqTransport{
		state:  skyplugin.TransportStateUninitialized,
		name:   name,
		broker: broker,
		logger: logger,
		config: config,
	}

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
