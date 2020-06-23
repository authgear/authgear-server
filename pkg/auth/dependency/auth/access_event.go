package auth

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var forwardedForRegex = regexp.MustCompile(`for=([^;]*)(?:[; ]|$)`)
var ipRegex = regexp.MustCompile(`^(?:(\d+\.\d+\.\d+\.\d+)|\[(.*)\])(?::\d+)?$`)

const HeaderSessionExtraInfo = "x-authgear-extra-info"

type AccessInfo struct {
	InitialAccess AccessEvent `json:"initial_access"`
	LastAccess    AccessEvent `json:"last_access"`
}

type AccessEvent struct {
	Timestamp time.Time            `json:"time"`
	RemoteIP  string               `json:"ip,omitempty"`
	UserAgent string               `json:"user_agent,omitempty"`
	Extra     AccessEventExtraInfo `json:"extra,omitempty"`
}

func NewAccessEvent(timestamp time.Time, req *http.Request, trustProxy bool) AccessEvent {
	remote := AccessEventConnInfo{
		RemoteAddr:    req.RemoteAddr,
		XForwardedFor: req.Header.Get("X-Forwarded-For"),
		XRealIP:       req.Header.Get("X-Real-IP"),
		Forwarded:     req.Header.Get("Forwarded"),
	}

	extra := AccessEventExtraInfo{}
	extraData, err := base64.StdEncoding.DecodeString(req.Header.Get(HeaderSessionExtraInfo))
	const extraDataSizeLimit = 1024
	if err == nil && len(extraData) <= extraDataSizeLimit {
		_ = json.Unmarshal(extraData, &extra)
	}

	return AccessEvent{
		Timestamp: timestamp,
		RemoteIP:  remote.IP(trustProxy),
		UserAgent: req.UserAgent(),
		Extra:     extra,
	}
}

type AccessEventConnInfo struct {
	RemoteAddr    string `json:"remote_addr,omitempty"`
	XForwardedFor string `json:"x_forwarded_for,omitempty"`
	XRealIP       string `json:"x_real_ip,omitempty"`
	Forwarded     string `json:"forwarded,omitempty"`
}

func (conn AccessEventConnInfo) IP(trustProxy bool) (ip string) {
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

	if trustProxy && conn.Forwarded != "" {
		if matches := forwardedForRegex.FindStringSubmatch(conn.Forwarded); len(matches) > 0 {
			ip = matches[1]
			return
		}
	}
	if trustProxy && conn.XForwardedFor != "" {
		parts := strings.SplitN(conn.XForwardedFor, ",", 2)
		ip = parts[0]
		return
	}
	if trustProxy && conn.XRealIP != "" {
		ip = conn.XRealIP
		return
	}
	ip = conn.RemoteAddr
	return
}

type AccessEventExtraInfo map[string]interface{}

func (i AccessEventExtraInfo) DeviceName() string {
	deviceName, _ := i["device_name"].(string)
	return deviceName
}
