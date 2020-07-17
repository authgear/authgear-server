package webapp

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"image/png"
	"io"
	"net/http"
	"net/url"
	"strconv"

	"github.com/authgear/authgear-server/pkg/httproute"
	"github.com/authgear/authgear-server/pkg/otp"
)

func ConfigureKeyURIImageRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("OPTIONS", "GET").
		WithPathPattern("/key_uri_image")
}

type KeyURIImageEndpoints interface {
	BaseURL() *url.URL
}

type KeyURIImageHandler struct {
	Endpoints KeyURIImageEndpoints
}

func (h *KeyURIImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// TODO(mfa): Do not accept secret from URL query
	// because intermediate proxy may store the secret.

	// TODO(mfa): Do not account name from URL query.
	// Read from interaction flow state instead.

	issuer := h.Endpoints.BaseURL().String()
	accountName := r.Form.Get("account_name")
	secret := r.Form.Get("secret")

	if accountName == "" {
		http.Error(w, "missing account_name", http.StatusBadRequest)
		return
	}

	if secret == "" {
		http.Error(w, "missing secret", http.StatusBadRequest)
		return
	}

	key, err := otp.MakeTOTPKey(otp.MakeTOTPKeyOptions{
		Issuer:      issuer,
		AccountName: accountName,
		Secret:      secret,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	img, err := key.Image(512, 512)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	buf := &bytes.Buffer{}
	hash := sha256.New()
	tee := io.MultiWriter(buf, hash)
	png.Encode(tee, img)
	etag := hex.EncodeToString(hash.Sum(nil))

	w.Header().Set("content-type", "image/png")
	w.Header().Set("content-length", strconv.Itoa(buf.Len()))
	w.Header().Set("etag", etag)
	buf.WriteTo(w)
}
