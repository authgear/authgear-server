package pkce

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"regexp"

	corerand "github.com/authgear/authgear-server/pkg/util/rand"
)

type CodeChallengeMethod = string

const CodeChallengeMethodS256 CodeChallengeMethod = "S256"

type Verifier struct {
	CodeChallengeMethod CodeChallengeMethod `json:"code_challenge_method"`
	CodeVerifier        string              `json:"code_verifier"`
}

var codeVerifierRegex = regexp.MustCompile(`^[-._~A-Za-z0-9]{43,128}$`)

var ErrInvalidCodeVerifier = errors.New("code_verifier must be a string between 43 and 128 characters long using A-Z, a-z, 0-9, -, ., _, ~")

func NewS256Verifier(codeVerifier string) (*Verifier, error) {
	if !codeVerifierRegex.MatchString(codeVerifier) {
		return nil, ErrInvalidCodeVerifier
	}

	return &Verifier{
		CodeChallengeMethod: CodeChallengeMethodS256,
		CodeVerifier:        codeVerifier,
	}, nil
}

func GenerateS256Verifier() *Verifier {
	// https://datatracker.ietf.org/doc/html/rfc7636#section-4.1
	// It is RECOMMENDED that the output of
	// a suitable random number generator be used to create a 32-octet
	// sequence.  The octet sequence is then base64url-encoded to produce a
	// 43-octet URL safe string to use as the code verifier.
	randBytes := make([]byte, 32)
	_, err := corerand.SecureRand.Read(randBytes)
	if err != nil {
		panic(err)
	}
	codeVerifier := base64.RawURLEncoding.EncodeToString(randBytes)
	verifier, err := NewS256Verifier(codeVerifier)
	if err != nil {
		panic(err)
	}
	return verifier
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
