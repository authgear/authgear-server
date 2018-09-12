package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	coreMiddleware "github.com/skygeario/skygear-server/pkg/core/middleware"
	"github.com/skygeario/skygear-server/pkg/gateway/middleware"
	"github.com/skygeario/skygear-server/pkg/gateway/provider"

	"github.com/gorilla/mux"
)

var routerMap map[string]*url.URL

func init() {
	auth, _ := url.Parse("http://localhost:3000")
	routerMap = map[string]*url.URL{
		"auth": auth,
	}
}

func main() {
	r := mux.NewRouter()

	r.Use(coreMiddleware.TenantConfigurationMiddleware{
		ConfigurationProvider: coreMiddleware.ConfigurationProviderFunc(provider.NewTenantConfigurationFromRequest),
	}.Handle)
	r.Use(middleware.TenantAuthzMiddleware{}.Handle)

	proxy := NewReverseProxy()
	r.HandleFunc("/{gear}/{rest:.*}", rewriteHandler(proxy))

	srv := &http.Server{
		Addr: "0.0.0.0:3001",
		// Good practice to set timeouts to avoid Slowloris attacks.
		WriteTimeout: time.Second * 15,
		ReadTimeout:  time.Second * 15,
		IdleTimeout:  time.Second * 60,
		Handler:      r, // Pass our instance of gorilla/mux in.
	}

	fmt.Println("Start gateway server")
	if err := srv.ListenAndServe(); err != nil {
		log.Println(err)
	}
}

func NewReverseProxy() *httputil.ReverseProxy {
	director := func(req *http.Request) {
		path := req.URL.Path
		req.URL = routerMap[req.Header.Get("X-Skygear-Gear")]
		req.URL.Path = path
	}
	return &httputil.ReverseProxy{Director: director}
}

func rewriteHandler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Set("X-Skygear-Gear", mux.Vars(r)["gear"])
		r.URL.Path = "/" + mux.Vars(r)["rest"]
		p.ServeHTTP(w, r)
	}
}
