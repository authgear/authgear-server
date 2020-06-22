package urlprefix

import (
	"net/http"
	"net/url"

	"github.com/skygeario/skygear-server/pkg/httputil"
)

type Provider struct {
	Prefix url.URL
}

func NewProvider(req *http.Request) Provider {
	if req == nil {
		return Provider{}
	}
	return Provider{url.URL{
		// FIXME: use ServerConfig
		Host:   httputil.GetHost(req, true),
		Scheme: httputil.GetProto(req, true),
	}}
}

func (p Provider) Value() *url.URL {
	return &p.Prefix
}
