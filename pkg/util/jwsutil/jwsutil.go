package jwsutil

import (
	"errors"
	"fmt"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/util/jwtutil"
)

var ErrNoKID = errors.New("no kid is found in the header")
var ErrKeyNotFound = errors.New("signing key not found in the key set")

// VerifyWithSet verify compact against keySet.
// The reason to have this function is because
// The keys in https://login.microsoftonline.com/common/discovery/v2.0/keys does not specify the alg.
// Therefore we need to look take both alg in the header and in the key into account.
func VerifyWithSet(keySet *jwk.Set, compact []byte) (hdr jws.Headers, payload jwt.Token, err error) {
	hdr, payload, err = jwtutil.SplitWithoutVerify(compact)
	if err != nil {
		return
	}

	keyID := hdr.KeyID()
	if keyID == "" {
		err = ErrNoKID
		return
	}

	keys := keySet.LookupKeyID(keyID)
	if len(keys) != 1 {
		err = ErrKeyNotFound
		return
	}

	key := keys[0]

	keyAlg := key.Algorithm()
	hdrAlg := hdr.Algorithm()

	// Prefer alg in key
	alg := keyAlg
	if alg == "" {
		alg = string(hdrAlg)
	}

	var rawKey interface{}
	if err = key.Raw(&rawKey); err != nil {
		err = fmt.Errorf("failed to materialize jwk.Key: %w", err)
		return
	}

	_, err = jws.Verify(compact, jwa.SignatureAlgorithm(alg), rawKey)
	if err != nil {
		return
	}

	return
}
