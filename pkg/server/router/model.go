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

package router

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/skygeario/skygear-server/pkg/server/skydb"
	"github.com/skygeario/skygear-server/pkg/server/skyerr"
)

// TODO(cheungpat): We should put these values into a struct that is stored
// in context.Context, rather than storing individual values. Before this
// happens, this list should be keep in sync with the following locations:
// - pkg/server/plugin/hook.go
// - pkg/server/plugin/transport.go

// ContextKey is the name of the values that are saved to context.Context.
type ContextKey string

var UserIDContextKey ContextKey = "UserID"
var AccessKeyTypeContextKey ContextKey = "AccessKeyType"
var RequestIDContextKey ContextKey = "RequestID"

// HandlerFunc specifies the function signature of a request handler function
type HandlerFunc func(*Payload, *Response)

// Handler specifies the interface of a request handler
type Handler interface {
	Setup()
	GetPreprocessors() []Processor
	Handle(*Payload, *Response)
}

// Processor specifies the function signature for a Processor
type Processor interface {
	Preprocess(*Payload, *Response) int
}

type funcHandler struct {
	Func HandlerFunc
}

func (h *funcHandler) Setup() {
	return
}

func (h *funcHandler) GetPreprocessors() []Processor {
	return []Processor{}
}

func (h *funcHandler) Handle(payload *Payload, response *Response) {
	h.Func(payload, response)
}

// NewFuncHandler is intend for using in test, not actual code.
func NewFuncHandler(f HandlerFunc) Handler {
	return &funcHandler{f}
}

// AccessKeyType is the type of the access key specified in client request
//go:generate stringer -type=AccessKeyType
type AccessKeyType int

const (
	// NoAccessKey denotes that an access key is not specified
	NoAccessKey AccessKeyType = 0 + iota

	// ClientAccessKey denotes that a client access key is specified
	ClientAccessKey

	// MasterAccessKey denotes that a master aclieneccess key is specified
	MasterAccessKey
)

// AccessToken is an interface to access information about the Access Token
// in the payload.
type AccessToken interface {
	// IssuedAt returns the time when the access token is issued. If the
	// information is not available, the IsZero method of the
	// returned time is true.
	IssuedAt() time.Time
}

// Payload is for passing payload to the actual handler
type Payload struct {
	// the raw http.Request of this payload
	// Think twice before accessing it
	Req *http.Request
	// URL parameters
	Params []string

	// Map of params such as Auth, TimeSteam, version
	Meta map[string]interface{}
	// Map of action payload
	Data map[string]interface{}

	Context context.Context

	AppName    string
	AuthInfoID string
	AuthInfo   *skydb.AuthInfo
	AccessKey  AccessKeyType

	// AccessToken stores access token for this payload.
	//
	// The field is injected by preprocessor. The field
	// is nil if the AccessToken does not exist or is not valid.
	AccessToken AccessToken

	DBConn   skydb.Conn
	Database skydb.Database

	User *skydb.Record
}

// RouteAction must exist for every request
func (p *Payload) RouteAction() string {
	actionStr, _ := p.Data["action"].(string)
	return actionStr
}

// APIKey returns the api key in the request.
func (p *Payload) APIKey() string {
	key, _ := p.Data["api_key"].(string)
	return key
}

// AccessTokenString return the user input string
// TODO: accept all header, json payload, query string(in order)
func (p *Payload) AccessTokenString() string {
	var token interface{}
	token = p.Data["access_token"]
	switch token := token.(type) {
	default:
		return ""
	case string:
		return token
	}
}

// HasMasterKey returns whether the payload has master access key
func (p *Payload) HasMasterKey() bool {
	return p.AccessKey == MasterAccessKey
}

// Response is interface for handler to write response to router
type Response struct {
	Meta       map[string][]string `json:"-"`
	Info       interface{}         `json:"info,omitempty"`
	Result     interface{}         `json:"result,omitempty"`
	Err        skyerr.Error        `json:"error,omitempty"`
	RequestID  string              `json:"request_id,omitempty"`
	DatabaseID string              `json:"database_id,omitempty"`
	writer     http.ResponseWriter
	writerOnce sync.Once
}

func NewResponse(writer http.ResponseWriter) *Response {
	return &Response{
		writer: writer,
	}
}

// Writer returns a http.ResponseWriter only once. If a writer is already
// returned, this function will always return nil.
//
// Any http.Handler handling the request may get a http.ResponseWriter
// by calling this function. Once called, the caller is responsible for
// writing a response. No other Handler will be able to write to the
// same ResponseWriter.
func (resp *Response) Writer() (writer http.ResponseWriter) {
	resp.writerOnce.Do(func() {
		writer = resp.writer
	})
	return
}
