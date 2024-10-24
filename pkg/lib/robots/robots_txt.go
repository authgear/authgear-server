package robots

import (
	"net/http"
	"strconv"
	"strings"
)

// Disallow all crawlers
var robotsTxt = strings.Trim(`
User-agent: *
Disallow: /
`, "\n")

type Handler struct{}

func (h *Handler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	body := []byte(robotsTxt)
	rw.Header().Set("Content-Type", "text/plain")
	rw.Header().Set("Content-Length", strconv.Itoa(len(body)))
	rw.WriteHeader(http.StatusOK)
	_, _ = rw.Write(body)
}
