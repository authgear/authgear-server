package phone

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// ErrNotInE164Format means the given phone number is not in E.164 format.
var ErrNotInE164Format = errors.New("not in E.164 format")

// EnsureE164 ensures the given phone is in E.164 format.
func EnsureE164(phone string) error {
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

// Parse is a very lenient function to parse nationalNumber and callingCode into e164.
func Parse(nationalNumber string, callingCode string) (e164 string, err error) {
	nationalNumber = strings.TrimSpace(nationalNumber)
	callingCode = strings.TrimSpace(callingCode)
	if !strings.HasPrefix(callingCode, "+") {
		callingCode = fmt.Sprintf("+%s", callingCode)
	}
	var rawInput string
	rawInput = fmt.Sprintf("%s%s", callingCode, nationalNumber)
	num, err := phonenumbers.Parse(rawInput, "")
	if err != nil {
		return
	}
	e164 = phonenumbers.Format(num, phonenumbers.E164)
	return
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
