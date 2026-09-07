package transport

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

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
