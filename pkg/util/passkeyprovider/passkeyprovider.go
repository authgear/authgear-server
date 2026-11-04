package passkeyprovider

import (
	_ "embed"
	"encoding/json"
	"strings"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/google/uuid"
)

// aaguid.json maps an AAGUID to its provider name. It is a names-only copy of
// the community list, see docs/specs/webauthn.md#passkey-provider.
//
//go:embed aaguid.json
var aaguidJSON []byte

type aaguidEntry struct {
	Name string `json:"name"`
}

var names map[string]string

func init() {
	names = load()
}

func load() map[string]string {
	var entries map[string]aaguidEntry
	err := json.Unmarshal(aaguidJSON, &entries)
	if err != nil {
		panic(err)
	}

	m := make(map[string]string, len(entries))
	for aaguid, entry := range entries {
		m[strings.ToLower(aaguid)] = entry.Name
	}
	return m
}

// AAGUIDFromAttestationResponse returns the AAGUID in the attested credential
// data of attestationResponse, formatted as a lowercase UUID.
func AAGUIDFromAttestationResponse(attestationResponse []byte) (string, error) {
	parsed, err := protocol.ParseCredentialCreationResponseBytes(attestationResponse)
	if err != nil {
		return "", err
	}

	aaguid, err := uuid.FromBytes(parsed.Response.AttestationObject.AuthData.AttData.AAGUID)
	if err != nil {
		return "", err
	}

	return aaguid.String(), nil
}

// Name returns the name of the provider that aaguid identifies.
// An authenticator that does not identify itself reports the all-zero AAGUID,
// which is never known.
func Name(aaguid string) (name string, ok bool) {
	name, ok = names[strings.ToLower(aaguid)]
	return
}
