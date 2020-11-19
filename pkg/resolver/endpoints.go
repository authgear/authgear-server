package resolver

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/lib/config"
)

type EndpointsProvider struct {
	HTTP *config.HTTPConfig
}

func (p *EndpointsProvider) BaseURL() *url.URL {
	u, err := url.Parse(p.HTTP.PublicOrigin)
	if err != nil {
		panic(err)
	}
	return u
}
