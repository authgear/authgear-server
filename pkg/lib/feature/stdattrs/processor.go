package stdattrs

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/authgear/authgear-server/pkg/lib/authn/stdattrs"
	"github.com/authgear/authgear-server/pkg/lib/config"
)

type Processor interface {
	Process(input interface{}) (out interface{}, err error)
}

type ProcessorFactory struct {
	PictureAttrProcessor *PictureAttrProcessor
}

func (f *ProcessorFactory) ProcessorWithAttrKey(key string) Processor {
	if key == stdattrs.Picture {
		return f.PictureAttrProcessor
	}
	return nil
}

func NewPictureAttrProcessor(
	r *http.Request,
	appID config.AppID,
	ImagesCDNHost config.ImagesCDNHost,
) *PictureAttrProcessor {
	imagesHost := ImagesCDNHost
	if imagesHost == "" {
		imagesHost = config.ImagesCDNHost(r.Host)
	}

	return &PictureAttrProcessor{
		ImagesHost: imagesHost,
		AppID:      appID,
	}
}

type PictureAttrProcessor struct {
	ImagesHost config.ImagesCDNHost
	AppID      config.AppID
}

func (p *PictureAttrProcessor) Process(input interface{}) (out interface{}, err error) {
	out = input
	if str, ok := input.(string); ok {
		u, err := url.Parse(str)
		if err != nil {
			return nil, fmt.Errorf("invalid profile url: %s: %w", str, err)
		}
		if u.Scheme == "authgearimages" {
			parts := strings.Split(u.Path, "/")
			if len(parts) >= 2 {
				objectID := parts[1]
				return fmt.Sprintf(
					"https://%s/_images/%s/%s/profile",
					p.ImagesHost,
					p.AppID,
					objectID,
				), nil
			}
		}
	}
	return
}
