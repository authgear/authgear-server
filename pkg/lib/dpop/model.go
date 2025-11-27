package dpop

import "net/url"

type MaybeDPoPProof interface {
	Get() (*DPoPProof, error)
}

// https://datatracker.ietf.org/doc/html/rfc9449#section-4.2
type DPoPProof struct {
	JTI string   `json:"jti"` // An unique identifier of the DPoP jwt
	HTM string   `json:"htm"` // The request method
	HTU *url.URL `json:"htu"` // The request uri

	// https://datatracker.ietf.org/doc/html/rfc9449#section-6.1
	JKT string `json:"jkt"` // base64url encoding of the JWK SHA-256 Thumbprint
}

var _ MaybeDPoPProof = (*DPoPProof)(nil)

func (p *DPoPProof) Get() (*DPoPProof, error) {
	return p, nil
}

type InvalidDPoPProofWithError struct {
	Error error
}

var _ MaybeDPoPProof = (*InvalidDPoPProofWithError)(nil)

func (p *InvalidDPoPProofWithError) Get() (*DPoPProof, error) {
	return nil, p.Error
}

type MissingDPoPProof struct {
}

var _ MaybeDPoPProof = (*MissingDPoPProof)(nil)

func (p *MissingDPoPProof) Get() (*DPoPProof, error) {
	return nil, nil
}
