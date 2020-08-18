package server

import (
	"fmt"
	"net/url"
	"strings"
)

func ParseListenAddress(addr string) (*url.URL, error) {
	if strings.HasPrefix(addr, "http://") || strings.HasPrefix(addr, "https://") {
		return url.Parse(addr)
	}
	return url.Parse(fmt.Sprintf("http://%s", addr))
}
