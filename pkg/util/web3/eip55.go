package web3

import (
	"fmt"
	"regexp"

	"github.com/ethereum/go-ethereum/common"

	"github.com/authgear/authgear-server/pkg/util/hexstring"
)

// https://eips.ethereum.org/EIPS/eip-55
type EIP55 string

var parseRegExp = regexp.MustCompile(`^(0x)0*([0-9a-fA-F]+)$`)

func NewEIP55(s string) (EIP55, error) {
	if !parseRegExp.MatchString(s) {
		return "", fmt.Errorf("hex string must match the regexp %q", parseRegExp)
	}

	// Special case for null address
	if s == "0x0" {
		return EIP55("0x0"), nil
	}

	address := common.HexToAddress(s)

	return EIP55(address.Hex()), nil
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

func (t EIP55) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func (t *EIP55) UnmarshalText(text []byte) error {
	parsed, err := NewEIP55(string(text))
	if err != nil {
		return err
	}
	*t = parsed
	return nil
}
