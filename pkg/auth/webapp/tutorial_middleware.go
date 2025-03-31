package webapp

import (
	"net/http"
)

//go:generate go tool mockgen -source=tutorial_middleware.go -destination=tutorial_middleware_mock_test.go -package webapp

type TutorialMiddlewareTutorialCookie interface {
	SetAll(rw http.ResponseWriter)
}

type TutorialMiddleware struct {
	TutorialCookie TutorialMiddlewareTutorialCookie
}

func (m *TutorialMiddleware) Handle(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		if q.Get("x_tutorial") == "true" {
			q.Del("x_tutorial")
			r.URL.RawQuery = q.Encode()
			m.TutorialCookie.SetAll(w)
		}

		next.ServeHTTP(w, r)
	})
}
