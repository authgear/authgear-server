package httputil

import (
	"net/http"
	"regexp"
	"strings"
)

var forwardedForRegex = regexp.MustCompile(`for=([^;]*)(?:[; ]|$)`)
var ipRegex = regexp.MustCompile(`^(?:(\d+\.\d+\.\d+\.\d+)|\[(.*)\])(?::\d+)?$`)

func GetIP(r *http.Request, trustProxy bool) (ip string) {
	remoteAddr := r.RemoteAddr
	forwardedFor := r.Header.Get("X-Forwarded-For")
	originalFor := r.Header.Get("X-Original-For")
	realIP := r.Header.Get("X-Real-IP")
	forwarded := r.Header.Get("Forwarded")

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

	if trustProxy && forwarded != "" {
		if matches := forwardedForRegex.FindStringSubmatch(forwarded); len(matches) > 0 {
			ip = matches[1]
			return
		}
	}

	if trustProxy && forwardedFor != "" {
		parts := strings.Split(forwardedFor, ",")
		ip = parts[0]
		return
	}

	if trustProxy && originalFor != "" {
		parts := strings.Split(originalFor, ",")
		ip = parts[0]
		return
	}

	if trustProxy && realIP != "" {
		ip = realIP
		return
	}

	ip = remoteAddr
	return
}
