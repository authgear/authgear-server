package passkeyprovider

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"testing"

	"github.com/go-webauthn/webauthn/protocol/webauthncbor"
	"github.com/google/uuid"
	. "github.com/smartystreets/goconvey/convey"
)

func makeAttestationResponse(aaguid string) []byte {
	rpIDHash := sha256.Sum256([]byte("example.com"))
	credentialID := []byte("credential-id")

	authData := append([]byte{}, rpIDHash[:]...)
	// User present, user verified, attested credential data included.
	authData = append(authData, 0x45)
	authData = append(authData, 0, 0, 0, 0)
	u := uuid.MustParse(aaguid)
	authData = append(authData, u[:]...)
	authData = append(authData, 0, byte(len(credentialID)))
	authData = append(authData, credentialID...)

	publicKey, err := webauthncbor.Marshal(map[int]any{
		1:  2,
		3:  -7,
		-1: 1,
		-2: make([]byte, 32),
		-3: make([]byte, 32),
	})
	if err != nil {
		panic(err)
	}
	authData = append(authData, publicKey...)

	attestationObject, err := webauthncbor.Marshal(map[string]any{
		"fmt":      "none",
		"attStmt":  map[string]any{},
		"authData": authData,
	})
	if err != nil {
		panic(err)
	}

	clientDataJSON, err := json.Marshal(map[string]any{
		"type":      "webauthn.create",
		"challenge": "challenge",
		"origin":    "https://example.com",
	})
	if err != nil {
		panic(err)
	}

	b, err := json.Marshal(map[string]any{
		"id":    base64.RawURLEncoding.EncodeToString(credentialID),
		"rawId": base64.RawURLEncoding.EncodeToString(credentialID),
		"type":  "public-key",
		"response": map[string]any{
			"clientDataJSON":    base64.RawURLEncoding.EncodeToString(clientDataJSON),
			"attestationObject": base64.RawURLEncoding.EncodeToString(attestationObject),
		},
	})
	if err != nil {
		panic(err)
	}
	return b
}

func TestAAGUIDFromAttestationResponse(t *testing.T) {
	Convey("AAGUIDFromAttestationResponse", t, func() {
		Convey("returns the AAGUID in the attested credential data", func() {
			aaguid, err := AAGUIDFromAttestationResponse(makeAttestationResponse("fbfc3007-154e-4ecc-8c0b-6e020557d7bd"))
			So(err, ShouldBeNil)
			So(aaguid, ShouldEqual, "fbfc3007-154e-4ecc-8c0b-6e020557d7bd")
		})

		Convey("returns the all-zero AAGUID", func() {
			aaguid, err := AAGUIDFromAttestationResponse(makeAttestationResponse("00000000-0000-0000-0000-000000000000"))
			So(err, ShouldBeNil)
			So(aaguid, ShouldEqual, "00000000-0000-0000-0000-000000000000")
		})

		Convey("returns an error for a malformed attestation response", func() {
			_, err := AAGUIDFromAttestationResponse([]byte(`{}`))
			So(err, ShouldNotBeNil)
		})
	})
}

func TestName(t *testing.T) {
	Convey("Name", t, func() {
		Convey("returns the name of a known AAGUID", func() {
			name, ok := Name("fbfc3007-154e-4ecc-8c0b-6e020557d7bd")
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, "Apple Passwords")

			name, ok = Name("ea9b8d66-4d01-1d21-3ce4-b6b48cb575d4")
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, "Google Password Manager")
		})

		Convey("ignores case", func() {
			name, ok := Name("BADA5566-A7AA-401F-BD96-45619A55120D")
			So(ok, ShouldBeTrue)
			So(name, ShouldEqual, "1Password")
		})

		Convey("does not know the all-zero AAGUID", func() {
			_, ok := Name("00000000-0000-0000-0000-000000000000")
			So(ok, ShouldBeFalse)
		})

		Convey("does not know an unlisted AAGUID", func() {
			_, ok := Name("11111111-2222-3333-4444-555555555555")
			So(ok, ShouldBeFalse)
		})
	})

	Convey("aaguid.json", t, func() {
		Convey("has a name for every lowercase UUID it lists", func() {
			var entries map[string]aaguidEntry
			err := json.Unmarshal(aaguidJSON, &entries)
			So(err, ShouldBeNil)
			So(entries, ShouldNotBeEmpty)
			for aaguid, entry := range entries {
				parsed, err := uuid.Parse(aaguid)
				So(err, ShouldBeNil)
				So(parsed.String(), ShouldEqual, aaguid)
				So(entry.Name, ShouldNotBeEmpty)
			}
		})
	})
}
