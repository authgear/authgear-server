package handler

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime"
	"net/http"
	"net/http/httputil"
	"strconv"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/log"
	"github.com/authgear/authgear-server/pkg/util/vipsutil"
)

//go:generate go tool mockgen -source=get.go -destination=get_mock_test.go -package handler

func ConfigureGetRoute(route httproute.Route) httproute.Route {
	return route.
		WithMethods("GET", "OPTIONS").
		WithPathPattern("/_images/:appid/:objectid/:options")
}

func ExtractKey(r *http.Request) string {
	return fmt.Sprintf(
		"%s/%s",
		httproute.GetParam(r, "appid"),
		httproute.GetParam(r, "objectid"),
	)
}

type GetHandlerLogger struct{ *log.Logger }

func NewGetHandlerLogger(lf *log.Factory) GetHandlerLogger {
	return GetHandlerLogger{lf.New("get-handler")}
}

type VipsDaemon interface {
	Process(i vipsutil.Input) (*vipsutil.Output, error)
}

type ImageVariant string

const (
	ImageVariantOriginal ImageVariant = "original"
	ImageVariantProfile  ImageVariant = "profile"
)

func ParseImageVariant(s string) (ImageVariant, bool) {
	switch s {
	case string(ImageVariantOriginal):
		return ImageVariantOriginal, true
	case string(ImageVariantProfile):
		return ImageVariantProfile, true
	default:
		return "", false
	}
}

type DirectorMaker interface {
	MakeDirector(extractKey func(*http.Request) string) func(*http.Request)
}

type GetHandler struct {
	DirectorMaker DirectorMaker
	Logger        GetHandlerLogger
	ImagesCDNHost imagesconfig.ImagesCDNHost
	HTTPHost      utilhttputil.HTTPHost
	HTTPProto     utilhttputil.HTTPProto
	VipsDaemon    VipsDaemon
}

func (h *GetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.ImagesCDNHost != "" {
		if string(h.ImagesCDNHost) != string(h.HTTPHost) {
			u := *r.URL
			u.Scheme = string(h.HTTPProto)
			u.Host = string(h.ImagesCDNHost)
			http.Redirect(w, r, u.String(), http.StatusFound)
			return
		}
	}

	imageVariant, ok := ParseImageVariant(httproute.GetParam(r, "options"))
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	director := h.DirectorMaker.MakeDirector(ExtractKey)

	reverseProxy := httputil.ReverseProxy{
		Director: director,
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			h.Logger.WithError(err).Errorf("reverse proxy error")
			w.WriteHeader(http.StatusBadGateway)
		},
		ModifyResponse: func(resp *http.Response) error {
			// Reset the header so that we will not accidentally return any headers we do not support,
			// such as Accept-Ranges, X-Amz-Request-Id, etc.
			resp.Header = make(http.Header)

			// Do not modify response with unknown status code.
			if resp.StatusCode != 200 {
				return nil
			}

			switch imageVariant {
			case ImageVariantOriginal:
				return nil
			case ImageVariantProfile:
				return h.modifyResponse(resp)
			default:
				return nil
			}
		},
	}
	reverseProxy.ServeHTTP(w, r)
}

func (h *GetHandler) modifyResponse(resp *http.Response) error {
	originalBody := resp.Body
	originalBytes, err := ioutil.ReadAll(originalBody)
	if err != nil {
		return err
	}
	defer originalBody.Close()

	input := vipsutil.Input{
		Reader: bytes.NewReader(originalBytes),
		Options: vipsutil.Options{
			ResizingModeType: vipsutil.ResizingModeTypeCover,
			Width:            240,
			Height:           240,
		},
	}

	output, err := h.VipsDaemon.Process(input)
	if err != nil {
		return err
	}

	// Set Content-Length
	resp.ContentLength = int64(len(output.Data))
	resp.Header.Set("Content-Length", strconv.Itoa(len(output.Data)))

	// Set Content-Type
	mediaType := mime.TypeByExtension(output.FileExtension)
	if mediaType != "" {
		resp.Header.Set("Content-Type", mediaType)
	} else {
		resp.Header.Set("Content-Type", "application/octet-stream")
	}

	// Cache the response for 15 minutes.
	resp.Header.Set("Cache-Control", "public, immutable, max-age=900")

	resp.Body = io.NopCloser(bytes.NewReader(output.Data))
	return nil
}
