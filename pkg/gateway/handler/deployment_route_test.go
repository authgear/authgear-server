package handler

import (
	"fmt"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

func TestGetForwardURL(t *testing.T) {
	Convey("Test getForwardURL", t, func(c C) {
		type testCase struct {
			url        string
			matchPath  string
			route      config.DeploymentRoute
			forwardURL string
		}

		tests := []testCase{
			{
				url:       "https://example.com/api/",
				matchPath: "/",
				route: config.DeploymentRoute{
					Type: "http-service",
					Path: "/api",
					TypeConfig: map[string]interface{}{
						"backend_url": "http://backend/",
					},
				},
				forwardURL: "http://backend/",
			},
			{
				url:       "https://example.com/api/login?user=test&password=1234",
				matchPath: "/login",
				route: config.DeploymentRoute{
					Type: "http-service",
					Path: "/api",
					TypeConfig: map[string]interface{}{
						"backend_url": "http://backend/",
					},
				},
				forwardURL: "http://backend/login?user=test&password=1234",
			},
			{
				url:       "https://example.com/api/login#form",
				matchPath: "/login",
				route: config.DeploymentRoute{
					Type: "http-service",
					Path: "/api",
					TypeConfig: map[string]interface{}{
						"backend_url": "http://backend/",
					},
				},
				forwardURL: "http://backend/login#form",
			},
			{
				url:       "https://example.com/login?test#form",
				matchPath: "/index-html",
				route: config.DeploymentRoute{
					Type: "static",
					Path: "/",
					TypeConfig: map[string]interface{}{
						"backend_url": "https://app.localhost/_asset/get",
						"asset_path_mapping": map[string]interface{}{
							"/index.html":              "index-html",
							"/assets/main.12345678.js": "main-js",
						},
						"asset_error_page_path": "/",
					},
				},
				forwardURL: "https://app.localhost/_asset/get/index-html?test#form",
			},
		}

		for _, test := range tests {
			Convey(fmt.Sprintf("%s -> %s", test.url, test.forwardURL), func() {
				for _, test := range tests {
					url, err := url.Parse(test.url)
					if err != nil {
						panic(err)
					}
					match := model.RouteMatch{Route: test.route, Path: test.matchPath}
					forwardURL, err := getForwardURL(url, match)
					So(err, ShouldBeNil)
					So(forwardURL.String(), ShouldEqual, test.forwardURL)
				}
			})
		}
	})
}
