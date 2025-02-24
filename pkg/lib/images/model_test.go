package images_test

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/lib/images"
)

func TestFileMetadata(t *testing.T) {
	ctx := context.Background()
	Convey("EncodeFileMetaData and DecodeFileMetaData", t, func() {
		Convey("upload by user", func() {
			metadata := &images.FileMetadata{
				UserID:     "userID",
				UploadedBy: images.UploadedByTypeUser,
			}

			encoded, err := images.EncodeFileMetaData(metadata)
			So(err, ShouldBeNil)

			decoded, err := images.DecodeFileMetadata(ctx, encoded)
			So(err, ShouldBeNil)
			So(decoded, ShouldResemble, metadata)
		})

		Convey("upload by admin API", func() {
			metadata := &images.FileMetadata{
				UploadedBy: images.UploadedByTypeAdminAPI,
			}

			encoded, err := images.EncodeFileMetaData(metadata)
			So(err, ShouldBeNil)

			decoded, err := images.DecodeFileMetadata(ctx, encoded)
			So(err, ShouldBeNil)
			So(decoded, ShouldResemble, metadata)
		})

		Convey("failed with validation error", func() {
			metadata := &images.FileMetadata{}

			encoded, err := images.EncodeFileMetaData(metadata)
			So(err, ShouldBeNil)

			_, err = images.DecodeFileMetadata(ctx, encoded)
			So(err, ShouldBeError, `invalid file metadata:
<root>: required
  map[actual:<nil> expected:[uploaded_by] missing:[uploaded_by]]
<root>: required
  map[actual:<nil> expected:[user_id] missing:[user_id]]`)
		})
	})
}
