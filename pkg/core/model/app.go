package model

import (
	"net/http"
	"strconv"

	"github.com/skygeario/skygear-server/pkg/core/config"
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
	ktv, err := strconv.Atoi(header(i).Get("X-Skygear-AccessKeyType"))
	if err != nil {
		return NoAccessKey
	}

	return KeyType(ktv)
}

func SetAccessKeyType(i interface{}, kt KeyType) {
	header(i).Set("X-Skygear-AccessKeyType", strconv.Itoa(int(kt)))
}

func GetAPIKey(i interface{}) string {
	return header(i).Get("X-Skygear-Api-Key")
}

func CheckAccessKeyType(config config.TenantConfiguration, apiKey string) KeyType {
	if apiKey == config.APIKey {
		return APIAccessKey
	}

	if apiKey == config.MasterKey {
		return MasterAccessKey
	}

	return NoAccessKey
}
