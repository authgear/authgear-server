package model

import (
	"net/url"
	"path"
	"strings"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

const maxRouteAttempt = 5

type RouteMatch struct {
	Route config.DeploymentRoute
	Path  string
}

func matchRoutePath(reqPath string, routes []config.DeploymentRoute) *RouteMatch {
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

func MatchRoute(reqPath string, routes []config.DeploymentRoute) *RouteMatch {
	attempt := 0
	for attempt < maxRouteAttempt {
		attempt++

		match := matchRoutePath(reqPath, routes)
		if match == nil {
			return nil
		}
		if match.Route.Type == DeploymentRouteTypeStatic {
			config := RouteTypeConfig(match.Route.TypeConfig)
			pathMapping := config.AssetPathMapping()
			assetPath := path.Join("/", match.Path)
			indexFile := config.AssetIndexFile()
			if indexFile == "" {
				indexFile = "index.html"
			}

			var assetName string
			if n, ok := pathMapping[assetPath]; ok {
				assetName = n
			} else if n, ok := pathMapping[path.Join(assetPath, indexFile)]; ok {
				assetName = n
			} else if fallback := config.AssetFallbackPath(); fallback != "" {
				reqPath = fallback
				continue
			} else {
				return nil
			}
			match.Path = "/" + assetName
			return match
		}
		return match
	}

	panic("route_match: maximum routing attempt exceeded")
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
