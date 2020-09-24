package service_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
	"github.com/authgear/authgear-server/pkg/portal/service"
)

func TestAdminAPIServiceResolveHost(t *testing.T) {
	Convey("AdminAPIService.ResolveHost static", t, func() {
		svc := &service.AdminAPIService{
			AdminAPIConfig: &portalconfig.AdminAPIConfig{
				HostTemplate: "localhost:3002",
			},
		}

		host, err := svc.ResolveHost("does-not-matter")
		So(err, ShouldBeNil)
		So(host, ShouldEqual, "localhost:3002")
	})

	Convey("AdminAPIService.ResolveHost dns-label", t, func() {
		svc := &service.AdminAPIService{
			AdminAPIConfig: &portalconfig.AdminAPIConfig{
				HostTemplate: "{{ .AppID }}.example.com",
			},
		}

		host, err := svc.ResolveHost("myapp")
		So(err, ShouldBeNil)
		So(host, ShouldEqual, "myapp.example.com")
	})
}
