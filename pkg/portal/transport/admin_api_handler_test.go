package transport

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httproute"
)

// The exemption is decided on the *path route parameter, not on the request
// path, so this pins how a real request URL maps onto it. Reading the
// graphiQLProxiedPaths entries as if they were full request paths is an easy
// mistake to make.
func TestAdminAPIRoutePathParam(t *testing.T) {
	Convey("the *path route parameter", t, func() {
		pathOf := func(requestPath string) string {
			var got string
			router := httproute.NewRouter()
			router.Add(ConfigureAdminAPIRoute(httproute.Route{}), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = httproute.GetParam(r, "path")
			}))
			req, _ := http.NewRequest("GET", "http://portal.example.com"+requestPath, nil)
			router.HTTPHandler().ServeHTTP(httptest.NewRecorder(), req)
			return got
		}

		Convey("drops the /api/apps/:appid prefix", func() {
			So(pathOf("/api/apps/QXBwOmFjY291bnRz/graphql"), ShouldEqual, "/graphql")
			So(pathOf("/api/apps/QXBwOmFjY291bnRz/graphql?query=abc"), ShouldEqual, "/graphql")
			So(pathOf("/api/apps/QXBwOmFjY291bnRz/_api/admin/graphql"), ShouldEqual, "/_api/admin/graphql")
			So(pathOf("/api/apps/QXBwOmFjY291bnRz/_api/admin/images/upload"), ShouldEqual, "/_api/admin/images/upload")
		})

		Convey("so the GraphiQL document URLs are exempt and nothing else is", func() {
			So(isGraphiQLDocumentRequest("GET", pathOf("/api/apps/QXBwOmFjY291bnRz/graphql")), ShouldBeTrue)
			So(isGraphiQLDocumentRequest("GET", pathOf("/api/apps/QXBwOmFjY291bnRz/_api/admin/graphql")), ShouldBeTrue)
			So(isGraphiQLDocumentRequest("GET", pathOf("/api/apps/QXBwOmFjY291bnRz/_api/admin/images/upload")), ShouldBeFalse)
		})
	})
}

func TestIsGraphiQLDocumentRequest(t *testing.T) {
	Convey("isGraphiQLDocumentRequest", t, func() {
		Convey("exempts the GraphiQL document paths on GET", func() {
			So(isGraphiQLDocumentRequest("GET", "/graphql"), ShouldBeTrue)
			So(isGraphiQLDocumentRequest("GET", "/_api/admin/graphql"), ShouldBeTrue)
		})

		Convey("does not exempt GraphQL execution, which uses POST", func() {
			So(isGraphiQLDocumentRequest("POST", "/graphql"), ShouldBeFalse)
			So(isGraphiQLDocumentRequest("POST", "/_api/admin/graphql"), ShouldBeFalse)
		})

		Convey("does not exempt any other Admin API endpoint on GET", func() {
			// These endpoints mint presigned upload URLs and return user PII.
			// Before this was path-gated, a bare unauthenticated GET reached them.
			So(isGraphiQLDocumentRequest("GET", "/_api/admin/images/upload"), ShouldBeFalse)
			So(isGraphiQLDocumentRequest("GET", "/_api/admin/users/export/some-task-id"), ShouldBeFalse)
			So(isGraphiQLDocumentRequest("GET", "/_api/admin/users/import/some-job-id"), ShouldBeFalse)
		})

		Convey("does not exempt paths that merely look like the GraphiQL path", func() {
			So(isGraphiQLDocumentRequest("GET", "/graphql/../_api/admin/images/upload"), ShouldBeFalse)
			So(isGraphiQLDocumentRequest("GET", "/graphql/"), ShouldBeFalse)
			So(isGraphiQLDocumentRequest("GET", "/graphqlx"), ShouldBeFalse)
			So(isGraphiQLDocumentRequest("GET", "/_api/admin/graphql/subresource"), ShouldBeFalse)
			So(isGraphiQLDocumentRequest("GET", ""), ShouldBeFalse)
		})
	})
}
