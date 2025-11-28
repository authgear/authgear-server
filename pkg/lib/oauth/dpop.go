package oauth

import (
	"context"

	"github.com/authgear/authgear-server/pkg/lib/config"
	"github.com/authgear/authgear-server/pkg/lib/dpop"
)

type OAuthDPoPChecker func(*dpop.DPoPProof) *dpop.UnmatchedJKTError

func CheckDPoPWithClient(
	ctx context.Context,
	client *config.OAuthClientConfig,
	checker OAuthDPoPChecker,
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
