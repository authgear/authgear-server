package tutorial

import (
	"bytes"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/readcloserthunk"
	"github.com/authgear/authgear-server/pkg/util/resource"
)

func TestServiceDetectProgresses(t *testing.T) {
	Convey("Service DetectProgresses", t, func() {
		s := &Service{}

		test := func(r *resource.ResourceFile, data []byte, expected []Progress) {
			actual, err := s.DetectProgresses(r, data)
			So(err, ShouldBeNil)
			So(actual, ShouldResemble, expected)
		}

		Convey("create_application", func() {
			test(&resource.ResourceFile{
				Location: resource.Location{
					Path: "authgear.yaml",
				},
				ReadCloserThunk: readcloserthunk.Reader(bytes.NewReader([]byte(`id: test
http:
  public_origin: http://test
`))),
			}, []byte(`id: test
http:
  public_origin: http://test
oauth:
  clients:
  - client_id: test
    name: test
    redirect_uris:
    - http://test
`), []Progress{ProgressCreateApplication})
		})

		Convey("sso", func() {
			test(&resource.ResourceFile{
				Location: resource.Location{
					Path: "authgear.yaml",
				},
				ReadCloserThunk: readcloserthunk.Reader(bytes.NewReader([]byte(`id: test
http:
  public_origin: http://test
`))),
			}, []byte(`id: test
http:
  public_origin: http://test
identity:
  oauth:
    providers:
    - type: google
      alias: google
      client_id: google
`), []Progress{ProgressSSO})
		})

		Convey("customize_ui", func() {
			test(&resource.ResourceFile{
				Location: resource.Location{
					Path: "static/authgear-light-theme.css",
				},
			}, nil, []Progress{ProgressCustomizeUI})
		})

	})
}
