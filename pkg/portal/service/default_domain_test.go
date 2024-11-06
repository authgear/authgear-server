package service

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/config"
	portalconfig "github.com/authgear/authgear-server/pkg/portal/config"
)

func TestDefaultDomainService(t *testing.T) {
	Convey("DefaultDomainService", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		domains := NewMockDefaultDomainDomainService(ctrl)

		s := DefaultDomainService{
			AppConfig: &portalconfig.AppConfig{},
			Domains:   domains,
		}

		Convey("GetLatestAppHost", func() {
			test := func(suffix string, appID string, expected string) {
				s.AppConfig.HostSuffix = suffix
				actual, err := s.GetLatestAppHost(appID)
				if expected == "" {
					So(errors.Is(err, ErrHostSuffixNotConfigured), ShouldBeTrue)
				} else {
					So(actual, ShouldEqual, expected)
				}
			}

			test("", "myapp", "")
			test(".localhost", "myapp", "myapp.localhost")
			test(".localhost:3100", "myapp", "myapp.localhost:3100")
		})

		Convey("CreateAllDefaultDomains", func() {
			Convey("HostSuffix only", func() {
				s.AppConfig.HostSuffix = ".localhost"

				domains.EXPECT().CreateDomain(gomock.Any(), "myapp", "myapp.localhost", true, false).Times(1)
				ctx := context.Background()
				err := s.CreateAllDefaultDomains(ctx, "myapp")
				So(err, ShouldBeNil)
			})

			Convey("HostSuffix and HostSuffixes are the same", func() {
				s.AppConfig.HostSuffix = ".localhost"
				s.AppHostSuffixes = config.AppHostSuffixes([]string{".localhost"})

				domains.EXPECT().CreateDomain(gomock.Any(), "myapp", "myapp.localhost", true, false).Times(1)
				ctx := context.Background()
				err := s.CreateAllDefaultDomains(ctx, "myapp")
				So(err, ShouldBeNil)
			})

			Convey("HostSuffix and HostSuffixes are different", func() {
				s.AppConfig.HostSuffix = ".localhost"
				s.AppHostSuffixes = config.AppHostSuffixes([]string{".local"})

				gomock.InOrder(
					domains.EXPECT().CreateDomain(gomock.Any(), "myapp", "myapp.localhost", true, false).Times(1),
					domains.EXPECT().CreateDomain(gomock.Any(), "myapp", "myapp.local", true, false).Times(1),
				)
				ctx := context.Background()
				err := s.CreateAllDefaultDomains(ctx, "myapp")
				So(err, ShouldBeNil)
			})
		})
	})
}
