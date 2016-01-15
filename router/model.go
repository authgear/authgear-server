package router

import (
	"bufio"
	"encoding/json"
	"net"
	"net/http"

	"github.com/oursky/skygear/skydb"
	"github.com/oursky/skygear/skyerr"
	"golang.org/x/net/context"
)

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
	UserInfoID string
	UserInfo   *skydb.UserInfo

	DBConn   skydb.Conn
	Database skydb.Database
}

func (p *Payload) NewPayload(req *http.Request) *Payload {
	return &Payload{
		Req:     req,
		Meta:    map[string]interface{}{},
		Data:    map[string]interface{}{},
		Context: context.Background(),
	}
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

// AccessToken return the user input string
// TODO: accept all header, json payload, query string(in order)
func (p *Payload) AccessToken() string {
	var token interface{}
	token = p.Data["access_token"]
	switch token := token.(type) {
	default:
		return ""
	case string:
		return token
	}
}

// Response is interface for handler to write response to router
type Response struct {
	Meta       map[string]interface{} `json:"-"`
	Info       interface{}            `json:"info,omitempty"`
	Result     interface{}            `json:"result,omitempty"`
	Err        skyerr.Error           `json:"error,omitempty"`
	RequestID  string                 `json:"request_id,omitempty"`
	DatabaseID string                 `json:"database_id,omitempty"`
	written    bool
	hijacked   bool
	writer     http.ResponseWriter
}

// Header returns the header map being written before return a response.
// Mutating the map after calling WriteEntity has no effects.
func (resp *Response) Header() http.Header {
	return resp.writer.Header()
}

func (resp *Response) WriteHeader(status int) {
	resp.writer.WriteHeader(status)
}

func (resp *Response) Hijack() (c net.Conn, w *bufio.ReadWriter, e error) {
	resp.hijacked = true
	hijacker := resp.writer.(http.Hijacker)
	c, w, e = hijacker.Hijack()
	return
}

// Write writes raw bytes as response to a request.
func (resp *Response) Write(b []byte) (int, error) {
	resp.written = true
	return resp.writer.Write(b)
}

// WriteEntity writes a value as response to a request. Currently it only
// writes JSON response.
func (resp *Response) WriteEntity(i interface{}) error {
	resp.written = true
	// hard code JSON write at the moment
	return json.NewEncoder(resp.writer).Encode(i)
}
