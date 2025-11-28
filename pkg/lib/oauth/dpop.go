package oauth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
)

type oauthDPoPChecker func(*dpop.DPoPProof) error

func checkDPoPWithClient(
	ctx context.Context,
	client *config.OAuthClientConfig,
	checker oauthDPoPChecker,
	errorLogger func(err error),
) error {
	maybeDpopProof := dpop.GetDPoPProof(ctx)
	dpopProof, err := maybeDpopProof.Get()
	if err != nil {
		if !client.DPoPDisabled {
			return err
		}
	}

	err = checker(dpopProof)
	if err != nil {
		errorLogger(err)
		if !client.DPoPDisabled {
			return err
		}
	}

	return nil
}
