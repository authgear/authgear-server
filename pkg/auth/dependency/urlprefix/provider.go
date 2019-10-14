package urlprefix

import (
	"net/http"
	"net/url"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type Provider struct {
	prefix url.URL
}

func NewProvider(req *http.Request) Provider {
	if req == nil {
		return Provider{}
	}

	u := url.URL{
		Host:   corehttp.GetHost(req),
		Scheme: corehttp.GetProto(req),
	}
	return Provider{u}
}

func (p Provider) Value() *url.URL {
	return &p.prefix
}
