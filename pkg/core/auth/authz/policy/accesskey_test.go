package policy

import (
	"net/http"
	"testing"

	"github.com/skygeario/skygear-server/pkg/core/handler/context"
	"github.com/skygeario/skygear-server/pkg/core/model"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDenyNoAccessKey(t *testing.T) {
	Convey("should return error if auth context has no access key", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AccessKeyType: model.NoAccessKey,
		}

		err := DenyNoAccessKey(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should not return error if auth context has api key ", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AccessKeyType: model.APIAccessKey,
		}

		err := DenyNoAccessKey(req, ctx)
		So(err, ShouldBeEmpty)
	})

	Convey("should not return error if auth context has master key ", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AccessKeyType: model.MasterAccessKey,
		}

		err := DenyNoAccessKey(req, ctx)
		So(err, ShouldBeEmpty)
	})
}

func TestRequireMasterKey(t *testing.T) {
	Convey("should return error if auth context has no access key", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AccessKeyType: model.NoAccessKey,
		}

		err := RequireMasterKey(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should return error if auth context has api key ", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AccessKeyType: model.APIAccessKey,
		}

		err := RequireMasterKey(req, ctx)
		So(err, ShouldNotBeEmpty)
	})

	Convey("should not return error if auth context has master key ", t, func() {
		req, _ := http.NewRequest("POST", "/", nil)
		ctx := context.AuthContext{
			AccessKeyType: model.MasterAccessKey,
		}

		err := RequireMasterKey(req, ctx)
		So(err, ShouldBeEmpty)
	})
}
