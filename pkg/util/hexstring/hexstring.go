package hexstring

import (
	"fmt"
	"math/big"
	"regexp"
	"strconv"
	"strings"
)

type T string

var parseRegExp = regexp.MustCompile(`^0x[0-9a-fA-F]+$`)

func (t T) ToBigInt() *big.Int {
	i := new(big.Int)
	i.SetString(string(t[2:]), 16)
	return i
}

func (t T) String() string {
	return string(t)
}

func NewFromInt64(v int64) (T, error) {
	if v < 0 {
		return "", fmt.Errorf("value must be positive")
	}

	return T(fmt.Sprintf("0x%s", strconv.FormatInt(v, 16))), nil
}

func NewFromBigInt(v *big.Int) (T, error) {
	if v.Cmp(big.NewInt(0)) < 0 {
		return "", fmt.Errorf("value must be positive")
	}

	return T(fmt.Sprintf("0x%s", v.Text(16))), nil
}

func Parse(s string) (T, error) {
	if parseRegExp.MatchString(s) {
		return T(strings.ToLower(s)), nil
	}
	return "", fmt.Errorf("hex string must match the regexp %q", parseRegExp)
}

func MustParse(s string) T {
	h, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return h
}

func FindSmallest(hexStrings []T) (T, int, bool) {
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
