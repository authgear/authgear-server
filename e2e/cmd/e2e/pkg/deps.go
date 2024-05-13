package e2e

import (
	"net/http"

	"github.com/google/wire"

	deps "github.com/authgear/authgear-server/pkg/lib/deps"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func ProvideEnd2EndHTTPRequest() *http.Request {
	r, _ := http.NewRequest("GET", "", nil)
	return r
}

func ProvideEnd2EndRemoteIP() httputil.RemoteIP {
	return httputil.RemoteIP("127.0.0.1")
}

func ProvideEnd2EndUserAgentString() httputil.UserAgentString {
	return httputil.UserAgentString("redis-queue")
}

func ProvideEnd2EndHTTPHost() httputil.HTTPHost {
	return httputil.HTTPHost("127.0.0.1")
}

func ProvideEnd2EndHTTPProto() httputil.HTTPProto {
	return httputil.HTTPProto("https")
}

var End2EndDependencySet = wire.NewSet(
	deps.AppRootDeps,
	ProvideEnd2EndHTTPRequest,
	ProvideEnd2EndRemoteIP,
	ProvideEnd2EndUserAgentString,
	ProvideEnd2EndHTTPHost,
	ProvideEnd2EndHTTPProto,
)
