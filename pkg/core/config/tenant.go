package config

import (
	"encoding/base64"
	"net/http"

	"github.com/kelseyhightower/envconfig"
)

// TenantConfiguration is a mock struct of tenant configuration
//go:generate msgp -tests=false
type TenantConfiguration struct {
	DBConnectionStr string `msg:"DATABASE_URL" envconfig:"DATABASE_URL" json:"DATABASE_URL"`
	APIKey          string `msg:"API_KEY" envconfig:"API_KEY" json:"API_KEY"`
	MasterKey       string `msg:"MASTER_KEY" envconfig:"MASTER_KEY" json:"MASTER_KEY"`
	AppName         string `msg:"APP_NAME" envconfig:"APP_NAME" json:"APP_NAME"`
	TokenStore      struct {
		Secret string `msg:"SECRET" envconfig:"TOKEN_STORE_SECRET" json:"SECRET"`
		Expiry int64  `msg:"EXPIRY" envconfig:"TOKEN_STORE_EXPIRY" json:"EXPIRY"`
	} `json:"TOKEN_STORE" msg:"TOKEN_STORE"`
}

func (c *TenantConfiguration) ReadFromEnv() error {
	return envconfig.Process("", c)
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
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic(err)
	}

	data, err = t.UnmarshalMsg(data)
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
	header(i).Set("X-Skygear-App-Config", base64.StdEncoding.EncodeToString(out))
}

// NewTenantConfigurationFromEnv implements ConfigurationProvider
func NewTenantConfigurationFromEnv(_ *http.Request) (TenantConfiguration, error) {
	c := TenantConfiguration{}
	err := envconfig.Process("", &c)
	return c, err
}
