package router

import (
	"log"
	"net/http"

	"github.com/oursky/ourd/authtoken"
	"github.com/oursky/ourd/hook"
	"github.com/oursky/ourd/oddb"
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

	TokenStore   authtoken.Store
	HookRegistry *hook.Registry

	AppName    string
	UserInfoID string
	UserInfo   *oddb.UserInfo

	DBConn   oddb.Conn
	Database oddb.Database
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
}
