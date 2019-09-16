package session

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/ua-parser/uap-go/uaparser"

	"github.com/skygeario/skygear-server/pkg/auth/model"
	"github.com/skygeario/skygear-server/pkg/core/auth"
)

func Format(session *auth.Session) (mSession model.Session) {
	mSession.ID = session.ID
	mSession.IdentityID = session.PrincipalID
	mSession.CreatedAt = session.CreatedAt
	mSession.LastAccessedAt = session.AccessedAt
	mSession.CreatedByIP = resolveIP(session.InitialAccess.Remote)
	mSession.LastAccessedByIP = resolveIP(session.LastAccess.Remote)
	mSession.UserAgent = parseUserAgent(session.LastAccess.UserAgent)
	mSession.UserAgent.DeviceName = session.LastAccess.Extra.DeviceName()
	return
}

var uaParser = uaparser.NewFromSaved()

var skygearUARegex = regexp.MustCompile(`^(.*)/(\d+)(?:\.(\d+)|)(?:\.(\d+)|)(?:\.(\d+)|) \(Skygear;`)

func parseUserAgent(ua string) (mUA model.SessionUserAgent) {
	mUA.Raw = ua

	client := uaParser.Parse(ua)
	if matches := skygearUARegex.FindStringSubmatch(ua); len(matches) > 0 {
		client.UserAgent.Family = matches[1]
		client.UserAgent.Major = matches[2]
		client.UserAgent.Minor = matches[3]
		client.UserAgent.Patch = matches[4]
	}

	if client.UserAgent.Family != "Other" {
		mUA.Name = client.UserAgent.Family
		mUA.Version = client.UserAgent.ToVersionString()
	}
	if client.Device.Family != "Other" {
		mUA.DeviceModel = fmt.Sprintf("%s %s", client.Device.Brand, client.Device.Model)
	}
	if client.Os.Family != "Other" {
		mUA.OS = client.Os.Family
		mUA.OSVersion = client.Os.ToVersionString()
	}

	return mUA
}

var forwardedForRegex = regexp.MustCompile(`for=([^;]*)(?:[; ]|$)`)
var ipRegex = regexp.MustCompile(`^(?:(\d+\.\d+\.\d+\.\d+)|\[(.*)\])(?::\d+)?$`)

func resolveIP(conn auth.SessionAccessEventConnInfo) (ip string) {
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
