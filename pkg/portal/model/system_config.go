package model

type SystemConfig struct {
	AuthgearClientID string `json:"authgearClientID"`
	AuthgearEndpoint string `json:"authgearEndpoint"`
	AppHostSuffix    string `json:"appHostSuffix"`
}
