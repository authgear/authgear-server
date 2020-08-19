package model

type Identity struct {
	Meta
	Type   string                 `json:"type"`
	Claims map[string]interface{} `json:"claims"`
}
