package dpop

import "github.com/authgear/authgear-server/pkg/lib/oauth/protocol"

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
