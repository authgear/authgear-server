package oauth

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"

	"github.com/authgear/authgear-server/pkg/util/httputil"
)

func TestWriteResponse(t *testing.T) {
	Convey("WriteResponse", t, func() {
		type testCase struct {
			ResponseMode       string
			ExpectedStatusCode int
			ExpectedHeaders    http.Header
			ExpectedBody       string
		}

		test := func(testCase testCase) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			r = r.WithContext(httputil.WithCSPNonce(r.Context(), "nonce"))
			redirectURI, _ := url.Parse("https://example.com")
			response := map[string]string{
				"code":  "this_is_the_code",
				"state": "this_is_the_state",
			}
			WriteResponse(w, r, redirectURI, testCase.ResponseMode, response)
			So(w.Body.String(), ShouldEqual, testCase.ExpectedBody)
			So(w.Result().StatusCode, ShouldEqual, testCase.ExpectedStatusCode)
			for k, expected := range testCase.ExpectedHeaders {
				actual := w.Header()[k]
				So(actual, ShouldEqual, expected)
			}
		}

		test(testCase{
			ResponseMode:       "",
			ExpectedStatusCode: 303,
			ExpectedHeaders: http.Header{
				"Location": []string{"https://example.com?code=this_is_the_code&state=this_is_the_state"},
			},
			ExpectedBody: `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url=https://example.com?code=this_is_the_code&amp;state=this_is_the_state" />
</head>
<body>
<script nonce="nonce">
window.location.href = "https:\/\/example.com?code=this_is_the_code\u0026state=this_is_the_state"
</script>
</body>
</html>
`,
		})

		test(testCase{
			ResponseMode:       "query",
			ExpectedStatusCode: 303,
			ExpectedHeaders: http.Header{
				"Location": []string{"https://example.com?code=this_is_the_code&state=this_is_the_state"},
			},
			ExpectedBody: `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url=https://example.com?code=this_is_the_code&amp;state=this_is_the_state" />
</head>
<body>
<script nonce="nonce">
window.location.href = "https:\/\/example.com?code=this_is_the_code\u0026state=this_is_the_state"
</script>
</body>
</html>
`,
		})

		test(testCase{
			ResponseMode:       "fragment",
			ExpectedStatusCode: 303,
			ExpectedHeaders: http.Header{
				"Location": []string{"https://example.com#code=this_is_the_code&state=this_is_the_state"},
			},
			ExpectedBody: `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url=https://example.com#code=this_is_the_code&amp;state=this_is_the_state" />
</head>
<body>
<script nonce="nonce">
window.location.href = "https:\/\/example.com#code=this_is_the_code\u0026state=this_is_the_state"
</script>
</body>
</html>
`,
		})

		test(testCase{
			ResponseMode:       "form_post",
			ExpectedStatusCode: 200,
			ExpectedBody: `<!DOCTYPE html>
<html>
<head>
<title>Submit this form</title>
</head>
<body>
<noscript>Please submit this form to proceed</noscript>
<form method="post" action="https://example.com">
<input type="hidden" name="code" value="this_is_the_code">
<input type="hidden" name="state" value="this_is_the_state">
<button type="submit" name="" value="">Submit</button>
</form>
<script nonce="nonce">
document.forms[0].submit();
</script>
</body>
</html>
`,
		})
	})
}
