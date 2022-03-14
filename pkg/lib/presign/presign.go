package presign

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/util/clock"
	"github.com/authgear/authgear-server/pkg/util/duration"
	"github.com/authgear/authgear-server/pkg/util/httpsigning"
	"github.com/authgear/authgear-server/pkg/util/httputil"
	"github.com/authgear/authgear-server/pkg/util/jwkutil"
)

// PresignPutExpires is how long the presign PUT request remains valid.
const PresignPutExpires time.Duration = 15 * duration.PerMinute

type Provider struct {
	Secret *config.ImagesKeyMaterials
	Clock  clock.Clock
	Host   httputil.HTTPHost
}

func (p *Provider) PresignPostRequest(url *url.URL) error {
	if p.Secret == nil {
		return apierrors.NewInternalError("missing images secret")
	}

	key, err := jwkutil.ExtractOctetKey(p.Secret.Set, "")
	if err != nil {
		return fmt.Errorf("presign: %w", err)
	}
	now := p.Clock.NowUTC()
	r := &http.Request{
		Method: "POST",
		URL:    url,
	}
	httpsigning.Sign(key, r, now, int(PresignPutExpires.Seconds()))
	*url = *r.URL
	return nil
}

func (p *Provider) Verify(r *http.Request) error {
	if p.Secret == nil {
		return apierrors.NewInternalError("missing images secret")
	}

	key, err := jwkutil.ExtractOctetKey(p.Secret.Set, "")
	if err != nil {
		return fmt.Errorf("presign: %w", err)
	}
	now := p.Clock.NowUTC()
	return httpsigning.Verify(key, p.Host, r, now)
}
