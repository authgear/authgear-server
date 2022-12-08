package jwtutil

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/lestrrat-go/jwx/jwa"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jws"
	"github.com/lestrrat-go/jwx/jwt"

	"github.com/authgear/authgear-server/pkg/api/apierrors"
)

var ErrInvalidJWTMutations = apierrors.Invalid.WithReason("InvalidJWTMutations")

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
	msg, err := jws.Parse(compact)
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

func BuildFromMap(m map[string]interface{}) (jwt.Token, error) {
	b := jwt.NewBuilder()
	for key, val := range m {
		b.Claim(key, val)
	}
	return b.Build()
}

func ToMap(t jwt.Token) (map[string]interface{}, error) {
	bytes, err := json.Marshal(t)
	if err != nil {
		return nil, err
	}

	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func PrepareForMutations(t jwt.Token) (
	forMutation map[string]interface{},
	forBackup map[string]interface{},
	err error,
) {
	cloned, err := t.Clone()
	if err != nil {
		return
	}

	forMutation, err = ToMap(t)
	if err != nil {
		return
	}

	forBackup, err = ToMap(cloned)
	if err != nil {
		return
	}

	return
}

func ApplyMutations(
	forMutation map[string]interface{},
	forBackup map[string]interface{},
) (applied jwt.Token, err error) {
	// We need to check 2 things here.
	// 1. No keys in forBackup were removed.
	// 2. All keys in forBackup were intact.

	removed := []string{}
	changed := []string{}

	for key := range forBackup {
		_, ok := forMutation[key]
		if !ok {
			removed = append(removed, key)
		} else {
			v1 := forBackup[key]
			v2 := forMutation[key]
			if !reflect.DeepEqual(v1, v2) {
				changed = append(changed, key)
			}
		}
	}

	if len(removed) > 0 || len(changed) > 0 {
		err = ErrInvalidJWTMutations.NewWithInfo("invalid JWT mutations", apierrors.Details{
			"removed": removed,
			"changed": changed,
		})
		return
	}

	applied, err = BuildFromMap(forMutation)
	if err != nil {
		return
	}

	return
}
