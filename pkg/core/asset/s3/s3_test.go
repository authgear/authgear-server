package s3

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestS3Store(t *testing.T) {

	Convey("S3 Asset Store", t, func() {
		Convey("able to create bucket with us-west-1", func() {
			_, err := NewAssetStore(
				"access_token",
				"secret_token",
				"us-west-1",
				"bucket-name",
				"http://bucket-name.s3-website-us-west-1.amazonaws.com/",
				true,
			)
			So(err, ShouldBeNil)
			//So(s3.(s3Store).urlPrefix, ShouldEqual, "http://bucket-name.s3-website-us-east-2.amazonaws.com/")
		})

		Convey("able to create bucket with us-east-2", func() {
			_, err := NewAssetStore(
				"access_token",
				"secret_token",
				"us-east-2",
				"bucket-name",
				"http://bucket-name.s3-website-us-east-2.amazonaws.com/",
				true,
			)
			So(err, ShouldBeNil)
			//So(s3.(s3Store).urlPrefix, ShouldEqual, "http://bucket-name.s3-website-us-east-2.amazonaws.com/")
		})
	})
}
