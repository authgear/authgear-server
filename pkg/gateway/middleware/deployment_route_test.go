package middleware

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/skygeario/skygear-server/pkg/gateway/model"
)

var mockRoutes = []*model.DeploymentRoute{
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
	routes       []*model.DeploymentRoute
	matchedRoute *model.DeploymentRoute
}{
	{
		"api root without trailing slash",
		"/api",
		mockRoutes,
		&model.DeploymentRoute{
			Path: "/api",
		},
	}, {
		"api root with trailing slash",
		"/api/",
		mockRoutes,
		&model.DeploymentRoute{
			Path: "/api/",
		},
	}, {
		"api path",
		"/api/user/1",
		mockRoutes,
		&model.DeploymentRoute{
			Path: "/api/",
		},
	}, {
		"random path match root",
		"/testing",
		mockRoutes,
		&model.DeploymentRoute{
			Path: "/",
		},
	}, {
		"web path",
		"/web/welcome",
		mockRoutes,
		&model.DeploymentRoute{
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
