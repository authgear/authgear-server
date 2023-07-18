package service_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/service"
)

func TestAdminAPIServiceResolveHost(t *testing.T) {
	Convey("AdminAPIService.ResolveHost", t, func() {
		svc := &service.AdminAPIService{
			AppConfig: &portalconfig.AppConfig{
				HostSuffix: ".localhost:3002",
			},
		}

		host, err := svc.ResolveHost("myapp")
		So(err, ShouldBeNil)
		So(host, ShouldEqual, "myapp.localhost:3002")
	})
}
