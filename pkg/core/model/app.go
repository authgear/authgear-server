package model

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type AccessKeyType string

const (
	// NoAccessKeyType means no correct access key
	NoAccessKeyType AccessKeyType = "no"
	// APIAccessKeyType means request is using api key
	APIAccessKeyType AccessKeyType = "api-key"
	// MasterAccessKeyType means request is using master key
	MasterAccessKeyType AccessKeyType = "master-key"
)

func (t AccessKeyType) IsValid() bool {
	switch t {
	case NoAccessKeyType, APIAccessKeyType, MasterAccessKeyType:
		return true
	default:
		return false
	}
}

type AccessKey struct {
	Type     AccessKeyType
	ClientID string
}

func (k AccessKey) IsNoAccessKey() bool {
	return k.Type == NoAccessKeyType
}

func (k AccessKey) IsMasterKey() bool {
	return k.Type == MasterAccessKeyType
}

func header(i interface{}) http.Header {
	switch i.(type) {
	case *http.Request:
		return (i.(*http.Request)).Header
	case http.ResponseWriter:
		return (i.(http.ResponseWriter)).Header()
	default:
		panic("Invalid type")
	}
}

func GetAccessKey(i interface{}) AccessKey {
	accessKeyType := AccessKeyType(header(i).Get(coreHttp.HeaderAccessKeyType))
	clientID := header(i).Get(coreHttp.HeaderClientID)

	if !accessKeyType.IsValid() {
		return AccessKey{Type: NoAccessKeyType}
	}

	return AccessKey{Type: accessKeyType, ClientID: clientID}
}

func SetAccessKey(i interface{}, k AccessKey) {
	header(i).Set(coreHttp.HeaderAccessKeyType, string(k.Type))
	header(i).Set(coreHttp.HeaderClientID, k.ClientID)
}

func GetAPIKey(i interface{}) string {
	return header(i).Get(coreHttp.HeaderAPIKey)
}

func GetClientConfig(c map[string]config.APIClientConfiguration, clientID string) (*config.APIClientConfiguration, bool) {
	for id, clientConfig := range c {
		if id == clientID {
			cc := clientConfig
			return &cc, true
		}
	}
	return nil, false
}

func NewAccessKey(clientID string) AccessKey {
	return AccessKey{
		Type:     APIAccessKeyType,
		ClientID: clientID,
	}
}

const httpHeaderAuthorization = "authorization"
const httpAuthzBearerScheme = "bearer"

func parseAuthorizationHeader(r *http.Request) (token string) {
	authorization := strings.SplitN(r.Header.Get(httpHeaderAuthorization), " ", 2)
	if len(authorization) != 2 {
		return
	}

	scheme := authorization[0]
	if strings.ToLower(scheme) != httpAuthzBearerScheme {
		return
	}

	return authorization[1]
}

var ErrTokenConflict = fmt.Errorf("tokens detected in different transports")

func GetAccessToken(r *http.Request) (token string, transport config.SessionTransportType, err error) {
	headerToken := parseAuthorizationHeader(r)
	if headerToken == "" {
		headerToken = r.Header.Get(coreHttp.HeaderAccessToken)
	}

	var cookieToken string
	cookie, err := r.Cookie(coreHttp.CookieNameSession)
	if err == nil {
		cookieToken = cookie.Value
	} else {
		cookieToken = ""
		err = nil
	}

	if cookieToken != "" && headerToken != "" {
		err = ErrTokenConflict
		return
	}

	if headerToken != "" {
		token, transport = headerToken, config.SessionTransportTypeHeader
	} else {
		token, transport = cookieToken, config.SessionTransportTypeCookie
	}
	return
}
