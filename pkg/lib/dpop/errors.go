package dpop

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var InvalidDPoPProof = apierrors.BadRequest.WithReason("InvalidDPoPProof")

var ErrMalformedJwt = InvalidDPoPProof.New("malformed jwt")
var ErrInvalidJwt = InvalidDPoPProof.New("invalid jwt")
var ErrInvalidJwtType = InvalidDPoPProof.New("invalid jwt typ")
var ErrInvalidJwtPayload = InvalidDPoPProof.New("invalid jwt payload")
var ErrInvalidJwtSignature = InvalidDPoPProof.New("invalid jwt signature")
var ErrInvalidJwtNoJwkProvided = InvalidDPoPProof.New("jwk not provided in jwt")
var ErrInvalidJwk = InvalidDPoPProof.New("invalid jwk")
var ErrProofExpired = InvalidDPoPProof.New("proof expired")
var ErrInvalidHTU = InvalidDPoPProof.New("htu in the DPoP proof is not a valid URI")
var ErrUnmatchedMethod = InvalidDPoPProof.New("htm in the DPoP proof does not match request method")
var ErrUnmatchedURI = InvalidDPoPProof.New("htu in the DPoP proof does not match request uri")
