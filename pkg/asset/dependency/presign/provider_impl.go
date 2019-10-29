package presign

import (
	"net/http"
	"time"

	"github.com/skygeario/skygear-server/pkg/core/http/httpsigning"
	"github.com/skygeario/skygear-server/pkg/core/skyerr"
	coreTime "github.com/skygeario/skygear-server/pkg/core/time"
)

type providerImpl struct {
	secret       []byte
	timeProvider coreTime.Provider
}

func NewProvider(secret string, timeProvider coreTime.Provider) Provider {
	return &providerImpl{
		secret:       []byte(secret),
		timeProvider: timeProvider,
	}
}

func (p *providerImpl) Presign(r *http.Request, expires time.Duration) {
	httpsigning.Sign(p.secret, r, p.timeProvider.NowUTC(), int(expires.Seconds()))
}

func (p *providerImpl) Verify(r *http.Request) error {
	err := httpsigning.Verify(p.secret, r, p.timeProvider.NowUTC())
	if err != nil {
		return skyerr.Invalid.WithReason("InvalidSignedURL").New(err.Error())
	}
	return nil
}
