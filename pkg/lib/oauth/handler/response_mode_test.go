package handler

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestWriteResponse(t *testing.T) {
	Convey("writeResponse", t, func() {
		test := func(responseMode string, expected string) {
			w := httptest.NewRecorder()
			r, _ := http.NewRequest("GET", "/", nil)
			redirectURI, _ := url.Parse("https://example.com")
			response := map[string]string{
				"code":  "this_is_the_code",
				"state": "this_is_the_state",
			}
			writeResponse(w, r, redirectURI, responseMode, response)
			So(w.Body.String(), ShouldEqual, expected)
		}

		test("", `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url=https://example.com?code=this_is_the_code&amp;state=this_is_the_state" />
</head>
<body>
<script>
window.location.href = "https:\/\/example.com?code=this_is_the_code\x26state=this_is_the_state"
</script>
</body>
</html>
`)

		test("query", `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url=https://example.com?code=this_is_the_code&amp;state=this_is_the_state" />
</head>
<body>
<script>
window.location.href = "https:\/\/example.com?code=this_is_the_code\x26state=this_is_the_state"
</script>
</body>
</html>
`)

		test("fragment", `<!DOCTYPE html>
<html>
<head>
<meta http-equiv="refresh" content="0;url=https://example.com#code=this_is_the_code&amp;state=this_is_the_state" />
</head>
<body>
<script>
window.location.href = "https:\/\/example.com#code=this_is_the_code\x26state=this_is_the_state"
</script>
</body>
</html>
`)

		test("form_post", `<!DOCTYPE html>
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
<script>
document.forms[0].submit();
</script>
</body>
</html>
`)
	})
}
