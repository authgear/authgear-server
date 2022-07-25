package passkey

type Service struct{}

func (s *Service) VerifyAttestationResponse(attestationResponse []byte) (credentialID string, signCount int64, err error) {
	// TODO(passkey): verify attestation response
	panic("not implemented")
}

func (s *Service) ParseAssertionResponse(assertionResponse []byte) (credentialID string, signCount int64, err error) {
	// TODO(passkey): parse assertion response
	panic("not implemented")
}
