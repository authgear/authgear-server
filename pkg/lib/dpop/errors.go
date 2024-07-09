package dpop

import "github.com/authgear/authgear-server/pkg/api/apierrors"

var InvalidDPoPProof = apierrors.BadRequest.WithReason("InvalidDPoPProof")

var ErrMalformedJwt = InvalidDPoPProof.New("malformed jwt")
var ErrInvalidJwt = InvalidDPoPProof.New("invalid jwt")
var ErrInvalidJwtPayload = InvalidDPoPProof.New("invalid jwt payload")
var ErrInvalidJwtSignature = InvalidDPoPProof.New("invalid jwt signature")
var ErrInvalidJwtNoJwkProvided = InvalidDPoPProof.New("jwk not provided in jwt")
var ErrInvalidJwk = InvalidDPoPProof.New("invalid jwk")
var ErrProofExpired = InvalidDPoPProof.New("proof expired")
