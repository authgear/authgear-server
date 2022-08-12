package identity

type PasskeySpec struct {
	AttestationResponse []byte `json:"attestation_response,omitempty"`
	AssertionResponse   []byte `json:"assertion_response,omitempty"`
}
