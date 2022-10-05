package model

type Authorization struct {
	Meta

	ClientID string   `json:"clientID"`
	Scopes   []string `json:"scopes"`
}
