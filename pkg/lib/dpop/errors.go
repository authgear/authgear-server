package dpop

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var InvalidDPoPProof = apierrors.BadRequest.WithReason("InvalidDPoPProof")

var ErrMalformedJwt = InvalidDPoPProof.New("malformed DPoP jwt")
var ErrInvalidJwt = InvalidDPoPProof.New("invalid DPoP jwt")
var ErrInvalidJwtType = InvalidDPoPProof.New("invalid DPoP jwt typ")
var ErrInvalidJwtPayload = InvalidDPoPProof.New("invalid DPoP jwt payload")
var ErrInvalidJwtSignature = InvalidDPoPProof.New("invalid DPoP jwt signature")
var ErrInvalidJwtNoJwkProvided = InvalidDPoPProof.New("jwk not provided in DPoP jwt")
var ErrInvalidJwk = InvalidDPoPProof.New("invalid DPoP jwk")
var ErrProofExpired = InvalidDPoPProof.New("DPoP proof expired")
var ErrInvalidHTU = InvalidDPoPProof.New("htu in the DPoP proof is not a valid URI")
var ErrUnmatchedMethod = InvalidDPoPProof.New("htm in the DPoP proof does not match request method")
var ErrUnmatchedURI = InvalidDPoPProof.New("htu in the DPoP proof does not match request uri")
