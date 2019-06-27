package handler

import (
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

type testItem struct {
	reqURL     string
	forwardURL string
}

var tests = []struct {
	name         string
	matchedRoute config.DeploymentRoute
	tests        []testItem
}{
	{
		"http handler config without trailing slash",
		config.DeploymentRoute{
			Path: "/api",
			Type: model.DeploymentRouteTypeHTTPHandler,
			TypeConfig: map[string]interface{}{
				"backend_url": "http://backend-domain/",
				"target_path": "/backend_function_path",
			},
		},
		[]testItem{
			{
				"http://public-domain/api",
				"http://backend-domain/backend_function_path",
			},
			{
				"http://public-domain/api/",
				"http://backend-domain/backend_function_path",
			},
			{
				"http://public-domain/api/user",
				"http://backend-domain/backend_function_path",
			},
		},
	}, {
		"http handler config with trailing slash",
		config.DeploymentRoute{
			Path: "/api/",
			Type: model.DeploymentRouteTypeHTTPHandler,
			TypeConfig: map[string]interface{}{
				"backend_url": "http://backend-domain/",
				"target_path": "/backend_function_path",
			},
		},
		[]testItem{
			{
				"http://public-domain/api",
				"http://backend-domain/backend_function_path",
			},
			{
				"http://public-domain/api/",
				"http://backend-domain/backend_function_path",
			},
			{
				"http://public-domain/api/user",
				"http://backend-domain/backend_function_path",
			},
		},
	}, {
		"http service config without trailing slash",
		config.DeploymentRoute{
			Path: "/api",
			Type: model.DeploymentRouteTypeHTTPService,
			TypeConfig: map[string]interface{}{
				"backend_url": "http://backend-domain/",
			},
		},
		[]testItem{
			testItem{
				"http://public-domain/api",
				"http://backend-domain/",
			},
			testItem{
				"http://public-domain/api/",
				"http://backend-domain/",
			},
			testItem{
				"http://public-domain/api/user",
				"http://backend-domain/user",
			},
			testItem{
				"http://public-domain/api/user/",
				"http://backend-domain/user/",
			},
		},
	}, {
		"http service config with trailing slash",
		config.DeploymentRoute{
			Path: "/api/",
			Type: model.DeploymentRouteTypeHTTPService,
			TypeConfig: map[string]interface{}{
				"backend_url": "http://backend-domain/",
			},
		},
		[]testItem{
			{
				"http://public-domain/api",
				"http://backend-domain/",
			},
			{
				"http://public-domain/api/",
				"http://backend-domain/",
			},
			{
				"http://public-domain/api/user",
				"http://backend-domain/user",
			},
			{
				"http://public-domain/api/user/",
				"http://backend-domain/user/",
			},
		},
	},
	{
		"http service with root path",
		config.DeploymentRoute{
			Path: "/",
			Type: model.DeploymentRouteTypeHTTPService,
			TypeConfig: map[string]interface{}{
				"backend_url": "http://backend-domain/",
			},
		},
		[]testItem{
			{
				"http://public-domain/",
				"http://backend-domain/",
			},
			{
				"http://public-domain/api",
				"http://backend-domain/api",
			},
			{
				"http://public-domain/api/",
				"http://backend-domain/api/",
			},
			{
				"http://public-domain/api/user",
				"http://backend-domain/api/user",
			},
		},
	},
}

func TestGetForwardURL(t *testing.T) {
	Convey("Test getForwardURL", t, func(c C) {
		for _, test := range tests {
			Convey(test.name, func() {
				for _, perTest := range test.tests {
					reqURL, _ := url.Parse(perTest.reqURL)
					forwardURL, _ := getForwardURL(reqURL, test.matchedRoute)
					So(forwardURL.String(), ShouldEqual, perTest.forwardURL)
				}
			})
		}
	})
}
