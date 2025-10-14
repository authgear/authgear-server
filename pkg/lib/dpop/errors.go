package dpop

import (
	"github.com/authgear/authgear-server/pkg/api/apierrors"
	"github.com/authgear/authgear-server/pkg/lib/oauth/protocol"
	"github.com/authgear/authgear-server/pkg/util/errorutil"
)

// From https://datatracker.ietf.org/doc/html/rfc9449#section-12.2
var InvalidDPoPProof = "invalid_dpop_proof"

func newInvalidDPoPProofError(msg string) error {
	return protocol.NewError(InvalidDPoPProof, msg)
}

var ErrMalformedJwt = newInvalidDPoPProofError("malformed DPoP jwt")
var ErrInvalidJwt = newInvalidDPoPProofError("invalid DPoP jwt")
var ErrInvalidJwtType = newInvalidDPoPProofError("invalid DPoP jwt typ")
var ErrInvalidJwtPayload = newInvalidDPoPProofError("invalid DPoP jwt payload")
var ErrInvalidJwtSignature = newInvalidDPoPProofError("invalid DPoP jwt signature")
var ErrInvalidJwk = newInvalidDPoPProofError("invalid DPoP jwk")
var ErrProofExpired = newInvalidDPoPProofError("DPoP proof expired")
var ErrInvalidHTU = newInvalidDPoPProofError("htu in the DPoP proof is not a valid URI")
var ErrUnmatchedMethod = newInvalidDPoPProofError("htm in the DPoP proof does not match request method")
var ErrUnmatchedURI = newInvalidDPoPProofError("htu in the DPoP proof does not match request uri")
var ErrUnsupportedAlg = newInvalidDPoPProofError("unsupported alg in DPoP jwt")

var UnmatchedJKT = apierrors.Invalid.WithReason("UnmatchedJKT")

func NewErrUnmatchedJKT(msg string, expected *string, actual *string) *apierrors.APIError {
	expectedStr := "null"
	actualStr := "null"
	if expected != nil {
		expectedStr = *expected
	}
	if actual != nil {
		actualStr = *actual
	}
	return UnmatchedJKT.NewWithInfo(msg, errorutil.Details{
		"expected": expectedStr,
		"actual":   actualStr,
	})
}
