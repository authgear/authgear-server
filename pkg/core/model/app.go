package model

import (
	"net/http"

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

func GetClientConfig(c []config.OAuthClientConfiguration, clientID string) (config.OAuthClientConfiguration, bool) {
	for _, clientConfig := range c {
		if clientConfig.ClientID() == clientID {
			cc := clientConfig
			return cc, true
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
