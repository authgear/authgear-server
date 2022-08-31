package httputil

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestFilesystemCache(t *testing.T) {
	Convey("Normally the make function is called once", t, func() {
		cache := NewFilesystemCache()

		callTime := 0

		var waitGroup sync.WaitGroup

		for i := 0; i < 10; i++ {
			r, _ := http.NewRequest("GET", "/a.json", nil)
			h := cache.Serve(r, func() ([]byte, error) {
				callTime++
				return []byte(`{"a": "b"}`), nil
			})
			w := httptest.NewRecorder()
			waitGroup.Add(1)
			go func() {
				h.ServeHTTP(w, r)
				waitGroup.Done()
			}()
		}

		waitGroup.Wait()
		So(callTime, ShouldEqual, 1)
	})

	Convey("The make function is called again if cache is cleared", t, func() {
		cache := NewFilesystemCache()

		callTime := 0

		serve := func() {
			r, _ := http.NewRequest("GET", "/a.json", nil)
			w := httptest.NewRecorder()
			h := cache.Serve(r, func() ([]byte, error) {
				callTime++
				return []byte(`{"a": "b"}`), nil
			})
			h.ServeHTTP(w, r)
		}

		serve()
		So(callTime, ShouldEqual, 1)

		So(cache.Clear(), ShouldBeNil)

		serve()
		So(callTime, ShouldEqual, 2)
	})
}
