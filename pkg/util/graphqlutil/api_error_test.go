package graphqlutil

import (
	"errors"
	"testing"

	"github.com/graphql-go/graphql/gqlerrors"
	. "github.com/smartystreets/goconvey/convey"
)

type stubOriginalErrorWrapper struct {
	inner error
}

func (w stubOriginalErrorWrapper) Error() string        { return "wrapper" }
func (w stubOriginalErrorWrapper) OriginalError() error { return w.inner }

func TestOriginalError(t *testing.T) {
	Convey("originalError", t, func() {
		Convey("stops at a *gqlerrors.Error whose OriginalError field is nil, rather than returning nil", func() {
			// This is the shape of graphql-go's own internally-constructed
			// errors -- a variable coercion failure, for one -- which never
			// wrap an underlying error. Descending into the nil field used
			// to make this function return nil, and the caller's
			// err.Error() would then panic on it.
			err := gqlerrors.NewError("Variable \"$input\" got invalid value", nil, "", nil, nil, nil)
			So(originalError(err), ShouldEqual, err)
			So(originalError(err), ShouldNotBeNil)
		})

		Convey("unwraps a chain of OriginalError() wrappers down to the root", func() {
			root := errors.New("root cause")
			err := stubOriginalErrorWrapper{inner: stubOriginalErrorWrapper{inner: root}}
			So(originalError(err), ShouldEqual, root)
		})

		Convey("returns nil unchanged", func() {
			So(originalError(nil), ShouldBeNil)
		})
	})
}
