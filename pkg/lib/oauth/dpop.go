package oauth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
)

type oauthDPoPChecker func(*dpop.DPoPProof) *dpop.UnmatchedJKTError

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

	dpopErr := checker(dpopProof)
	if dpopErr != nil {
		errorLogger(dpopErr)
		if !client.DPoPDisabled {
			return err
		}
	}

	return nil
}
