package hook

import (
	"context"
	"net/http"
	"testing"

	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"
)

func TestDenoClient(t *testing.T) {
	Convey("DenoClient", t, func() {
		endpoint := "http://localhost:8090"
		httpClient := &http.Client{}
		gock.InterceptClient(httpClient)
		defer gock.Off()

		denoClient := &DenoClient{
			Endpoint:   endpoint,
			HTTPClient: httpClient,
			Logger:     Logger{logrus.New().WithFields(logrus.Fields{})},
		}

		Convey("Run", func() {
			ctx := context.Background()

			gock.New("http://localhost:8090").
				Post("/run").
				JSON(map[string]interface{}{
					"script": "export default function(a) { return a }",
					"input":  42,
				}).
				Reply(200).
				JSON(map[string]interface{}{
					"output": 42,
				})
			defer func() { gock.Flush() }()

			actual, err := denoClient.Run(ctx, "export default function(a) { return a }", 42)
			So(err, ShouldBeNil)
			So(actual, ShouldEqual, 42)
		})

		Convey("invalid status code", func() {
			ctx := context.Background()

			gock.New("http://localhost:8090").
				Post("/run").
				JSON(map[string]interface{}{
					"script": "export default function(a) { return a }",
					"input":  42,
				}).
				Reply(500)
			defer func() { gock.Flush() }()

			_, err := denoClient.Run(ctx, "export default function(a) { return a }", 42)
			So(err, ShouldBeError, "invalid status code")
		})

		Convey("runtime error", func() {
			ctx := context.Background()

			gock.New("http://localhost:8090").
				Post("/run").
				JSON(map[string]interface{}{
					"script": "export default function(a) { return a }",
					"input":  42,
				}).
				Reply(200).
				JSON(map[string]interface{}{
					"error": "script error",
					"stderr": map[string]interface{}{
						"string": "undefined variable",
					},
				})
			defer func() { gock.Flush() }()

			_, err := denoClient.Run(ctx, "export default function(a) { return a }", 42)
			So(err, ShouldBeError, "script error")
		})
	})
}
