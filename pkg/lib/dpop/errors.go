package dpop

import (
	"fmt"

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

func NewErrUnmatchedJKT(msg string, expected *string, actual *string) *UnmatchedJKTError {
	expectedStr := "null"
	actualStr := "null"
	if expected != nil {
		expectedStr = *expected
	}
	if actual != nil {
		actualStr = *actual
	}

	return &UnmatchedJKTError{
		Message:     msg,
		ExpectedJKT: expectedStr,
		ActualJKT:   actualStr,
	}
}

type UnmatchedJKTError struct {
	Message     string
	ExpectedJKT string
	ActualJKT   string
}

var _ error = (*UnmatchedJKTError)(nil)
var _ errorutil.Detailer = (*UnmatchedJKTError)(nil)

func (e *UnmatchedJKTError) Error() string {
	return fmt.Sprintf("dpop: unmatched jkt. message: %s. expected: %s actual: %s", e.Message, e.ExpectedJKT, e.ActualJKT)
}

func (e *UnmatchedJKTError) FillDetails(d errorutil.Details) {
	d["expected_jkt"] = e.ExpectedJKT
	d["actual_jkt"] = e.ActualJKT
}
