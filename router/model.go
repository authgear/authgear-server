package router

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/oursky/ourd/asset"
	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/oddb"
)

// Payload is for passing payload to the actual handler
type Payload struct {
	// Map of params such as Auth, TimeSteam, version
	Meta map[string]interface{}
	// Map of action payload
	Data map[string]interface{}
	// URL parameters
	Params     []string
	TokenStore authtoken.Store
	AssetStore asset.Store
	AppName    string
	UserInfoID string
	DBConn     oddb.Conn
	Database   oddb.Database
	UserInfo   *oddb.UserInfo
	// the raw http.Request of this payload
	// Think twice before accessing it
	Req *http.Request
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

// WriteEntity writes a value as response to a request. Currently it only
// writes JSON response.
func (resp *Response) WriteEntity(i interface{}) error {
	resp.written = true
	// hard code JSON write at the moment
	return json.NewEncoder(resp.writer).Encode(i)
}
