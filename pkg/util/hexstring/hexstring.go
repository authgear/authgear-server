package hexstring

import (
	"fmt"
	"math/big"
	"strconv"
	"strings"
)

type HexString string

func (t HexString) ToBigInt() *big.Int {
	i := new(big.Int)
	i.SetString(string(t[2:]), 16)
	return i
}

func (t HexString) String() string {
	return string(t)
}

func NewFromInt64(v int64) (HexString, error) {
	if v < 0 {
		return "", fmt.Errorf("value must be positive")
	}

	return HexString(fmt.Sprintf("0x%s", strconv.FormatInt(v, 16))), nil
}

func NewFromBigInt(v *big.Int) (HexString, error) {
	if v.Cmp(big.NewInt(0)) < 0 {
		return "", fmt.Errorf("value must be positive")
	}

	return HexString(fmt.Sprintf("0x%s", v.Text(16))), nil
}

func Parse(s string) (HexString, error) {
	if !strings.HasPrefix(s, "0x") {
		return "", fmt.Errorf("hex string must start with 0x")
	}
	return HexString(s), nil
}

func MustParse(s string) HexString {
	h, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return h
}

func FindSmallest(hexStrings []HexString) (HexString, int, bool) {
	smHex, err := NewFromInt64(0)
	if err != nil {
		return "", -1, false
	}

	if len(hexStrings) == 0 {
		return "", -1, false
	}

	index := -1
	smDec := hexStrings[0].ToBigInt()
	for i, h := range hexStrings {
		dec := h.ToBigInt()

		if dec.Cmp(smDec) <= 0 {
			smHex = h
			smDec = dec
			index = i
		}
	}
	return smHex, index, true
}
