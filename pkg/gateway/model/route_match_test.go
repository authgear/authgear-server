package model

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

func TestRouteMatch(t *testing.T) {
	Convey("MatchRoute", t, func(c C) {
		type testCase struct {
			Path             string
			MatchedRoutePath string
			MatchedPath      string
		}

		var tests = []struct {
			name   string
			routes []config.DeploymentRoute
			cases  []testCase
		}{
			{
				"root path",
				[]config.DeploymentRoute{
					{Type: "http-service", Path: "/"},
				},
				[]testCase{
					{Path: "", MatchedRoutePath: "/", MatchedPath: "/"},
					{Path: "/", MatchedRoutePath: "/", MatchedPath: "/"},
					{Path: "/index.html", MatchedRoutePath: "/", MatchedPath: "/index.html"},
					{Path: "/api", MatchedRoutePath: "/", MatchedPath: "/api"},
					{Path: "/api/", MatchedRoutePath: "/", MatchedPath: "/api/"},
					{Path: "/api/login", MatchedRoutePath: "/", MatchedPath: "/api/login"},
					{Path: "/api/login/", MatchedRoutePath: "/", MatchedPath: "/api/login/"},
				},
			},
			{
				"match longest prefix",
				[]config.DeploymentRoute{
					{Type: "http-service", Path: "/"},
					{Type: "http-service", Path: "/api"},
				},
				[]testCase{
					{Path: "", MatchedRoutePath: "/", MatchedPath: "/"},
					{Path: "/", MatchedRoutePath: "/", MatchedPath: "/"},
					{Path: "/index.html", MatchedRoutePath: "/", MatchedPath: "/index.html"},
					{Path: "/api", MatchedRoutePath: "/api", MatchedPath: "/"},
					{Path: "/api/", MatchedRoutePath: "/api", MatchedPath: "/"},
					{Path: "/api/login", MatchedRoutePath: "/api", MatchedPath: "/login"},
					{Path: "/api/login/", MatchedRoutePath: "/api", MatchedPath: "/login/"},
				},
			},
		}

		for _, test := range tests {
			Convey(test.name, func() {
				for _, testCase := range test.cases {
					match := MatchRoute(testCase.Path, test.routes)

					matchedRoutePath := ""
					matchedPath := ""
					if match != nil {
						matchedRoutePath = match.Route.Path
						matchedPath = match.Path
					}

					So(matchedRoutePath, ShouldEqual, testCase.MatchedRoutePath)
					So(matchedPath, ShouldEqual, testCase.MatchedPath)
				}
			})
		}
	})

	Convey("RouteMatch.ToURL", t, func() {
		toURL := func(path, baseURL string) string {
			u, err := url.Parse(baseURL)
			if err != nil {
				panic(err)
			}
			m := RouteMatch{Path: path}
			return m.ToURL(u).String()
		}
		So(toURL("/", "http://backend"), ShouldEqual, "http://backend/")
		So(toURL("/", "http://backend/"), ShouldEqual, "http://backend/")
		So(toURL("/api", "http://backend"), ShouldEqual, "http://backend/api")
		So(toURL("/api/", "http://backend"), ShouldEqual, "http://backend/api/")
		So(toURL("/", "http://backend/api"), ShouldEqual, "http://backend/api")
		So(toURL("/", "http://backend/api/"), ShouldEqual, "http://backend/api/")
		So(toURL("/login", "http://backend/api"), ShouldEqual, "http://backend/api/login")
		So(toURL("/login/", "http://backend/api"), ShouldEqual, "http://backend/api/login/")
	})
}
