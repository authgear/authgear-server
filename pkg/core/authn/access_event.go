package authn

import (
	"regexp"
	"strings"
	"time"
)

var forwardedForRegex = regexp.MustCompile(`for=([^;]*)(?:[; ]|$)`)
var ipRegex = regexp.MustCompile(`^(?:(\d+\.\d+\.\d+\.\d+)|\[(.*)\])(?::\d+)?$`)

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

func (conn AccessEventConnInfo) IP() (ip string) {
	defer func() {
		ip = strings.TrimSpace(ip)
		// remove ports from IP
		if matches := ipRegex.FindStringSubmatch(ip); len(matches) > 0 {
			ip = matches[1]
			if len(matches[2]) > 0 {
				ip = matches[2]
			}
		}
	}()

	if conn.XRealIP != "" {
		ip = conn.XRealIP
		return
	}
	if conn.XForwardedFor != "" {
		parts := strings.SplitN(conn.XForwardedFor, ",", 2)
		ip = parts[0]
		return
	}
	if conn.Forwarded != "" {
		if matches := forwardedForRegex.FindStringSubmatch(conn.Forwarded); len(matches) > 0 {
			ip = matches[1]
			return
		}
	}
	ip = conn.RemoteAddr
	return
}

type AccessEventExtraInfo map[string]interface{}

func (i AccessEventExtraInfo) DeviceName() string {
	deviceName, _ := i["device_name"].(string)
	return deviceName
}
