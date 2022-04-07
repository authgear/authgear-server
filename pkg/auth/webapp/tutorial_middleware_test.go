package webapp

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	. "github.com/smartystreets/goconvey/convey"
)

func TestTutorialMiddleware(t *testing.T) {
	Convey("TutorialMiddleware", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		tutorialCookie := NewMockTutorialMiddlewareTutorialCookie(ctrl)
		middleware := TutorialMiddleware{
			TutorialCookie: tutorialCookie,
		}

		makeHandler := func() http.Handler {
			dummy := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})
			h := middleware.Handle(dummy)
			return h
		}

		Convey("x_tutorial", func() {
			test := func(requestQuery string, resultQuery string, setCookie bool) {
				w := httptest.NewRecorder()
				r, _ := http.NewRequest("GET", requestQuery, nil)
				if setCookie {
					tutorialCookie.EXPECT().SetAll(w)
				}
				makeHandler().ServeHTTP(w, r)
				So(r.URL.RawQuery, ShouldEqual, resultQuery)
			}

			test("/", "", false)
			test("/?abc=test", "abc=test", false)
			test("/?x_tutorial=true&abc=test", "abc=test", true)
			test("/?abc=test&x_tutorial=true&state=test", "abc=test&state=test", true)
		})
	})
}
