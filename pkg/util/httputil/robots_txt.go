package httputil

import (
	"net/http"
	"strconv"
)

// Disallow all crawlers
var robotsTxt = `User-agent: *
Disallow: /
`

func RobotsTXTHandler(rw http.ResponseWriter, r *http.Request) {
	body := []byte(robotsTxt)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Content-Length", strconv.Itoa(len(body)))
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(body)
}
