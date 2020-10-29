package phone

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/nyaruka/phonenumbers"
)

// ErrNotInE164Format means the given phone number is not in E.164 format.
var ErrNotInE164Format = errors.New("not in E.164 format")

// ErrPhoneNumberInvalid means phone number doesn't pass validation
var ErrPhoneNumberInvalid = errors.New("invalid phone number")

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

func combineCallingCodeWithNumber(nationalNumber string, callingCode string) string {
	nationalNumber = strings.TrimSpace(nationalNumber)
	callingCode = strings.TrimSpace(callingCode)
	if !strings.HasPrefix(callingCode, "+") {
		callingCode = fmt.Sprintf("+%s", callingCode)
	}
	return fmt.Sprintf("%s%s", callingCode, nationalNumber)
}

// According to https://godoc.org/github.com/nyaruka/phonenumbers#Parse,
// no validation is performed during parsing, so we need to call IsValidNumber to check
func Parse(nationalNumber string, callingCode string) (e164 string, err error) {
	rawInput := combineCallingCodeWithNumber(nationalNumber, callingCode)
	// check if rawInput contains non-numeric character(s)
	// The nationalNumber part of phone is parsed to uint64,
	// letters in input phone number will be parsed successfully.
	isNumericString, _ := regexp.Match(`^\+[0-9\ \-]*$`, []byte(rawInput))
	if !isNumericString {
		err = ErrNotInE164Format
		return
	}
	num, err := phonenumbers.Parse(rawInput, "")
	if err != nil {
		return
	}
	isPhoneValid := phonenumbers.IsValidNumber(num)
	if !isPhoneValid {
		err = ErrPhoneNumberInvalid
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
