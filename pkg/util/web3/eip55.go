package web3

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/authgear/authgear-server/pkg/util/hexstring"
	"golang.org/x/crypto/sha3"
)

// https://eips.ethereum.org/EIPS/eip-55
type EIP55 string

var parseRegExp = regexp.MustCompile(`^(0x)0*([0-9a-fA-F]+)$`)

func NewEIP55(s string) (EIP55, error) {
	if !parseRegExp.MatchString(s) {
		return "", fmt.Errorf("hex string must match the regexp %q", parseRegExp)
	}

	strippedHex := strings.ToLower(strings.ReplaceAll(s, "0x", ""))

	sha := sha3.NewLegacyKeccak256()
	sha.Write([]byte(strippedHex))
	hash := sha.Sum(nil)

	result := []byte(strippedHex)
	for i := 0; i < len(result); i++ {
		hashByte := hash[i/2]
		if i%2 == 0 {
			hashByte = hashByte >> 4
		} else {
			hashByte &= 0xf
		}
		if result[i] > '9' && hashByte > 7 {
			result[i] -= 32
		}
	}

	return EIP55("0x" + string(result)), nil
}

func (t EIP55) String() string {
	return string(t)
}

func NewEIP55FromHexstring(h hexstring.T) (EIP55, error) {
	return NewEIP55(h.String())
}

func (t EIP55) ToHexstring() (hexstring.T, error) {
	return hexstring.Parse(t.String())
}
