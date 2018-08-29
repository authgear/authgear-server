package config

import (
	"fmt"
	"net/http"
)

// TenantConfiguration is a mock struct of tenant configuration
//go:generate msgp -tests=false
type TenantConfiguration struct {
	DBConnectionStr string `msg:"DATABASE_URL"`
	APIKey          string `msg:"API_KEY"`
	MasterKey       string `msg:"MASRER_KEY"`
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

func GetTenantConfig(i interface{}) TenantConfiguration {
	s := header(i).Get("X-Skygear-App-Config")
	var t TenantConfiguration
	data := []byte(s)
	data, err := t.UnmarshalMsg(data)
	if err != nil {
		panic(err)
	}
	return t
}

func SetTenantConfig(i interface{}, t TenantConfiguration) {
	out, err := t.MarshalMsg(nil)
	if err != nil {
		panic(err)
	}
	header(i).Set("X-Skygear-App-Config", fmt.Sprintf("%s", out))
}
