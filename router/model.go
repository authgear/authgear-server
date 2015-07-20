package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/oursky/ourd/asset"
	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
	"github.com/oursky/ourd/provider"
)

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

	TokenStore       authtoken.Store
	AssetStore       asset.Store
	HookRegistry     *hook.Registry
	ProviderRegistry *provider.Registry

	AppName    string
	UserInfoID string
	UserInfo   *oddb.UserInfo

	DBConn   oddb.Conn
	Database oddb.Database
}

func (p *Payload) NewPayload(req *http.Request) *Payload {
	return &Payload{
		Req:  req,
		Meta: map[string]interface{}{},
		Data: map[string]interface{}{},
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

// IsAuth tell the middleware is this payload is an auth request
func (p *Payload) IsAuth() bool {
	defer func() {
		if r := recover(); r != nil {
			log.Println("IsAuth recover")
		}
		return
	}()
	return p.Data["action"].(string)[0:5] == "auth:"
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
	Meta        map[string]interface{} `json:"-"`
	Result      interface{}            `json:"result,omitempty"`
	Err         error                  `json:"error,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	DatabaseID  string                 `json:"database_id,omitempty"`
	OtherResult interface{}            `json:"other_result,omitempty"`
	written     bool
	writer      http.ResponseWriter
}

// Header returns the header map being written before return a response.
// Mutating the map after calling WriteEntity has no effects.
func (resp *Response) Header() http.Header {
	return resp.writer.Header()
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
