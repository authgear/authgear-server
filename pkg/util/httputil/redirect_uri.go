package httputil

import (
	"errors"
	"net/http"
	"net/url"
)

func HostRelative(u *url.URL) *url.URL {
	uu := &url.URL{Path: "/"}
	if u.Path != "" {
		uu.Path = u.Path
	}
	return uu
}

func GetRedirectURI(r *http.Request, trustProxy bool) (out string, err error) {
	formRedirectURI := r.Form.Get("redirect_uri")
	queryRedirectURI := r.URL.Query().Get("redirect_uri")

	// Look at form body first
	if queryRedirectURI == "" && formRedirectURI != "" {
		out, err = parseRedirectURI(r, formRedirectURI, true, trustProxy)
		return
	}

	// Look at query then
	if queryRedirectURI != "" {
		out, err = parseRedirectURI(r, queryRedirectURI, false, trustProxy)
		return
	}

	err = errors.New("not found")
	return
}

func parseRedirectURI(r *http.Request, redirectURL string, allowRecursive bool, trustProxy bool) (out string, err error) {
	u, err := r.URL.Parse(redirectURL)
	if err != nil {
		return
	}

	recursive := u.Path == r.URL.Path || (u.RawPath != "" && u.RawPath == r.URL.RawPath)
	sameOrigin := (u.Scheme == "" && u.Host == "") ||
		(u.Scheme == GetProto(r, trustProxy) && u.Host == GetHost(r, trustProxy))

	if !sameOrigin {
		err = errors.New("not the same origin")
		return
	}

	if recursive && !allowRecursive {
		err = errors.New("recursive")
		return
	}

	out = u.String()
	return
}
