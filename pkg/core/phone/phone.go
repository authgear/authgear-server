package phone

import (
	"errors"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// ErrNotInE164Format means the given phone number is not in E.164 format.
var ErrNotInE164Format = errors.New("not in E.164 format")

// Parse ensures the given phone is in E.164 format.
func Parse(phone string) error {
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return err
	}
	formatted := phonenumbers.Format(num, phonenumbers.E164)
	if formatted != phone {
		return ErrNotInE164Format
	}
	return nil
}

// Mask masks the give phone number.
func Mask(phone string) string {
	var buf strings.Builder
	num, err := phonenumbers.Parse(phone, "")
	if err != nil {
		return ""
	}

	countryCallingCode := int(num.GetCountryCode())
	nationalSignificantNumber := phonenumbers.GetNationalSignificantNumber(num)
	runes := []rune(nationalSignificantNumber)
	length := len(runes)
	halfLength := length / 2

	buf.WriteRune('+')
	buf.WriteString(strconv.Itoa(countryCallingCode))
	for i := 0; i < length; i++ {
		if i < halfLength {
			buf.WriteRune(runes[i])
		} else {
			buf.WriteRune('*')
		}
	}

	return buf.String()
}
