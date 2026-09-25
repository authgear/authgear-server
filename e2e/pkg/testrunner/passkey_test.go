package testrunner

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-webauthn/webauthn/protocol"
)

// The generated credentials are checked with go-webauthn directly, the same
// library and version the server verifies with, so a failure here is a real
// failure rather than a disagreement between two implementations.

const (
	testRPID      = "app.authgeare2e.localhost"
	testOrigin    = "http://app.authgeare2e.localhost:4000"
	testChallenge = "wRlLbQPZQ6Qh0vYHsX5nJ4pQ3s8Kk2sVc9JbYy0bC1A"
)

func creationOptions() any {
	return map[string]any{
		"publicKey": map[string]any{
			"challenge": testChallenge,
			"rp":        map[string]any{"id": testRPID, "name": "Test"},
			"user":      map[string]any{"id": "dXNlcg", "name": "e2e", "displayName": "e2e"},
		},
	}
}

func requestOptions() any {
	return map[string]any{
		"publicKey": map[string]any{
			"challenge": testChallenge,
			"rpId":      testRPID,
		},
	}
}

func TestGeneratePasskeyAttestation(t *testing.T) {
	raw, err := GeneratePasskeyAttestation(creationOptions(), testOrigin)
	if err != nil {
		t.Fatalf("generate: %v", err)
	}

	parsed, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(raw))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}

	_, err = parsed.Verify(
		testChallenge,
		false, // verifyUser
		true,  // verifyUserPresence
		testRPID,
		[]string{testOrigin},
		[]string{testOrigin},
		protocol.TopOriginExplicitVerificationMode,
		nil,
		[]protocol.CredentialParameter{
			{Type: protocol.PublicKeyCredentialType, Algorithm: -7},
		},
	)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
}

func TestGeneratePasskeyAssertion(t *testing.T) {
	// Registration must run first: the assertion is signed with the key whose
	// public half the attestation carried.
	attestationRaw, err := GeneratePasskeyAttestation(creationOptions(), testOrigin)
	if err != nil {
		t.Fatalf("generate attestation: %v", err)
	}
	attestation, err := protocol.ParseCredentialCreationResponseBody(strings.NewReader(attestationRaw))
	if err != nil {
		t.Fatalf("parse attestation: %v", err)
	}
	credentialBytes := attestation.Response.AttestationObject.AuthData.AttData.CredentialPublicKey

	assertionRaw, err := GeneratePasskeyAssertion(requestOptions(), testOrigin)
	if err != nil {
		t.Fatalf("generate assertion: %v", err)
	}

	parsed, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(assertionRaw))
	if err != nil {
		t.Fatalf("parse assertion: %v", err)
	}

	err = parsed.Verify(
		testChallenge,
		testRPID,
		[]string{testOrigin},
		[]string{testOrigin},
		protocol.TopOriginExplicitVerificationMode,
		"",    // appID
		false, // verifyUser
		true,  // verifyUserPresence
		credentialBytes,
	)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
}

// The origin is the point of the whole exercise: it is supplied by the caller,
// ends up inside the signed client data, and is what the server checks.
func TestGeneratePasskeyAssertionRejectedAtWrongOrigin(t *testing.T) {
	if _, err := GeneratePasskeyAttestation(creationOptions(), testOrigin); err != nil {
		t.Fatalf("generate attestation: %v", err)
	}

	assertionRaw, err := GeneratePasskeyAssertion(requestOptions(), "http://evil.localhost:4000")
	if err != nil {
		t.Fatalf("generate assertion: %v", err)
	}

	var body map[string]any
	if err := json.Unmarshal([]byte(assertionRaw), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	parsed, err := protocol.ParseCredentialRequestResponseBody(strings.NewReader(assertionRaw))
	if err != nil {
		t.Fatalf("parse assertion: %v", err)
	}

	if got := parsed.Response.CollectedClientData.Origin; got != "http://evil.localhost:4000" {
		t.Fatalf("origin = %q, want the one supplied", got)
	}

	err = parsed.Response.CollectedClientData.Verify(
		testChallenge,
		protocol.AssertCeremony,
		[]string{testOrigin},
		[]string{testOrigin},
		protocol.TopOriginExplicitVerificationMode,
	)
	if err == nil {
		t.Fatal("expected the origin check to reject this, but it passed")
	}
	if !strings.Contains(strings.ToLower(err.Error()), "origin") {
		t.Fatalf("expected an origin error, got: %v", err)
	}
}
