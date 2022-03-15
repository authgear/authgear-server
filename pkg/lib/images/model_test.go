package images_test

import (
	"testing"

	"github.com/authgear/authgear-server/pkg/lib/images"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFileMetadata(t *testing.T) {
	Convey("EncodeFileMetaData and DecodeFileMetaData", t, func() {
		Convey("success", func() {
			metadata := &images.FileMetadata{
				UserID:     "userID",
				UploadedBy: images.UploadedByTypeUser,
			}

			encoded, err := images.EncodeFileMetaData(metadata)
			So(err, ShouldBeNil)

			decoded, err := images.DecodeFileMetadata(encoded)
			So(err, ShouldBeNil)
			So(decoded, ShouldResemble, metadata)
		})

		Convey("failed with validation error", func() {
			metadata := &images.FileMetadata{}

			encoded, err := images.EncodeFileMetaData(metadata)
			So(err, ShouldBeNil)

			_, err = images.DecodeFileMetadata(encoded)
			So(err, ShouldBeError, "invalid file metadata:\n<root>: required\n  map[actual:<nil> expected:[uploaded_by user_id] missing:[uploaded_by user_id]]")
		})
	})
}
