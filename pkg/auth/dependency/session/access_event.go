package session

import "time"

type AccessEvent struct {
	Timestamp time.Time            `json:"time"`
	Remote    AccessEventConnInfo  `json:"remote,omitempty"`
	UserAgent string               `json:"user_agent,omitempty"`
	Extra     AccessEventExtraInfo `json:"extra,omitempty"`
}

type AccessEventConnInfo struct {
	RemoteAddr    string `json:"remote_addr,omitempty"`
	XForwardedFor string `json:"x_forwarded_for,omitempty"`
	XRealIP       string `json:"x_real_ip,omitempty"`
	Forwarded     string `json:"forwarded,omitempty"`
}

type AccessEventExtraInfo map[string]interface{}

func (i AccessEventExtraInfo) DeviceName() string {
	deviceName, _ := i["device_name"].(string)
	return deviceName
}
