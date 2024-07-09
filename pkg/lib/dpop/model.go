package dpop

type DPoPJwt struct {
	JTI string `json:"jti"` // An unique identifier of the DPoP jwt
	JKT string `json:"jkt"` // https://datatracker.ietf.org/doc/html/rfc9449#section-6.1
}
