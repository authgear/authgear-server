package stdattrs

import (
	"net/url"
	"path"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/httputil"
)

type Transformer interface {
	StorageFormToRepresentationForm(key string, value interface{}) (interface{}, error)
	RepresentationFormToStorageForm(key string, value interface{}) (interface{}, error)
}

type PictureTransformer struct {
	HTTPProto     httputil.HTTPProto
	HTTPHost      httputil.HTTPHost
	ImagesCDNHost config.ImagesCDNHost
}

var _ Transformer = &PictureTransformer{}

func (t *PictureTransformer) StorageFormToRepresentationForm(key string, value interface{}) (interface{}, error) {
	if key != stdattrs.Picture {
		return value, nil
	}

	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "authgearimages" {
		return value, nil
	}

	host := string(t.HTTPHost)
	if t.ImagesCDNHost != "" {
		host = string(t.ImagesCDNHost)
	}

	u = &url.URL{
		Scheme: string(t.HTTPProto),
		Host:   host,
		Path:   path.Join("_images", u.Path, "profile"),
	}
	return u.String(), nil
}

func (t *PictureTransformer) RepresentationFormToStorageForm(key string, value interface{}) (interface{}, error) {
	if key != stdattrs.Picture {
		return value, nil
	}

	str, ok := value.(string)
	if !ok {
		return value, nil
	}

	u, err := url.Parse(str)
	if err != nil {
		return nil, err
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return value, nil
	}

	if u.Host != string(t.HTTPHost) && u.Host != string(t.ImagesCDNHost) {
		return value, nil
	}

	parts := strings.Split(u.Path, "/")
	i := -1
	for j, part := range parts {
		if part == "_images" {
			i = j
		}
	}
	if i == -1 {
		return value, nil
	}

	p := path.Join("/", parts[i+1], parts[i+2])

	u = &url.URL{
		Scheme: "authgearimages",
		Path:   p,
	}
	return u.String(), nil
}
