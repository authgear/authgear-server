package dpop

// https://datatracker.ietf.org/doc/html/rfc9449#section-4.2
type DPoPProof struct {
	JTI string `json:"jti"` // An unique identifier of the DPoP jwt
	HTM string `json:"htm"` // The request method
	HTU string `json:"htu"` // The request uri

	// https://datatracker.ietf.org/doc/html/rfc9449#section-6.1
	JKT string `json:"jkt"` // base64url encoding of the JWK SHA-256 Thumbprint
}
