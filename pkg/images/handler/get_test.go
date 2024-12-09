package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"gopkg.in/h2non/gock.v1"

	"github.com/authgear/authgear-server/pkg/util/httproute"
	"github.com/authgear/authgear-server/pkg/util/vipsutil"
)

func TestGetHandler(t *testing.T) {
	Convey("GetHandler", t, func() {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		gock.Intercept()
		defer gock.Off()

		directorMaker := NewMockDirectorMaker(ctrl)
		vipsDaemon := NewMockVipsDaemon(ctrl)

		router := httproute.NewRouter()
		h := GetHandler{
			DirectorMaker: directorMaker,
			Logger:        GetHandlerLogger{logrus.NewEntry(logrus.New())},
			VipsDaemon:    vipsDaemon,
		}
		router.Add(ConfigureGetRoute(httproute.Route{}), &h)

		Convey("Ignore any non-200 response", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/profile", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})
			gock.New("http://localhost:3004").Reply(404)

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
			So(w.Result().Header.Get("Cache-Control"), ShouldEqual, "")
			So(gock.IsDone(), ShouldBeTrue)
		})

		Convey("return 404 for invalid options", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/invalid", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 404)
			So(w.Result().Header.Get("Cache-Control"), ShouldEqual, "")
			So(gock.IsDone(), ShouldBeTrue)
		})

		Convey("strip upstream headers", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/profile", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})
			vipsDaemon.EXPECT().Process(gomock.Any()).Times(1).Return(&vipsutil.Output{
				Data:          nil,
				FileExtension: "",
			}, nil)
			gock.New("http://localhost:3004").
				Reply(200).
				SetHeader("foobar", "42")

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
			So(w.Result().Header.Get("foobar"), ShouldBeEmpty)
			So(gock.IsDone(), ShouldBeTrue)
		})

		Convey("set headers", func() {
			r, _ := http.NewRequest("GET", "http://localhost:3004/_images/app/image.jpg/profile", nil)
			w := httptest.NewRecorder()

			directorMaker.EXPECT().MakeDirector(gomock.Any()).AnyTimes().Return(func(r *http.Request) {})
			vipsDaemon.EXPECT().Process(gomock.Any()).Times(1).Return(&vipsutil.Output{
				Data:          nil,
				FileExtension: ".jpeg",
			}, nil)
			gock.New("http://localhost:3004").
				Reply(200)

			router.HTTPHandler().ServeHTTP(w, r)
			So(w.Result().StatusCode, ShouldEqual, 200)
			So(w.Result().Header.Get("Content-Length"), ShouldEqual, "0")
			So(w.Result().Header.Get("Content-Type"), ShouldEqual, "image/jpeg")
			So(w.Result().Header.Get("Cache-Control"), ShouldEqual, "public, immutable, max-age=900")
			So(w.Result().ContentLength, ShouldEqual, 0)
			So(gock.IsDone(), ShouldBeTrue)
		})
	})
}
