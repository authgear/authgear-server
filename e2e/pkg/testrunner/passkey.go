package testrunner

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/fxamacker/cbor/v2"
)

// A software WebAuthn authenticator, enough to exercise passkey ceremonies over
// the Authentication Flow API.
//
// There is no browser here, so this stands in for one: it takes the options the
// server issued, produces the response a real authenticator would, and signs it.
// The origin is supplied by the caller rather than observed, which is the whole
// point — it is what lets a test assert that the server checks it.
//
// Attestation format is "none". The server asks for "direct" but does not
// require it, and an empty attestation statement keeps this to one key pair.

const (
	// Flags in authenticator data: user present, user verified, attested
	// credential data included.
	flagUserPresent        = 0x01
	flagUserVerified       = 0x04
	flagAttestedCredential = 0x40

	// Platform authenticators report 0 and never increment, so do the same:
	// the server compares this against the stored value.
	passkeySignCount = 0
)

type passkeyCredential struct {
	key          *ecdsa.PrivateKey
	credentialID []byte
}

// Credentials are held per relying party, as a real authenticator holds them.
//
// One per process would be wrong in two ways: credential IDs are unique across
// the whole database rather than per app, so two tests registering in parallel
// would collide, and an authenticator that handed the same credential to every
// site is not what is being modelled.
var (
	passkeyMutex       sync.Mutex
	passkeyCredentials = map[string]passkeyCredential{}
)

// getPasskeyCredential returns the credential for an rpID, creating it on first
// use. Registration and the assertions that follow it share one key pair, since
// the server stores the public key at registration and verifies against it.
func getPasskeyCredential(rpID string) passkeyCredential {
	passkeyMutex.Lock()
	defer passkeyMutex.Unlock()

	if existing, ok := passkeyCredentials[rpID]; ok {
		return existing
	}

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		panic(err)
	}

	credentialID := make([]byte, 32)
	if _, err := rand.Read(credentialID); err != nil {
		panic(err)
	}

	created := passkeyCredential{key: key, credentialID: credentialID}
	passkeyCredentials[rpID] = created

	return created
}

// coseKey encodes the public key as a COSE_Key, which is what goes inside
// attested credential data. The map keys are the COSE labels.
func coseKey(pub *ecdsa.PublicKey) ([]byte, error) {
	x := make([]byte, 32)
	y := make([]byte, 32)
	pub.X.FillBytes(x)
	pub.Y.FillBytes(y)

	return cbor.Marshal(map[int]any{
		1:  2,  // kty: EC2
		3:  -7, // alg: ES256
		-1: 1,  // crv: P-256
		-2: x,
		-3: y,
	})
}

// authenticatorData is rpIdHash ‖ flags ‖ signCount, optionally followed by
// attested credential data during registration.
func authenticatorData(rpID string, includeCredential bool) ([]byte, error) {
	rpIDHash := sha256.Sum256([]byte(rpID))

	flags := byte(flagUserPresent | flagUserVerified)
	if includeCredential {
		flags |= flagAttestedCredential
	}

	data := make([]byte, 0, 128)
	data = append(data, rpIDHash[:]...)
	data = append(data, flags)

	signCount := make([]byte, 4)
	binary.BigEndian.PutUint32(signCount, passkeySignCount)
	data = append(data, signCount...)

	if !includeCredential {
		return data, nil
	}

	credential := getPasskeyCredential(rpID)

	// AAGUID is all zeroes, which is what a platform authenticator reports.
	data = append(data, make([]byte, 16)...)

	credentialIDLength := make([]byte, 2)
	binary.BigEndian.PutUint16(credentialIDLength, uint16(len(credential.credentialID)))
	data = append(data, credentialIDLength...)
	data = append(data, credential.credentialID...)

	publicKey, err := coseKey(&credential.key.PublicKey)
	if err != nil {
		return nil, err
	}
	data = append(data, publicKey...)

	return data, nil
}

func clientDataJSON(ceremonyType string, challenge string, origin string) []byte {
	// Marshalled rather than formatted, so an origin containing a quote cannot
	// produce something that is still valid JSON.
	b, err := json.Marshal(map[string]any{
		"type":        ceremonyType,
		"challenge":   challenge,
		"origin":      origin,
		"crossOrigin": false,
	})
	if err != nil {
		panic(err)
	}
	return b
}

func b64(b []byte) string {
	return base64.RawURLEncoding.EncodeToString(b)
}

// publicKeyOf digs the `publicKey` object out of a creation_options or
// request_options value as it arrives from the flow response.
func publicKeyOf(options any) (map[string]any, error) {
	m, ok := options.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("passkey: options is %T, want an object", options)
	}
	publicKey, ok := m["publicKey"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("passkey: options has no publicKey object")
	}
	return publicKey, nil
}

func stringField(m map[string]any, key string) (string, error) {
	v, ok := m[key].(string)
	if !ok {
		return "", fmt.Errorf("passkey: %q is missing or not a string", key)
	}
	return v, nil
}

// GeneratePasskeyAttestation performs a registration ceremony against the given
// creation_options and returns the creation_response as JSON.
func GeneratePasskeyAttestation(options any, origin string) (string, error) {
	publicKey, err := publicKeyOf(options)
	if err != nil {
		return "", err
	}

	challenge, err := stringField(publicKey, "challenge")
	if err != nil {
		return "", err
	}

	rp, ok := publicKey["rp"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("passkey: creation options have no rp object")
	}
	rpID, err := stringField(rp, "id")
	if err != nil {
		return "", err
	}

	authData, err := authenticatorData(rpID, true)
	if err != nil {
		return "", err
	}

	attestationObject, err := cbor.Marshal(map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	})
	if err != nil {
		return "", err
	}

	credentialID := getPasskeyCredential(rpID).credentialID
	response, err := json.Marshal(map[string]any{
		"id":    b64(credentialID),
		"rawId": b64(credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientDataJSON("webauthn.create", challenge, origin)),
			"attestationObject": b64(attestationObject),
		},
		"clientExtensionResults": map[string]any{},
	})
	if err != nil {
		return "", err
	}

	return string(response), nil
}

// GeneratePasskeyAssertion performs an authentication ceremony against the given
// request_options and returns the assertion_response as JSON. It signs with the
// key registered by GeneratePasskeyAttestation, so that must have run first.
func GeneratePasskeyAssertion(options any, origin string) (string, error) {
	publicKey, err := publicKeyOf(options)
	if err != nil {
		return "", err
	}

	challenge, err := stringField(publicKey, "challenge")
	if err != nil {
		return "", err
	}
	rpID, err := stringField(publicKey, "rpId")
	if err != nil {
		return "", err
	}

	authData, err := authenticatorData(rpID, false)
	if err != nil {
		return "", err
	}

	clientData := clientDataJSON("webauthn.get", challenge, origin)
	clientDataHash := sha256.Sum256(clientData)

	// The signature covers the authenticator data concatenated with the hash of
	// the client data — which is how the origin ends up being authenticated
	// even though the authenticator never sees it.
	signed := sha256.Sum256(append(append([]byte{}, authData...), clientDataHash[:]...))

	credential := getPasskeyCredential(rpID)
	signature, err := ecdsa.SignASN1(rand.Reader, credential.key, signed[:])
	if err != nil {
		return "", err
	}

	response, err := json.Marshal(map[string]any{
		"id":    b64(credential.credentialID),
		"rawId": b64(credential.credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    b64(clientData),
			"authenticatorData": b64(authData),
			"signature":         b64(signature),
		},
		"clientExtensionResults": map[string]any{},
	})
	if err != nil {
		return "", err
	}

	return string(response), nil
}
