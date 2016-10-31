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

package http

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/logging"
	skyplugin "github.com/skygeario/skygear-server/pkg/server/plugin"
	"github.com/skygeario/skygear-server/pkg/server/plugin/common"
	pluginrequest "github.com/skygeario/skygear-server/pkg/server/plugin/request"
	"github.com/skygeario/skygear-server/pkg/server/skyconfig"
	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skydb/skyconv"
	"golang.org/x/net/context"
)

var log = logging.LoggerEntry("plugin")

type httpTransport struct {
	URL         string
	Path        string
	Args        []string
	initHandler skyplugin.TransportInitHandler
	state       skyplugin.TransportState
	httpClient  http.Client
	config      skyconfig.Configuration
}

func (p *httpTransport) rpc(req *pluginrequest.Request) (out []byte, err error) {
	data, err := p.ipc(req)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Result json.RawMessage   `json:"result"`
		Err    *common.ExecError `json:"error"`
	}

	jsonErr := json.Unmarshal(data, &resp)
	if jsonErr != nil {
		err = fmt.Errorf("failed to parse response: %v", jsonErr)
		return
	}

	if resp.Err != nil {
		err = resp.Err
		return
	}

	out = resp.Result
	return
}

func (p *httpTransport) ipc(req *pluginrequest.Request) (out []byte, err error) {
	in, err := json.Marshal(req)
	if err != nil {
		return
	}

	httpreq, err := http.NewRequest("POST", p.Path, bytes.NewReader(in))
	if err != nil {
		return
	}

	httpresp, err := p.httpClient.Do(httpreq)
	if err != nil {
		return nil, err
	}
	defer httpresp.Body.Close()

	return ioutil.ReadAll(httpresp.Body)
}

func (p *httpTransport) State() skyplugin.TransportState {
	return p.state
}

func (p *httpTransport) SetInitHandler(f skyplugin.TransportInitHandler) {
	p.initHandler = f
}

func (p *httpTransport) setState(state skyplugin.TransportState) {
	if state != p.state {
		oldState := p.state
		p.state = state
		log.Infof("Transport state changes from %v to %v.", oldState, p.state)
	}
}

func (p *httpTransport) RequestInit() {
	out, err := p.RunInit()
	if p.initHandler != nil {
		handlerError := p.initHandler(out, err)
		if err != nil || handlerError != nil {
			p.setState(skyplugin.TransportStateError)
			return
		}
	}
	p.setState(skyplugin.TransportStateReady)
}

func (p *httpTransport) RunInit() (out []byte, err error) {
	param := struct {
		Config skyconfig.Configuration `json:"config"`
	}{p.config}
	req := pluginrequest.Request{Kind: "init", Param: param}
	for {
		out, err = p.ipc(&req)
		if err == nil {
			return
		}
		time.Sleep(time.Second)
		log.WithField("err", err).
			Warnf(`http: Unable to send init request to plugin "%s". Retrying...`, p.Path)
	}
}

func (p *httpTransport) SendEvent(name string, in []byte) (out []byte, err error) {
	out, err = p.rpc(pluginrequest.NewEventRequest(name, in))
	return
}

func (p *httpTransport) RunLambda(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out, err = p.rpc(pluginrequest.NewLambdaRequest(ctx, name, in))
	return
}

func (p *httpTransport) RunHandler(ctx context.Context, name string, in []byte) (out []byte, err error) {
	out, err = p.rpc(pluginrequest.NewHandlerRequest(ctx, name, in))
	return
}

func (p *httpTransport) RunHook(ctx context.Context, hookName string, record *skydb.Record, originalRecord *skydb.Record) (*skydb.Record, error) {
	out, err := p.rpc(pluginrequest.NewHookRequest(ctx, hookName, record, originalRecord))
	if err != nil {
		return nil, err
	}

	var recordout skydb.Record
	if err := json.Unmarshal(out, (*skyconv.JSONRecord)(&recordout)); err != nil {
		log.WithField("data", string(out)).Error("failed to unmarshal record")
		return nil, fmt.Errorf("failed to unmarshal record: %v", err)
	}
	recordout.OwnerID = record.OwnerID
	recordout.CreatedAt = record.CreatedAt
	recordout.CreatorID = record.CreatorID
	recordout.UpdatedAt = record.UpdatedAt
	recordout.UpdaterID = record.UpdaterID

	return &recordout, nil
}

func (p *httpTransport) RunTimer(name string, in []byte) (out []byte, err error) {
	req := pluginrequest.Request{Kind: "timer", Name: name}
	out, err = p.rpc(&req)
	return
}

func (p *httpTransport) RunProvider(request *skyplugin.AuthRequest) (*skyplugin.AuthResponse, error) {
	req := pluginrequest.NewAuthRequest(request)
	out, err := p.rpc(req)

	resp := skyplugin.AuthResponse{}

	err = json.Unmarshal(out, &resp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &resp, nil
}

type httpTransportFactory struct {
}

func (f httpTransportFactory) Open(path string, args []string, config skyconfig.Configuration) (transport skyplugin.Transport) {
	transport = &httpTransport{
		Path:       path,
		Args:       args,
		state:      skyplugin.TransportStateUninitialized,
		httpClient: http.Client{},
		config:     config,
	}
	return
}

func init() {
	skyplugin.RegisterTransport("http", httpTransportFactory{})
}
