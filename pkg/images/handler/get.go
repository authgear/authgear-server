package handler

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"mime"
	"net/http"
	"net/http/httputil"
	"strconv"

	imagesconfig "github.com/authgear/authgear-server/pkg/images/config"
	"github.com/authgear/authgear-server/pkg/util/httproute"
	utilhttputil "github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/slogutil"
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

var GetHandlerLogger = slogutil.NewLogger("get-handler")

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

// sniffLen is the number of leading bytes http.DetectContentType looks at.
const sniffLen = 512

// safeInlineMediaTypes are the media types the original variant may serve
// inline. Everything else is served as an opaque download.
//
// The uploaded object is arbitrary attacker-controlled bytes, and this endpoint
// is same origin with the Auth UI in the default deployment, so a response the
// browser is willing to treat as an active document is a stored XSS on the
// origin that holds the session cookie.
var safeInlineMediaTypes = map[string]struct{}{
	"image/jpeg":   {},
	"image/png":    {},
	"image/gif":    {},
	"image/webp":   {},
	"image/bmp":    {},
	"image/x-icon": {},
}

// bodyWithPrefix rejoins the bytes consumed for sniffing with the rest of the
// body, while still closing the original body.
type bodyWithPrefix struct {
	io.Reader
	io.Closer
}

type GetHandler struct {
	DirectorMaker DirectorMaker
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
			// #nosec G710 -- u.Scheme and u.Host are set from the server-configured ImagesCDNHost, not user input.
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
			ctx := r.Context()
			logger := GetHandlerLogger.GetLogger(ctx)
			logger.WithError(err).Error(ctx, "reverse proxy error")
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
				return h.modifyResponseOriginal(resp)
			case ImageVariantProfile:
				return h.modifyResponse(resp)
			default:
				return nil
			}
		},
	}
	reverseProxy.ServeHTTP(w, r)
}

// modifyResponseOriginal serves the stored object unmodified, but always under
// a Content-Type this handler chose.
//
// The header map has just been reset, so without this the response carries no
// Content-Type at all, and net/http sniffs the body and declares whatever it
// finds. For an uploaded HTML file that is text/html, which the browser then
// executes as a document on this origin. X-Content-Type-Options does not help,
// because it only stops the browser from overriding a declared type, and here
// the server is the one declaring it.
//
// The Content-Type recorded at upload is not trustworthy either: it comes from
// the client, so the media type is determined from the bytes.
func (h *GetHandler) modifyResponseOriginal(resp *http.Response) error {
	originalBody := resp.Body

	prefix := make([]byte, sniffLen)
	n, err := io.ReadFull(originalBody, prefix)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) {
		return err
	}
	prefix = prefix[:n]

	// The body is passed through untouched, so Content-Length still holds.
	resp.Body = bodyWithPrefix{
		Reader: io.MultiReader(bytes.NewReader(prefix), originalBody),
		Closer: originalBody,
	}

	mediaType, _, err := mime.ParseMediaType(http.DetectContentType(prefix))
	if err != nil {
		mediaType = ""
	}

	if _, ok := safeInlineMediaTypes[mediaType]; ok {
		resp.Header.Set("Content-Type", mediaType)
	} else {
		resp.Header.Set("Content-Type", "application/octet-stream")
		resp.Header.Set("Content-Disposition", "attachment")
	}

	return nil
}

func (h *GetHandler) modifyResponse(resp *http.Response) error {
	originalBody := resp.Body
	originalBytes, err := io.ReadAll(originalBody)
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
