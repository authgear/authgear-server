package middleware

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/core/config"
)

var mockRoutes = []config.DeploymentRoute{
	{
		Path: "/api",
	}, {
		Path: "/api/",
	}, {
		Path: "/web",
	}, {
		Path: "/",
	},
}

var tests = []struct {
	name         string
	reqPath      string
	routes       []config.DeploymentRoute
	matchedRoute *config.DeploymentRoute
}{
	{
		"api root without trailing slash",
		"/api",
		mockRoutes,
		&config.DeploymentRoute{
			Path: "/api",
		},
	},
	{
		"api root with trailing slash",
		"/api/",
		mockRoutes,
		&config.DeploymentRoute{
			Path: "/api/",
		},
	},
	{
		"api path",
		"/api/user/1",
		mockRoutes,
		&config.DeploymentRoute{
			Path: "/api/",
		},
	},
	{
		"random path match root",
		"/testing",
		mockRoutes,
		&config.DeploymentRoute{
			Path: "/",
		},
	},
	{
		"web path",
		"/web/welcome",
		mockRoutes,
		&config.DeploymentRoute{
			Path: "/web",
		},
	},
}

func TestGetForwardURL(t *testing.T) {
	Convey("Test findMatchedRoute", t, func(c C) {
		for _, test := range tests {
			Convey(test.name, func() {
				matchedRoute := findMatchedRoute(test.reqPath, test.routes)
				So(matchedRoute.Path, ShouldEqual, test.matchedRoute.Path)
			})
		}
	})
}
