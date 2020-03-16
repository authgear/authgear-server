package urlprefix

import (
	"net/http"
	"net/url"

	corehttp "github.com/skygeario/skygear-server/pkg/core/http"
)

type Provider struct {
	Prefix url.URL
}

func NewProvider(req *http.Request) Provider {
	if req == nil {
		return Provider{}
	}
	return Provider{url.URL{
		Host:   corehttp.GetHost(req),
		Scheme: corehttp.GetProto(req),
	}}
}

func (p Provider) Value() *url.URL {
	return &p.Prefix
}
