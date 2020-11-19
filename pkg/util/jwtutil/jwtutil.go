package jwtutil

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"
)

func Sign(t jwt.Token, alg jwa.SignatureAlgorithm, key interface{}) (token []byte, err error) {
	return SignWithHeader(t, jws.NewHeaders(), alg, key)
}

func SignWithHeader(t jwt.Token, hdr jws.Headers, alg jwa.SignatureAlgorithm, key interface{}) (token []byte, err error) {
	buf, err := json.Marshal(t)
	if err != nil {
		return
	}

	if _, ok := hdr.Get("alg"); !ok {
		err = hdr.Set("alg", alg.String())
		if err != nil {
			return
		}
	}

	if _, ok := hdr.Get("typ"); !ok {
		err = hdr.Set("typ", "JWT")
		if err != nil {
			return
		}
	}

	// Assume key is the actual key.
	realKey := key
	if _, ok := hdr.Get("kid"); !ok {
		if jwk, ok := key.(jwk.Key); ok {
			err = jwk.Raw(&realKey)
			if err != nil {
				return
			}

			kid := jwk.KeyID()
			if kid != "" {
				err = hdr.Set("kid", jwk.KeyID())
				if err != nil {
					return
				}
			}
		}
	}

	token, err = jws.Sign(buf, alg, realKey, jws.WithHeaders(hdr))
	return
}

// SplitWithoutVerify deserializes compact into hdr and payload.
func SplitWithoutVerify(compact []byte) (hdr jws.Headers, payload jwt.Token, err error) {
	msg, err := jws.Parse(bytes.NewReader(compact))
	if err != nil {
		return
	}

	sigs := msg.Signatures()
	if len(sigs) != 1 {
		err = fmt.Errorf("jwtutil: expected exact 1 signature but found: %v", len(sigs))
		return
	}
	sig := sigs[0]

	hdr = sig.ProtectedHeaders()

	payload = jwt.New()
	err = json.Unmarshal(msg.Payload(), payload)
	if err != nil {
		return
	}

	return
}
