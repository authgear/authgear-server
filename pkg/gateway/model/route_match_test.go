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
			{
				"match static assets",
				[]config.DeploymentRoute{
					{
						Type: "static",
						Path: "/",
						TypeConfig: map[string]interface{}{
							"asset_path_mapping": map[string]interface{}{
								"/index.html":       "index",
								"/login/index.html": "login-index",
							},
						},
					},
					{Type: "http-service", Path: "/api"},
				},
				[]testCase{
					{Path: "", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/api", MatchedRoutePath: "/api", MatchedPath: "/"},
					{Path: "/api/", MatchedRoutePath: "/api", MatchedPath: "/"},
					{Path: "/index.html", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/index.html/", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/login", MatchedRoutePath: "/", MatchedPath: "/login-index"},
					{Path: "/login/", MatchedRoutePath: "/", MatchedPath: "/login-index"},
					{Path: "/login/index.html", MatchedRoutePath: "/", MatchedPath: "/login-index"},
					{Path: "/login/index.html/index.html", MatchedRoutePath: "", MatchedPath: ""},
					{Path: "/sign-up", MatchedRoutePath: "", MatchedPath: ""},
				},
			},
			{
				"match static error page path",
				[]config.DeploymentRoute{
					{
						Type: "static",
						Path: "/",
						TypeConfig: map[string]interface{}{
							"asset_path_mapping": map[string]interface{}{
								"/index.html":              "index",
								"/assets/main.12345678.js": "main-js",
							},
							"asset_error_page_path": "/",
						},
					},
					{Type: "http-service", Path: "/api"},
				},
				[]testCase{
					{Path: "", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/api", MatchedRoutePath: "/api", MatchedPath: "/"},
					{Path: "/api/login", MatchedRoutePath: "/api", MatchedPath: "/login"},
					{Path: "/index.html", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/login.html", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/signup", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/assets", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/assets/main.12345678.js", MatchedRoutePath: "/", MatchedPath: "/main-js"},
					{Path: "/assets/main.12345678.js/no", MatchedRoutePath: "/", MatchedPath: "/index"},
				},
			},
			{
				"common SPA deployment",
				[]config.DeploymentRoute{
					{
						Type: "static",
						Path: "/",
						TypeConfig: map[string]interface{}{
							"asset_path_mapping": map[string]interface{}{
								"/index.html":  "index",
								"/favicon.ico": "icon",
							},
							"asset_error_page_path": "/",
						},
					},
					{
						Type: "static",
						Path: "/assets",
						TypeConfig: map[string]interface{}{
							"asset_path_mapping": map[string]interface{}{
								"/main.12345678.js": "main-js",
							},
						},
					},
				},
				[]testCase{
					{Path: "/", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/login", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/user/1", MatchedRoutePath: "/", MatchedPath: "/index"},
					{Path: "/favicon.ico", MatchedRoutePath: "/", MatchedPath: "/icon"},
					{Path: "/assets", MatchedRoutePath: "", MatchedPath: ""},
					{Path: "/assets/main.12345678.css", MatchedRoutePath: "", MatchedPath: ""},
					{Path: "/assets/main.12345678.js", MatchedRoutePath: "/assets", MatchedPath: "/main-js"},
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

		Convey("limit maximum routing attempt", func() {
			So(func() {
				MatchRoute("/index.html", []config.DeploymentRoute{{
					Type: "static",
					Path: "/",
					TypeConfig: map[string]interface{}{
						"asset_path_mapping":    map[string]interface{}{},
						"asset_error_page_path": "/index.html",
					},
				}})
			}, ShouldPanicWith, "route_match: maximum routing attempt exceeded")
		})
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
