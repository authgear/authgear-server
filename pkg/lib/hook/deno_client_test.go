package hook

import (
	"context"
	"net/http"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/h2non/gock"
)

func TestDenoClient(t *testing.T) {
	Convey("DenoClient", t, func() {
		endpoint := "http://localhost:8090"
		httpClient := &http.Client{}
		gock.InterceptClient(httpClient)
		defer gock.Off()

		denoClient := &DenoClientImpl{
			Endpoint:   endpoint,
			HTTPClient: httpClient,
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

		Convey("Check", func() {
			ctx := context.Background()

			gock.New("http://localhost:8090").
				Post("/check").
				JSON(map[string]interface{}{
					"script": "export default function(a) { return a }",
				}).
				Reply(200).
				JSON(map[string]interface{}{})
			defer func() { gock.Flush() }()

			err := denoClient.Check(ctx, "export default function(a) { return a }")
			So(err, ShouldBeNil)
		})

		Convey("Check with error", func() {
			ctx := context.Background()

			gock.New("http://localhost:8090").
				Post("/check").
				JSON(map[string]interface{}{
					"script": "syntax error",
				}).
				Reply(200).
				JSON(map[string]interface{}{
					"stderr": "error: The module's source code could not be parsed: Expected ';', '}' or <eof> at FILE:1:8",
				})
			defer func() { gock.Flush() }()

			err := denoClient.Check(ctx, "syntax error")
			So(err, ShouldBeError, "error: The module's source code could not be parsed: Expected ';', '}' or <eof> at FILE:1:8")
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
