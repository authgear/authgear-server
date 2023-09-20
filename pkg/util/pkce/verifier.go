package pkce

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"

	"github.com/authgear/authgear-server/pkg/util/base32"
	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type CodeChallengeMethod = string

const CodeChallengeMethodS256 CodeChallengeMethod = "S256"

type Verifier struct {
	CodeChallengeMethod CodeChallengeMethod `json:"code_challenge_method"`
	CodeVerifier        string              `json:"code_verifier"`
}

func NewS256Verifier(codeVerifier string) *Verifier {
	return &Verifier{
		CodeChallengeMethod: CodeChallengeMethodS256,
		CodeVerifier:        codeVerifier,
	}
}

func GenerateS256Verifier() *Verifier {
	return NewS256Verifier(corerand.StringWithAlphabet(
		16,
		base32.Alphabet,
		corerand.SecureRand))
}

func (v *Verifier) Challenge() string {
	switch v.CodeChallengeMethod {
	case CodeChallengeMethodS256:
		verifierHash := sha256.Sum256([]byte(v.CodeVerifier))
		return base64.RawURLEncoding.EncodeToString(verifierHash[:])
	default:
		panic("unknown CodeChallengeMethod")
	}
}

func (v *Verifier) Verify(challenge string) bool {
	expectedChallenge := v.Challenge()
	return subtle.ConstantTimeCompare([]byte(challenge), []byte(expectedChallenge)) == 1
}
