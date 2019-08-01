package model

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
	coreHttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type KeyType int

const (
	// NoAccessKey means no correct access key
	NoAccessKey KeyType = iota
	// APIAccessKey means request is using api key
	APIAccessKey
	// MasterAccessKey means request is using master key
	MasterAccessKey
)

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

func GetAccessKeyType(i interface{}) KeyType {
	ktv, err := strconv.Atoi(header(i).Get(coreHttp.HeaderAccesskeytype))
	if err != nil {
		return NoAccessKey
	}

	return KeyType(ktv)
}

func SetAccessKeyType(i interface{}, kt KeyType) {
	header(i).Set(coreHttp.HeaderAccesskeytype, strconv.Itoa(int(kt)))
}

func GetAPIKey(i interface{}) string {
	return header(i).Get(coreHttp.HeaderAPIKey)
}

func CheckAccessKeyType(config config.TenantConfiguration, apiKey string) KeyType {
	if apiKey == config.UserConfig.APIKey {
		return APIAccessKey
	}

	if apiKey == config.UserConfig.MasterKey {
		return MasterAccessKey
	}

	return NoAccessKey
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

func GetAccessToken(r *http.Request) string {
	token := parseAuthorizationHeader(r)
	if token != "" {
		return token
	}
	return r.Header.Get(coreHttp.HeaderAccessToken)
}
