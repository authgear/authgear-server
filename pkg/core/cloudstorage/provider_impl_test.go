package cloudstorage

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/h2non/gock"

	. "github.com/smartystreets/goconvey/convey"
)

func TestProvider(t *testing.T) {
	Convey("Provider", t, func() {
		appID := "myapp"
		storage := &MockStorage{}
		p := NewProvider(appID, storage)

		Convey("PresignPutRequest", func() {
			gock.InterceptClient(http.DefaultClient)
			defer gock.Off()
			defer gock.RestoreClient(http.DefaultClient)

			u := &url.URL{
				Scheme: "http",
				Host:   "localhost",
				Path:   "/a",
			}
			storage.PutRequest = &http.Request{
				Method: "PUT",
				URL:    u,
			}
			storage.GetURL = u

			gock.New("http://localhost").
				Get("/a").Reply(200)
			gock.New("http://localhost").
				Head("/a").Reply(200)

			Convey("check duplicate if name is random", func() {
				_, err := p.PresignPutRequest(&PresignUploadRequest{
					Headers: map[string]interface{}{},
				})
				So(err, ShouldNotBeNil)
				So(err, ShouldBeError, "Duplicated: duplicate asset")
			})

			Convey("do not check duplicate if name is exact", func() {
				_, err := p.PresignPutRequest(&PresignUploadRequest{
					ExactName: "name",
					Headers:   map[string]interface{}{},
				})
				So(err, ShouldBeNil)
			})
		})
	})
}
