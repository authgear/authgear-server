package model

import (
	"net/url"
	"path"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

type RouteMatch struct {
	Route config.DeploymentRoute
	Path  string
}

func MatchRoute(reqPath string, routes []config.DeploymentRoute) *RouteMatch {
	// convert path to /path/ format
	testReqPath := appendtrailingSlash(reqPath)

	var matchedRoute *config.DeploymentRoute
	matchedLen := 0
	for _, r := range routes {
		if reqPath == r.Path {
			matchedRoute = &r
			break
		}
		testPath := appendtrailingSlash(r.Path)
		if !strings.HasPrefix(testReqPath, testPath) {
			continue
		}
		if len(r.Path) > matchedLen {
			matchedLen = len(r.Path)
			// We must copy the route because r is
			// reused in all iterations.
			copied := r
			matchedRoute = &copied
		}
	}

	if matchedRoute == nil {
		return nil
	}

	routePath := strings.TrimSuffix(matchedRoute.Path, "/")
	matchPath := strings.TrimPrefix(reqPath, routePath)
	if len(matchPath) == 0 {
		matchPath = "/"
	}
	return &RouteMatch{
		Route: *matchedRoute,
		Path:  matchPath,
	}
}

func (m RouteMatch) ToURL(baseURL *url.URL) *url.URL {
	url := *baseURL
	url.Path = path.Join(url.Path, m.Path)
	// path.Join will clean the result and the returned path ends in a
	// slash only if it is the root "/".
	// check and add back the trailing slash if necessary
	if !strings.HasSuffix(url.Path, "/") {
		var needTrailingSlash bool
		if m.Path == "/" {
			needTrailingSlash = strings.HasSuffix(baseURL.Path, "/")
		} else {
			needTrailingSlash = strings.HasSuffix(m.Path, "/")
		}
		if needTrailingSlash {
			url.Path = url.Path + "/"
		}
	}
	return &url
}

func appendtrailingSlash(path string) string {
	if !strings.HasSuffix(path, "/") {
		return path + "/"
	}
	return path
}
