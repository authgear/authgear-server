package endpoints

import (
	"net/url"

	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/google/wire"
)

var DependencySet = wire.NewSet(
	wire.Struct(new(OAuthEndpoints), "*"),
	wire.Struct(new(Endpoints), "*"),
)

func NewOAuthEndpoints(origin *url.URL) *OAuthEndpoints {
	return &OAuthEndpoints{
		HTTPHost:  httputil.HTTPHost(origin.Host),
		HTTPProto: httputil.HTTPProto(origin.Scheme),
	}
}
