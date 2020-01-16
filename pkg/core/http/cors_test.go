package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFixupCORSHeaders(t *testing.T) {
	Convey("FixupCORSHeaders", t, func() {
		Convey("return CORS headers to downstream", func() {
			downstream := httptest.NewRecorder()
			downstream.Header().Set("Vary", "Origin, Authorization")
			downstream.Header().Set("Access-Control-Allow-Origin", "*")
			downstream.Header().Set("Request-Id", "12345678")

			upstream := &http.Response{Header: http.Header{}}
			upstream.Header.Set("Accept", "application/json")
			upstream.Header.Set("Vary", "Accept")

			FixupCORSHeaders(downstream, upstream)

			So(downstream.Result().Header, ShouldResemble, http.Header{
				"Vary":                        []string{"Origin, Authorization"},
				"Access-Control-Allow-Origin": []string{"*"},
				"Request-Id":                  []string{"12345678"},
			})
		})
		Convey("allow upstream to manage CORS headers", func() {
			downstream := httptest.NewRecorder()
			downstream.Header().Set("Vary", "Origin, Authorization")
			downstream.Header().Set("Access-Control-Allow-Origin", "*")
			downstream.Header().Set("Request-Id", "12345678")

			upstream := &http.Response{Header: http.Header{}}
			upstream.Header.Set("Accept", "application/json")
			upstream.Header.Set("Vary", "Accept, Origin")
			upstream.Header.Set("Access-Control-Allow-Origin", "example.com")

			FixupCORSHeaders(downstream, upstream)

			So(downstream.Result().Header, ShouldResemble, http.Header{
				"Vary":       []string{"Authorization"},
				"Request-Id": []string{"12345678"},
			})
		})
	})
}
