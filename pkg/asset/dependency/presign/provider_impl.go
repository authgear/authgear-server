package presign

import (
	"net/http"

	"github.com/skygeario/skygear-server/pkg/core/http/httpsigning"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

type providerImpl struct {
	secret       []byte
	timeProvider coreTime.Provider
}

const (
	expires = 15 * 60
)

func NewProvider(secret string, timeProvider coreTime.Provider) Provider {
	return &providerImpl{
		secret:       []byte(secret),
		timeProvider: timeProvider,
	}
}

func (p *providerImpl) Presign(r *http.Request) {
	httpsigning.Sign(p.secret, r, p.timeProvider.NowUTC(), expires)
}

func (p *providerImpl) Verify(r *http.Request) error {
	return httpsigning.Verify(p.secret, r, p.timeProvider.NowUTC())
}
