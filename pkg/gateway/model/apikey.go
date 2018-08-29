package model

import (
	"net/http"
	"strconv"
)

type Key struct {
	APIKey    string
	MasterKey string
}

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
	return header(i).Get("X-Skygear-APIKey")
}

func GetAppName(i interface{}) string {
	return header(i).Get("X-Skygear-AppName")
}

func SetAppName(i interface{}, appName string) {
	header(i).Set("X-Skygear-AppName", appName)
}

func CheckAccessKeyType(app App, apiKey string) KeyType {
	if apiKey == app.Config.APIKey {
		return APIAccessKey
	}

	if apiKey == app.Config.MasterKey {
		return MasterAccessKey
	}

	return NoAccessKey
}
